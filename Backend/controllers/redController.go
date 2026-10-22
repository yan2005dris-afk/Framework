// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers/helpers"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// RedController maneja las peticiones de la red multinivel con DI.
type RedController struct {
	dbRepo    repository.DatabaseRepository
	cacheRepo repository.CacheRepository
	appEnv    string
}

// NewRedController crea una instancia de RedController.
// appEnv controla la sanitización de errors: "dev" los expone, otro valor los oculta.
func NewRedController(dbRepo repository.DatabaseRepository, cacheRepo repository.CacheRepository, appEnv string) *RedController {
	if appEnv == "" {
		appEnv = "dev"
	}
	return &RedController{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		appEnv:    appEnv,
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
			return models.APIError{Status: fiber.StatusNotFound, Error: "Estructura de red no encontrada"}.Write(c)
		}
		public, detail := helpers.SanitizeErrorDetails(rc.appEnv, err)
		return models.APIError{Status: fiber.StatusInternalServerError, Error: public, Detalle: detail}.Write(c)
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
	ctrl := NewRedController(&repository.PostgresRepository{}, &repository.RedisCacheRepository{}, "dev")
	return ctrl.GetRed(c)
}
