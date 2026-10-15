// Package controllers implementa los controladores de la aplicación con inyección de dependencias.
package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// AuthController maneja la autenticación de usuarios.
type AuthController struct {
	dbRepo repository.DatabaseRepository
}

// NewAuthController instancia un AuthController con dependencias inyectadas.
func NewAuthController(dbRepo repository.DatabaseRepository) *AuthController {
	return &AuthController{dbRepo: dbRepo}
}

// Login procesa las credenciales del usuario y retorna un token JWT simulado y su rol.
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cuerpo de petición inválido"})
	}

	user, err := ac.dbRepo.BuscarUsuarioPorCredenciales(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Credenciales incorrectas"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error interno del servidor"})
	}

	return c.JSON(fiber.Map{
		"token": "fake-jwt-token-123",
		"email": user.Email,
		"rol":   user.Rol,
	})
}

// Compatibilidad hacia atrás
func Login(context *fiber.Ctx) error {
	controller := NewAuthController(&repository.PostgresRepository{})
	return controller.Login(context)
}
