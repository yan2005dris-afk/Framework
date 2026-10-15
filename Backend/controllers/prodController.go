// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetProductos retorna la lista completa de productos desde la base de datos con Cache-Aside.
func GetProductos(c *fiber.Ctx) error {
	ctx := context.Background()
	cacheKey := "productos:all"

	// 1. Intentar leer desde Redis (Cache Hit)
	cachedData, err := config.RedisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var productos []models.Producto
		if err := json.Unmarshal(cachedData, &productos); err == nil {
			return c.JSON(productos)
		}
	}

	// 2. Cache Miss: consultar PostgreSQL
	productos, err := repository.ObtenerProductos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al obtener productos",
		})
	}

	// 3. Guardar en Redis con TTL de 5 minutos
	data, err := json.Marshal(productos)
	if err == nil {
		config.RedisClient.Set(ctx, cacheKey, data, 5*time.Minute)
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
