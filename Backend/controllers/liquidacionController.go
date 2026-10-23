package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// LiquidacionController expone los endpoints para ledger de comisiones y pagos a distribuidores.
type LiquidacionController struct {
	repo   repository.DatabaseRepository
	appEnv string
}

// NewLiquidacionController instancia el controlador con inyección de dependencias.
func NewLiquidacionController(repo repository.DatabaseRepository, appEnv string) *LiquidacionController {
	return &LiquidacionController{
		repo:   repo,
		appEnv: appEnv,
	}
}

// ObtenerComisionesMias devuelve las comisiones devengadas y el saldo pendiente del usuario autenticado.
func (lc *LiquidacionController) ObtenerComisionesMias(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals(middleware.LocalUsuarioID).(int)
	if !ok || usuarioID < 0 {
		return models.APIError{Status: fiber.StatusUnauthorized, Error: "Usuario no autenticado o sesión inválida"}.Write(c)
	}

	resumen, err := lc.repo.ObtenerComisionesPorReferido(usuarioID)
	if err != nil {
		log.Printf("⚠️ Error al obtener comisiones para usuario %d: %v", usuarioID, err)
		public, detail := helpers.SanitizeErrorDetails(lc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(resumen)
}

// Liquidar ejecuta la liquidación de comisiones de un referido (Solo Admin).
func (lc *LiquidacionController) Liquidar(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	var req models.LiquidarRequest
	if err := c.BodyParser(&req); err != nil || req.ReferidoID < 0 {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "ID de referido inválido", Detalle: "El campo referidoId es requerido"}.Write(c)
	}

	liq, err := lc.repo.LiquidarComisiones(req.ReferidoID)
	if err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: err.Error()}.Write(c)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"mensaje":     "Liquidación procesada exitosamente",
		"liquidacion": liq,
	})
}

// ListarLiquidaciones lista todas las liquidaciones realizadas en el sistema (Solo Admin).
func (lc *LiquidacionController) ListarLiquidaciones(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	liquidaciones, err := lc.repo.ListarLiquidaciones()
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(lc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(fiber.Map{
		"liquidaciones": liquidaciones,
	})
}

// ListarLiquidacionesMias devuelve el historial de pagos recibidos por el distribuidor autenticado.
func (lc *LiquidacionController) ListarLiquidacionesMias(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals(middleware.LocalUsuarioID).(int)
	if !ok || usuarioID < 0 {
		return models.APIError{Status: fiber.StatusUnauthorized, Error: "Usuario no autenticado o sesión inválida"}.Write(c)
	}

	liquidaciones, err := lc.repo.ListarLiquidacionesPorReferido(usuarioID)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(lc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(fiber.Map{
		"liquidaciones": liquidaciones,
	})
}
