// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
)

// redInicialData contiene el árbol jerárquico de referidos multinivel.
var redInicialData = models.Referido{
	ID:     0,
	Nombre: "Tú",
	Nivel:  0,
	Ventas: 2400,
	Hijos: []models.Referido{
		{
			ID:     1,
			Nombre: "Ana García",
			Nivel:  1,
			Ventas: 1200,
			Hijos: []models.Referido{
				{
					ID:     4,
					Nombre: "Carlos Ruiz",
					Nivel:  2,
					Ventas: 500,
					Hijos: []models.Referido{
						{ID: 7, Nombre: "Diana Paz", Nivel: 3, Ventas: 300},
					},
				},
				{ID: 5, Nombre: "Sofía León", Nivel: 2, Ventas: 430},
			},
		},
		{
			ID:     2,
			Nombre: "Luis Poveda",
			Nivel:  1,
			Ventas: 850,
			Hijos: []models.Referido{
				{ID: 6, Nombre: "Marco Díaz", Nivel: 2, Ventas: 380},
			},
		},
		{ID: 3, Nombre: "Marta Sánchez", Nivel: 1, Ventas: 430},
	},
}

// GetRed retorna la estructura jerárquica de la red de referidos.
func GetRed(c *fiber.Ctx) error {
	return c.JSON(redInicialData)
}
