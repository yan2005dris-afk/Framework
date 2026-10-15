// Package controllers implementa la lógica de negocio para el endpoint de checkout.
package controllers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// CheckoutController controla el flujo transaccional de checkout con dependencias inyectadas.
type CheckoutController struct {
	dbRepo    repository.DatabaseRepository
	cacheRepo repository.CacheRepository
}

// NewCheckoutController crea una nueva instancia de CheckoutController.
func NewCheckoutController(dbRepo repository.DatabaseRepository, cacheRepo repository.CacheRepository) *CheckoutController {
	return &CheckoutController{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
	}
}

// Checkout procesa la compra con transacciones ACID en PostgreSQL y cálculo asíncrono vía Goroutine.
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

	// Por defecto asociamos a usuario_id = 2 (cliente semilla)
	usuarioID := 2

	// Paso 3: Transacción ACID en PostgreSQL (tx.Begin, INSERTs, Rollback/Commit)
	pedido, err := c.dbRepo.CrearPedido(usuarioID, req)
	if err != nil {
		log.Printf("❌ Error en transacción de checkout: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Error al procesar el pedido. Transacción revertida.",
			"detalle": err.Error(),
		})
	}

	// Paso 4: Concurrencia con Goroutine en segundo plano (Misión Crítica)
	// No bloquea la respuesta al cliente
	go func(total float64, uID int, email string) {
		log.Printf("🧵 [Goroutine] Iniciando cálculo asíncrono de comisiones para usuario %d, total $%.2f", uID, total)
		if err := c.dbRepo.ActualizarComisiones(uID, total); err != nil {
			log.Printf("⚠️ [Goroutine] Error al actualizar comisiones: %v", err)
		} else {
			log.Printf("✅ [Goroutine] Comisiones actualizadas con éxito en PostgreSQL")
		}

		// Paso 5: Invalidación de Caché en Redis
		if c.cacheRepo != nil {
			// Clave global del árbol
			_ = c.cacheRepo.Del("red:tree")
			// Clave específica por usuario si aplica
			if email != "" {
				_ = c.cacheRepo.Del(fmt.Sprintf("red:tree:%s", email))
			}
			log.Printf("🧹 [Goroutine] Caché de red invalidada en Redis")
		}
	}(pedido.Total, usuarioID, req.EmailCliente)

	numeroPedido := fmt.Sprintf("MC-%06d", pedido.ID)

	return ctx.Status(fiber.StatusCreated).JSON(models.CheckoutResponse{
		PedidoID: pedido.ID,
		Numero:   numeroPedido,
		Total:    pedido.Total,
		Mensaje:  "Pedido procesado exitosamente",
	})
}
