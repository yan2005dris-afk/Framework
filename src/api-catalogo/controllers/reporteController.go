package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// ReporteController provee endpoints de métricas y KPIs de negocio (Solo Admin).
type ReporteController struct {
	dbRepo repository.DatabaseRepository
	appEnv string
}

// NewReporteController inicializa el controlador con dependencias.
func NewReporteController(dbRepo repository.DatabaseRepository, appEnv string) *ReporteController {
	if appEnv == "" {
		appEnv = "dev"
	}
	return &ReporteController{
		dbRepo: dbRepo,
		appEnv: appEnv,
	}
}

// Resumen retorna total vendido, número de pedidos, ticket medio y comisiones.
func (rc *ReporteController) Resumen(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	desde := c.Query("desde", "")
	hasta := c.Query("hasta", "")

	resumen, err := rc.dbRepo.ObtenerReporteResumen(desde, hasta)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(rc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(resumen)
}

// TopProductos retorna los productos más vendidos en ventas acumuladas.
func (rc *ReporteController) TopProductos(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "5"))
	top, err := rc.dbRepo.ObtenerTopProductos(limit)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(rc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(fiber.Map{
		"data": top,
	})
}

// TopReferidos retorna los distribuidores con más ventas consolidadas.
func (rc *ReporteController) TopReferidos(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	top, err := rc.dbRepo.ObtenerTopReferidos(limit)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(rc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(fiber.Map{
		"data": top,
	})
}

// ComisionesPendientes agrupa la deuda por pagar a distribuidores.
func (rc *ReporteController) ComisionesPendientes(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	deudas, err := rc.dbRepo.ObtenerComisionesPendientes()
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(rc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(fiber.Map{
		"data": deudas,
	})
}
