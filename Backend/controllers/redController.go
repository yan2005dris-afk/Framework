package controllers

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/repository"
)

// GetRed retorna la estructura jerárquica de la red de referidos.
func GetRed(c *fiber.Ctx) error {
	return c.JSON(repository.GetRed())
}
