// Package handlers contiene los handlers HTTP del auth-service.
// En GA1 solo hay health; GA2+ agrega /auth/login, /auth/validate, etc.
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"auth-service/internal/db"
)

// HealthHandler expone /auth/health con ping a la BD.
// Devuelve 200 {"status":"ok"} si la BD responde; 503 si no.
type HealthHandler struct {
	DB *db.DB
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dbStatus := "up"
	if err := h.DB.Ping(ctx); err != nil {
		dbStatus = "down"
	}

	status := "ok"
	httpStatus := fiber.StatusOK
	if dbStatus == "down" {
		status = "degraded"
		httpStatus = fiber.StatusServiceUnavailable
	}

	return c.Status(httpStatus).JSON(fiber.Map{
		"status":     status,
		"components": fiber.Map{"db": dbStatus},
	})
}
