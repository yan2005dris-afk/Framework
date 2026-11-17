// Package controllers implementa la lógica de negocio para el endpoint de checkout.
package controllers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
	"multicatalogo-backend/events"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// CheckoutController controla el flujo transaccional de checkout con dependencias inyectadas.
type CheckoutController struct {
	dbRepo    repository.DatabaseRepository
	cacheRepo repository.CacheRepository
	publisher events.Publisher
	appEnv    string
}

// NewCheckoutController crea una nueva instancia de CheckoutController.
// appEnv controla la sanitización de errors: "dev" los expone, otro valor los oculta.
func NewCheckoutController(dbRepo repository.DatabaseRepository, cacheRepo repository.CacheRepository, appEnv string) *CheckoutController {
	if appEnv == "" {
		appEnv = "dev"
	}
	var pub events.Publisher
	eventPub, err := events.NewEventPublisher()
	if err != nil {
		log.Printf("[checkout] Valkey EventPublisher no inicializado: %v", err)
	} else {
		pub = eventPub
	}
	return &CheckoutController{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		publisher: pub,
		appEnv:    appEnv,
	}
}

// SetPublisher permite inyectar un Publisher alternativo o mock para testing.
func (c *CheckoutController) SetPublisher(pub events.Publisher) {
	c.publisher = pub
}

// Checkout procesa la compra con transacciones ACID en PostgreSQL y cálculo asíncrono vía Goroutine.
//
// Identidad: el usuarioID y el email se leen de c.Locals (puestos por
// middleware.RequireAuth). El campo emailCliente del body NO se persiste
// ni atribuye comisiones; es información de envío/tax solamente y se
// valida que sea no vacío.
func (c *CheckoutController) Checkout(ctx *fiber.Ctx) error {
	var req models.CheckoutRequest
	if err := ctx.BodyParser(&req); err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "Cuerpo de petición inválido"}.Write(ctx)
	}

	if len(req.Items) == 0 {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El carrito no contiene items"}.Write(ctx)
	}

	// Validar cantidad por item ANTES de la transacción.
	for _, item := range req.Items {
		if item.Cantidad <= 0 {
			public, detail := helpers.SanitizeErrorDetails(c.appEnv, repository.ErrCantidadInvalida)
			return models.APIError{Status: fiber.StatusBadRequest, Error: public, Detalle: detail}.Write(ctx)
		}
	}

	// Identidad tomada del token (RequireAuth debe haber corrido antes).
	usuarioIDVal, ok := ctx.Locals(middleware.LocalUsuarioID).(int)
	if !ok {
		// Fallback defensivo: si el middleware no corrió (test directo), devolvemos 401.
		return models.APIError{Status: fiber.StatusUnauthorized, Error: "Autenticación requerida"}.Write(ctx)
	}
	tokenEmail, _ := ctx.Locals(middleware.LocalEmail).(string)

	// Seguridad: el email persistido en pedidos es SIEMPRE el del token.
	// El body puede mentir y la fila en BD no debe atribuirse al impostor.
	if tokenEmail != "" {
		req.EmailCliente = tokenEmail
	}

	// Idempotency-Key (U4-T3, S10): la clave que el cliente manda en el reintento
	// de un checkout. Opcional a proposito: un cliente viejo, o uno que no este
	// en la PWA, sigue funcionando sin ella, solo que sin proteccion contra el
	// reintento que duplica pedidos. El largo se limita porque va a una columna
	// VARCHAR(255) y es un identificador que genera el cliente, no texto libre.
	idempotencyKey := strings.TrimSpace(ctx.Get("Idempotency-Key"))
	if len(idempotencyKey) > 255 {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El header Idempotency-Key excede los 255 caracteres"}.Write(ctx)
	}

	// Paso 3: Transacción ACID en PostgreSQL (tx.Begin, INSERTs, Rollback/Commit)
	pedido, replay, err := c.dbRepo.CrearPedido(usuarioIDVal, req, idempotencyKey)
	if err != nil {
		log.Printf("[checkout] Error en transacción de checkout: %v", err)
		public, detail := helpers.SanitizeErrorDetails(c.appEnv, err)
		if public == "" {
			public = "Error al procesar el pedido. Transacción revertida."
		}
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(ctx)
	}

	// Un replay no vuelve a cobrar comisiones: el pedido ya las aplicó el primer
	// intento. Volver a sumarlas seria el mismo bug de duplicado, un paso mas
	// abajo y mas dificil de detectar.
	if replay {
		return ctx.Status(fiber.StatusOK).JSON(models.CheckoutResponse{
			PedidoID: pedido.ID,
			Numero:   fmt.Sprintf("MC-%06d", pedido.ID),
			Total:    pedido.Total,
			Mensaje:  "Este pedido ya se había confirmado. Te mostramos el original.",
		})
	}

	// Paso 4: Concurrencia con Goroutine en segundo plano (Misión Crítica)
	// No bloquea la respuesta al cliente.
	// Importante: el email que se usa para invalidar la caché es el del
	// token, NO el del body. Así la atribución de comisiones siempre va
	// al usuario autenticado, incluso si el body miente.
	go func(ped *models.Pedido, reqItems []models.CheckoutItem, uID int, email string) {
		log.Printf("[goroutine] Iniciando cálculo asíncrono de comisiones para usuario %d, total $%.2f", uID, ped.Total)
		if err := c.dbRepo.ActualizarComisiones(uID, ped.Total); err != nil {
			log.Printf("[goroutine] Error al actualizar comisiones: %v", err)
		} else {
			log.Printf("[goroutine] Comisiones actualizadas con éxito en PostgreSQL")
		}
		if _, err := c.dbRepo.RegistrarComisionesPedido(ped.ID, uID, ped.Total); err != nil {
			log.Printf("[goroutine] Error al registrar ledger de comisiones: %v", err)
		} else {
			log.Printf("[goroutine] Ledger de comisiones registrado con éxito")
		}

		// Invalidación de Caché en Valkey.
		// productos:all debe invalidarse porque el stock cambió.
		// red:tree y red:tree:{email} deben invalidarse porque las ventas cambiaron.
		if c.cacheRepo != nil {
			_ = c.cacheRepo.Del("productos:all")
			_ = c.cacheRepo.Del("red:tree")
			if email != "" {
				_ = c.cacheRepo.Del(fmt.Sprintf("red:tree:%s", email))
			}
			log.Printf("[goroutine] Caché de productos y red invalidada en Valkey")
		}

		// Publicación de evento de dominio en Valkey Streams
		if c.publisher != nil {
			totalCantidad := 0
			for _, item := range reqItems {
				totalCantidad += item.Cantidad
			}

			event := events.PedidoCreadoEvent{
				PedidoID:  ped.ID,
				TenantID:  uID,
				Cliente:   ped.NombreCliente,
				Cantidad:  totalCantidad,
				Total:     ped.Total,
				CreatedAt: time.Now().UTC(),
			}

			if _, err := c.publisher.PublishPedidoCreado(context.Background(), event); err != nil {
				log.Printf("[events] Error publicando pedido.creado: %v", err)
			}
		}
	}(pedido, req.Items, usuarioIDVal, tokenEmail)

	numeroPedido := fmt.Sprintf("MC-%06d", pedido.ID)

	return ctx.Status(fiber.StatusCreated).JSON(models.CheckoutResponse{
		PedidoID: pedido.ID,
		Numero:   numeroPedido,
		Total:    pedido.Total,
		Mensaje:  "Pedido procesado exitosamente",
	})
}