package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetRed retorna la estructura jerárquica de la red de referidos desde la base de datos con Cache-Aside.
func GetRed(c *fiber.Ctx) error {
	ctx := context.Background()
	cacheKey := "red:tree"

	// 1. Intentar leer desde Redis (Cache Hit)
	cachedData, err := config.RedisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var red models.Referido
		if err := json.Unmarshal(cachedData, &red); err == nil {
			return c.JSON(red)
		}
	}

	// 2. Cache Miss: consultar PostgreSQL
	red, err := repository.ObtenerRed()
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

	// 3. Guardar en Redis con TTL de 10 minutos
	data, err := json.Marshal(red)
	if err == nil {
		config.RedisClient.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return c.JSON(red)
}
