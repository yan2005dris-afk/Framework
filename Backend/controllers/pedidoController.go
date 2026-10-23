package controllers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// PedidoController expone las operaciones de consulta y administración del ciclo de pedidos.
type PedidoController struct {
	dbRepo repository.DatabaseRepository
	appEnv string
}

// NewPedidoController crea una nueva instancia de PedidoController.
func NewPedidoController(dbRepo repository.DatabaseRepository, appEnv string) *PedidoController {
	if appEnv == "" {
		appEnv = "dev"
	}
	return &PedidoController{
		dbRepo: dbRepo,
		appEnv: appEnv,
	}
}

// ListarPedidosMios retorna el historial de compras del usuario autenticado.
func (pc *PedidoController) ListarPedidosMios(c *fiber.Ctx) error {
	usuarioID, ok := c.Locals(middleware.LocalUsuarioID).(int)
	if !ok || usuarioID <= 0 {
		return models.APIError{Status: fiber.StatusUnauthorized, Error: "Usuario no autenticado"}.Write(c)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	pedidos, err := pc.dbRepo.ListarPedidosMios(usuarioID, limit, offset)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(fiber.Map{
		"data":  pedidos,
		"page":  page,
		"limit": limit,
	})
}

// ListarPedidosAdmin retorna todos los pedidos con filtro opcional de estado (Solo Admin).
func (pc *PedidoController) ListarPedidosAdmin(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	estado := c.Query("estado", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	pedidos, err := pc.dbRepo.ListarPedidosAdmin(estado, limit, offset)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(fiber.Map{
		"data":   pedidos,
		"page":   page,
		"limit":  limit,
		"estado": estado,
	})
}

// ObtenerPedidoPorID devuelve la cabecera e ítems detallados si el usuario es dueño o admin.
func (pc *PedidoController) ObtenerPedidoPorID(c *fiber.Ctx) error {
	usuarioID, _ := c.Locals(middleware.LocalUsuarioID).(int)
	rol, _ := c.Locals(middleware.LocalRol).(string)

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El ID del pedido debe ser numérico"}.Write(c)
	}

	pedido, err := pc.dbRepo.ObtenerPedidoDetalle(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.APIError{Status: fiber.StatusNotFound, Error: "Pedido no encontrado"}.Write(c)
		}
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	// Regla de autorización: solo el dueño del pedido o un admin pueden ver el detalle
	if rol != "admin" && pedido.UsuarioID != usuarioID {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Acceso no autorizado al pedido"}.Write(c)
	}

	return c.JSON(pedido)
}

// ActualizarEstado modifica el estado de un pedido (Solo Admin).
func (pc *PedidoController) ActualizarEstado(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El ID del pedido debe ser numérico"}.Write(c)
	}

	var req models.ActualizarEstadoPedidoRequest
	if err := c.BodyParser(&req); err != nil || req.Estado == "" {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El campo estado es requerido"}.Write(c)
	}

	// Validar transiciones permitidas
	estadosPermitidos := map[string]bool{
		"pendiente":  true,
		"preparando": true,
		"enviado":    true,
		"entregado":  true,
		"cancelado":  true,
	}
	if !estadosPermitidos[req.Estado] {
		return models.APIError{
			Status:  fiber.StatusBadRequest,
			Error:   "Estado inválido",
			Detalle: "Los estados válidos son: pendiente, preparando, enviado, entregado, cancelado",
		}.Write(c)
	}

	if err := pc.dbRepo.ActualizarEstadoPedido(id, req.Estado); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.APIError{Status: fiber.StatusNotFound, Error: "Pedido no encontrado"}.Write(c)
		}
		return models.APIError{Status: fiber.StatusBadRequest, Error: err.Error()}.Write(c)
	}

	return c.JSON(fiber.Map{
		"mensaje": "Estado de pedido actualizado",
		"estado":  req.Estado,
	})
}
