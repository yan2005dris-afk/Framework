package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func Login(context *fiber.Ctx) error {
	var req models.LoginRequest

	if err := context.BodyParser(&req); err != nil {
		return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cuerpo de petición inválido"})
	}

	user, err := repository.BuscarUsuarioPorCredenciales(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return context.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Credenciales incorrectas"})
		}
		return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error interno del servidor"})
	}

	return context.JSON(fiber.Map{
		"token": "fake-jwt-token-123",
		"email": user.Email,
		"rol":   user.Rol,
	})
}
