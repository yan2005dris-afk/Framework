// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/repository"
)

// GetRed retorna la estructura jerárquica de la red de referidos desde la base de datos.
func GetRed(c *fiber.Ctx) error {
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
	return c.JSON(red)
}
