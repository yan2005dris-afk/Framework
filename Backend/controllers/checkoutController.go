// Package controllers implementa la lógica de negocio para el endpoint de checkout.
package controllers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// CheckoutController controla el flujo transaccional de checkout con dependencias inyectadas.
type CheckoutController struct {
	dbRepo    repository.DatabaseRepository
	cacheRepo repository.CacheRepository
	appEnv    string
}

// NewCheckoutController crea una nueva instancia de CheckoutController.
// appEnv controla la sanitización de errors: "dev" los expone, otro valor los oculta.
func NewCheckoutController(dbRepo repository.DatabaseRepository, cacheRepo repository.CacheRepository, appEnv string) *CheckoutController {
	if appEnv == "" {
		appEnv = "dev"
	}
	return &CheckoutController{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		appEnv:    appEnv,
	}
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
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cuerpo de petición inválido",
		})
	}

	if len(req.Items) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El carrito no contiene items",
		})
	}

	// Validar cantidad por item ANTES de la transacción.
	for _, item := range req.Items {
		if item.Cantidad <= 0 {
			public, detail := helpers.SanitizeErrorDetails(c.appEnv, repository.ErrCantidadInvalida)
			resp := fiber.Map{"error": public}
			if detail != "" {
				resp["detalle"] = detail
			}
			return ctx.Status(fiber.StatusBadRequest).JSON(resp)
		}
	}

	// Identidad tomada del token (RequireAuth debe haber corrido antes).
	usuarioIDVal, ok := ctx.Locals(middleware.LocalUsuarioID).(int)
	if !ok {
		// Fallback defensivo: si el middleware no corrió (test directo), devolvemos 401.
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Autenticación requerida",
		})
	}
	tokenEmail, _ := ctx.Locals(middleware.LocalEmail).(string)

	// Seguridad: el email persistido en pedidos es SIEMPRE el del token.
	// El body puede mentir y la fila en BD no debe atribuirse al impostor.
	if tokenEmail != "" {
		req.EmailCliente = tokenEmail
	}

	// Paso 3: Transacción ACID en PostgreSQL (tx.Begin, INSERTs, Rollback/Commit)
	pedido, err := c.dbRepo.CrearPedido(usuarioIDVal, req)
	if err != nil {
		log.Printf("❌ Error en transacción de checkout: %v", err)
		public, detail := helpers.SanitizeErrorDetails(c.appEnv, err)
		resp := fiber.Map{
			"error": "Error al procesar el pedido. Transacción revertida.",
		}
		if public != "" {
			resp["error"] = public
		}
		if detail != "" {
			resp["detalle"] = detail
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(resp)
	}

	// Paso 4: Concurrencia con Goroutine en segundo plano (Misión Crítica)
	// No bloquea la respuesta al cliente.
	// Importante: el email que se usa para invalidar la caché es el del
	// token, NO el del body. Así la atribución de comisiones siempre va
	// al usuario autenticado, incluso si el body miente.
	go func(total float64, uID int, email string) {
		log.Printf("🧵 [Goroutine] Iniciando cálculo asíncrono de comisiones para usuario %d, total $%.2f", uID, total)
		if err := c.dbRepo.ActualizarComisiones(uID, total); err != nil {
			log.Printf("⚠️ [Goroutine] Error al actualizar comisiones: %v", err)
		} else {
			log.Printf("✅ [Goroutine] Comisiones actualizadas con éxito en PostgreSQL")
		}

		// Paso 5: Invalidación de Caché en Redis.
		// productos:all debe invalidarse porque el stock cambió.
		// red:tree y red:tree:{email} deben invalidarse porque las ventas cambiaron.
		if c.cacheRepo != nil {
			_ = c.cacheRepo.Del("productos:all")
			_ = c.cacheRepo.Del("red:tree")
			if email != "" {
				_ = c.cacheRepo.Del(fmt.Sprintf("red:tree:%s", email))
			}
			log.Printf("🧹 [Goroutine] Caché de productos y red invalidada en Redis")
		}
	}(pedido.Total, usuarioIDVal, tokenEmail)

	numeroPedido := fmt.Sprintf("MC-%06d", pedido.ID)

	return ctx.Status(fiber.StatusCreated).JSON(models.CheckoutResponse{
		PedidoID: pedido.ID,
		Numero:   numeroPedido,
		Total:    pedido.Total,
		Mensaje:  "Pedido procesado exitosamente",
	})
}