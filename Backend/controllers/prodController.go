// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// ProductoController maneja las peticiones sobre el catálogo de productos con DI.
type ProductoController struct {
	dbRepo    repository.DatabaseRepository
	cacheRepo repository.CacheRepository
	appEnv    string
}

// NewProductoController crea una instancia de ProductoController.
// appEnv controla la sanitización de errors: "dev" los expone, otro valor los oculta.
func NewProductoController(dbRepo repository.DatabaseRepository, cacheRepo repository.CacheRepository, appEnv string) *ProductoController {
	if appEnv == "" {
		appEnv = "dev"
	}
	return &ProductoController{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		appEnv:    appEnv,
	}
}

// GetProductos retorna la lista completa de productos desde la base de datos con Cache-Aside.
func (pc *ProductoController) GetProductos(c *fiber.Ctx) error {
	cacheKey := "productos:all"

	// 1. Intentar leer desde Redis (Cache Hit)
	if pc.cacheRepo != nil {
		cachedData, err := pc.cacheRepo.Get(cacheKey)
		if err == nil {
			var productos []models.Producto
			if err := json.Unmarshal(cachedData, &productos); err == nil {
				return c.JSON(productos)
			}
		}
	}

	// 2. Cache Miss: consultar PostgreSQL
	productos, err := pc.dbRepo.ObtenerProductos()
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		resp := fiber.Map{"error": public}
		if detail != "" {
			resp["detalle"] = detail
		}
		return c.Status(fiber.StatusInternalServerError).JSON(resp)
	}

	// 3. Guardar en Redis con TTL de 5 minutos (300 segundos)
	if pc.cacheRepo != nil {
		data, err := json.Marshal(productos)
		if err == nil {
			_ = pc.cacheRepo.Set(cacheKey, data, 300)
		}
	}

	return c.JSON(productos)
}

// GetProductoByID retorna un único producto según el parámetro de ruta ID.
func (pc *ProductoController) GetProductoByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El ID del producto debe ser numérico"}.Write(c)
	}

	producto, err := pc.dbRepo.ObtenerProductoPorID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Producto no encontrado",
			})
		}
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		resp := fiber.Map{"error": public}
		if detail != "" {
			resp["detalle"] = detail
		}
		return c.Status(fiber.StatusInternalServerError).JSON(resp)
	}

	return c.JSON(producto)
}

// Funciones globales para compatibilidad hacia atrás
func GetProductos(c *fiber.Ctx) error {
	ctrl := NewProductoController(&repository.PostgresRepository{}, &repository.RedisCacheRepository{}, "dev")
	return ctrl.GetProductos(c)
}

func GetProductoByID(c *fiber.Ctx) error {
	ctrl := NewProductoController(&repository.PostgresRepository{}, &repository.RedisCacheRepository{}, "dev")
	return ctrl.GetProductoByID(c)
}

// GetProductosAdmin retorna todos los productos incluidos inactivos con soporte de búsqueda ?q= (Solo Admin).
func (pc *ProductoController) GetProductosAdmin(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	q := c.Query("q", "")
	productos, err := pc.dbRepo.ObtenerProductosAdmin(q)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	return c.JSON(productos)
}

// CrearProducto da de alta un nuevo producto y limpia la caché (Solo Admin).
func (pc *ProductoController) CrearProducto(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	var req models.ProductoRequest
	if err := c.BodyParser(&req); err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "Cuerpo de petición inválido"}.Write(c)
	}

	if req.Nombre == "" || req.Precio <= 0 || req.Categoria == "" {
		return models.APIError{
			Status:  fiber.StatusBadRequest,
			Error:   "Validación fallida",
			Detalle: "Los campos nombre, categoria y precio (> 0) son obligatorios",
		}.Write(c)
	}

	producto, err := pc.dbRepo.CrearProducto(req)
	if err != nil {
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	// Invalidar caché de catálogo
	if pc.cacheRepo != nil {
		_ = pc.cacheRepo.Del("productos:all")
	}

	return c.Status(fiber.StatusCreated).JSON(producto)
}

// ActualizarProducto edita campos de un producto existente y limpia la caché (Solo Admin).
func (pc *ProductoController) ActualizarProducto(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El ID del producto debe ser numérico"}.Write(c)
	}

	var req models.ProductoRequest
	if err := c.BodyParser(&req); err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "Cuerpo de petición inválido"}.Write(c)
	}

	producto, err := pc.dbRepo.ActualizarProducto(id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.APIError{Status: fiber.StatusNotFound, Error: "Producto no encontrado"}.Write(c)
		}
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	// Invalidar caché de catálogo
	if pc.cacheRepo != nil {
		_ = pc.cacheRepo.Del("productos:all")
	}

	return c.JSON(producto)
}

// EliminarProducto realiza borrado lógico (activo=false) y limpia la caché (Solo Admin).
func (pc *ProductoController) EliminarProducto(c *fiber.Ctx) error {
	rol, ok := c.Locals(middleware.LocalRol).(string)
	if !ok || rol != "admin" {
		return models.APIError{Status: fiber.StatusForbidden, Error: "Se requieren permisos de administrador"}.Write(c)
	}

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "El ID del producto debe ser numérico"}.Write(c)
	}

	if err := pc.dbRepo.EliminarProductoLogico(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.APIError{Status: fiber.StatusNotFound, Error: "Producto no encontrado"}.Write(c)
		}
		public, detail := helpers.SanitizeErrorDetails(pc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
	}

	// Invalidar caché de catálogo
	if pc.cacheRepo != nil {
		_ = pc.cacheRepo.Del("productos:all")
	}

	return c.JSON(fiber.Map{
		"mensaje": "Producto desactivado exitosamente",
	})
}
