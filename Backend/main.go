package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"multicatalogo-backend/routes"
)

func main() {
	app := fiber.New()

	// Middleware de logger para ver las peticiones en consola
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))

	// Implementamos el middleware CORS a nivel global
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
	}))

	// Registramos las rutas
	routes.SetupRoutes(app)

	fmt.Println("🚀 Servidor Backend iniciado en http://localhost:8080")

	// Escuchamos en el puerto 8080 (el puerto 3000 estaba ocupado por otro servicio del sistema)
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}