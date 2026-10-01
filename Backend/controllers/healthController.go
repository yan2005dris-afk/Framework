package controllers

import "github.com/gofiber/fiber/v2"

// GetHealth verifica el estado operativo del servidor.
func GetHealth(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Servidor Go/Fiber operativo",
	})
}
