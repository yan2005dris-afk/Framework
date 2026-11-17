// Package controllers implementa los controladores de la aplicación.
package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"multicatalogo-backend/config"
)

// HealthController verifica el estado operativo del servidor y sus dependencias críticas
// (PostgreSQL + Redis). Implementa el contrato:
//   - 200 + status="ok"       -> todo sano.
//   - 200 + status="degraded" -> PG OK, Redis caído (servidor responde pero sin caché).
//   - 503 + status="unhealthy"-> PG caído (no se puede servir la API correctamente).
type HealthController struct {
	dbPool *pgxpool.Pool
	redis  *redis.Client
}

// NewHealthController crea un HealthController con las dependencias de pool y cliente Redis.
// Si dbPool o redis vienen nil, se intenta usar config.DB / config.RedisClient como fallback.
func NewHealthController(dbPool *pgxpool.Pool, redisClient *redis.Client) *HealthController {
	if dbPool == nil {
		dbPool = config.DB
	}
	if redisClient == nil {
		redisClient = config.RedisClient
	}
	return &HealthController{dbPool: dbPool, redis: redisClient}
}

// Health ejecuta un ping a PostgreSQL y Redis con timeout corto (no usa el contexto del request).
// Devuelve JSON con el estado desglosado por componente.
func (h *HealthController) Health(c *fiber.Ctx) error {
	pgStatus := "down"
	redisStatus := "down"

	if h.dbPool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := h.dbPool.Ping(ctx)
		cancel()
		if err == nil {
			pgStatus = "up"
		}
	}

	if h.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := h.redis.Ping(ctx).Err()
		cancel()
		if err == nil {
			redisStatus = "up"
		}
	}

	components := fiber.Map{
		"postgres": pgStatus,
		"redis":    redisStatus,
	}

	switch {
	case pgStatus == "down":
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":     "unhealthy",
			"components": components,
		})
	case redisStatus == "down":
		return c.JSON(fiber.Map{
			"status":     "degraded",
			"components": components,
			"message":    "Funcionando sin caché: Redis no responde",
		})
	default:
		return c.JSON(fiber.Map{
			"status":     "ok",
			"components": components,
		})
	}
}

// GetHealth conserva compatibilidad con código que aún use la firma package-level.
// Devuelve siempre un status estático (útil solo cuando no hay DI disponible).
func GetHealth(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Servidor Go/Fiber operativo (sin chequeo de dependencias)",
	})
}
