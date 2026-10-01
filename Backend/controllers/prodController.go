// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetProductos retorna la lista completa de productos.
func GetProductos(c *fiber.Ctx) error {
	return c.JSON(repository.GetProductos())
}

// GetProductoByID retorna un único producto según el parámetro de ruta ID.
func GetProductoByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El ID del producto debe ser numérico",
			Details: idParam,
		})
	}

	producto, encontrado := repository.GetProductoByID(id)
	if !encontrado {
		return c.Status(fiber.StatusNotFound).JSON(models.APIError{
			Status:  fiber.StatusNotFound,
			Message: "Producto no encontrado",
		})
	}

	return c.JSON(producto)
}
