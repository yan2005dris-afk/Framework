// Package controllers implementa la lógica de negocio para los endpoints de la API.
package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
)

// productosData almacena el catálogo de productos predefinidos.
var productosData = []models.Producto{
	{
		ID:          1,
		Nombre:      "Serum Revitalizante",
		Descripcion: "Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.",
		Precio:      45.0,
		Categoria:   "Serum",
		Img:         "https://picsum.photos/seed/serum/600",
		Galeria: []string{
			"https://picsum.photos/seed/serum1/600",
			"https://picsum.photos/seed/serum2/600",
			"https://picsum.photos/seed/serum3/600",
		},
	},
	{
		ID:          2,
		Nombre:      "Crema Hidratante Pro",
		Descripcion: "Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.",
		Precio:      32.5,
		Categoria:   "Crema",
		Img:         "https://picsum.photos/seed/crema/600",
		Galeria: []string{
			"https://picsum.photos/seed/crema1/600",
			"https://picsum.photos/seed/crema2/600",
			"https://picsum.photos/seed/crema3/600",
		},
	},
	{
		ID:          3,
		Nombre:      "Tónico Purificante",
		Descripcion: "Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.",
		Precio:      28.0,
		Categoria:   "Tónico",
		Img:         "https://picsum.photos/seed/tonico/600",
		Galeria: []string{
			"https://picsum.photos/seed/tonico1/600",
			"https://picsum.photos/seed/tonico2/600",
			"https://picsum.photos/seed/tonico3/600",
		},
	},
	{
		ID:          4,
		Nombre:      "Mascarilla Nocturna",
		Descripcion: "Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.",
		Precio:      50.0,
		Categoria:   "Mascarilla",
		Img:         "https://picsum.photos/seed/mascarilla/600",
		Galeria: []string{
			"https://picsum.photos/seed/mascarilla1/600",
			"https://picsum.photos/seed/mascarilla1/600",
			"https://picsum.photos/seed/mascarilla2/600",
			"https://picsum.photos/seed/mascarilla3/600",
		},
	},
	{
		ID:          5,
		Nombre:      "Serum Anti-Edad Retinol",
		Descripcion: "Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.",
		Precio:      58.0,
		Categoria:   "Serum",
		Img:         "https://picsum.photos/seed/retinol/600",
		Galeria: []string{
			"https://picsum.photos/seed/retinol1/600",
			"https://picsum.photos/seed/retinol2/600",
			"https://picsum.photos/seed/retinol3/600",
		},
	},
	{
		ID:          6,
		Nombre:      "Crema Contorno de Ojos",
		Descripcion: "Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.",
		Precio:      26.0,
		Categoria:   "Crema",
		Img:         "https://picsum.photos/seed/ojos/600",
		Galeria: []string{
			"https://picsum.photos/seed/ojos1/600",
			"https://picsum.photos/seed/ojos2/600",
			"https://picsum.photos/seed/ojos3/600",
		},
	},
	{
		ID:          7,
		Nombre:      "Tónico Exfoliante AHA-BHA",
		Descripcion: "Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.",
		Precio:      34.0,
		Categoria:   "Tónico",
		Img:         "https://picsum.photos/seed/exfoliante/600",
		Galeria: []string{
			"https://picsum.photos/seed/exfoliante1/600",
			"https://picsum.photos/seed/exfoliante2/600",
			"https://picsum.photos/seed/exfoliante3/600",
		},
	},
	{
		ID:          8,
		Nombre:      "Kit Rutina Completa",
		Descripcion: "Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.",
		Precio:      129.0,
		Categoria:   "Kit",
		Img:         "https://picsum.photos/seed/kit/600",
		Galeria: []string{
			"https://picsum.photos/seed/kit1/600",
			"https://picsum.photos/seed/kit2/600",
			"https://picsum.photos/seed/kit3/600",
		},
	},
}

// GetProductos retorna la lista completa de productos.
func GetProductos(c *fiber.Ctx) error {
	return c.JSON(productosData)
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

	for _, p := range productosData {
		if p.ID == id {
			return c.JSON(p)
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "Producto no encontrado",
	})
}