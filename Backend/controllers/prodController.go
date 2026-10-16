// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El ID del producto debe ser numérico",
		})
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
