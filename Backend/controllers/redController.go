// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// RedController maneja las peticiones de la red multinivel con DI.
type RedController struct {
	dbRepo    repository.DatabaseRepository
	cacheRepo repository.CacheRepository
}

// NewRedController crea una instancia de RedController.
func NewRedController(dbRepo repository.DatabaseRepository, cacheRepo repository.CacheRepository) *RedController {
	return &RedController{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
	}
}

// GetRed retorna la estructura jerárquica de la red de referidos con Cache-Aside.
func (rc *RedController) GetRed(c *fiber.Ctx) error {
	cacheKey := "red:tree"

	// 1. Intentar leer desde Redis (Cache Hit)
	if rc.cacheRepo != nil {
		cachedData, err := rc.cacheRepo.Get(cacheKey)
		if err == nil {
			var red models.Referido
			if err := json.Unmarshal(cachedData, &red); err == nil {
				return c.JSON(red)
			}
		}
	}

	// 2. Cache Miss: consultar PostgreSQL
	red, err := rc.dbRepo.ObtenerRed()
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Estructura de red no encontrada",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al obtener la red de referidos",
		})
	}

	// 3. Guardar en Redis con TTL de 10 minutos (600 segundos)
	if rc.cacheRepo != nil {
		data, err := json.Marshal(red)
		if err == nil {
			_ = rc.cacheRepo.Set(cacheKey, data, 600)
		}
	}

	return c.JSON(red)
}

// Compatibilidad hacia atrás
func GetRed(c *fiber.Ctx) error {
	ctrl := NewRedController(&repository.PostgresRepository{}, &repository.RedisCacheRepository{})
	return ctrl.GetRed(c)
}
