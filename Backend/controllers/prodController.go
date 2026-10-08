// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/repository"
)

// GetProductos retorna la lista completa de productos desde la base de datos.
func GetProductos(c *fiber.Ctx) error {
	productos, err := repository.ObtenerProductos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al obtener productos",
		})
	}
	return c.JSON(productos)
}

// GetProductoByID retorna un único producto según el parámetro de ruta ID.
func GetProductoByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El ID del producto debe ser numérico",
		})
	}

	producto, err := repository.ObtenerProductoPorID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Producto no encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al consultar el producto",
		})
	}

	return c.JSON(producto)
}