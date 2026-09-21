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
		AllowOrigins: "http://localhost:5173, http://127.0.0.1:5173, http://172.17.82.108:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
	}))

	// Registramos las rutas
	routes.SetupRoutes(app)

	fmt.Println("🚀 Servidor Backend iniciado en http://localhost:3000")

	// Escuchamos en el puerto 3000
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}