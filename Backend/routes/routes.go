// Package routes registra y enlaza los endpoints con los controladores correspondientes.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers"
)

// SetupRoutes configura las rutas HTTP del servidor Fiber.
func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Autenticación
	api.Post("/login", controllers.Login)

	// Catálogo de Productos
	api.Get("/productos", controllers.GetProductos)
	api.Get("/productos/:id", controllers.GetProductoByID)

	// Red Multinivel
	api.Get("/red", controllers.GetRed)
}
