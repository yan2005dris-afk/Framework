package controllers

import (
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func Login(context *fiber.Ctx) error {
	var req models.LoginRequest

	if err := context.BodyParser(&req); err != nil {
		return context.Status(400).JSON(fiber.Map{"error": "Cuerpo de petición inválido"})
	}

	switch {
	case req.Email == "admin@upse.edu.ec" && req.Password == "123456":
		return context.JSON(fiber.Map{"token": "fake-jwt-token-123", "email":req.Email, "rol": "admin"})

	case req.Email == "cliente@upse.edu.ec" && req.Password == "123456":
		return context.JSON(fiber.Map{"token": "fake-jwt-token-123", "email":req.Email, "rol": "cliente"})
	default:
		return context.Status(401).JSON(fiber.Map{"error": "Credenciales incorrectas"})
	}
}
