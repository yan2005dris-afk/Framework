package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"multicatalogo-backend/config"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/routes"
)

func main() {
	// Conexión a la base de datos PostgreSQL
	dbPool, err := config.ConnectDB()
	if err != nil {
		log.Printf("⚠️ Advertencia: No se pudo conectar a PostgreSQL: %v", err)
	} else {
		defer config.CloseDB()
	}

	// Conexión a Redis
	redisClient, err := config.ConnectRedis()
	if err != nil {
		log.Printf("⚠️ Advertencia: No se pudo conectar a Redis: %v", err)
	} else {
		defer config.CloseRedis()
	}

	// Instanciación de repositorios (Inyección de Dependencias)
	dbRepo := repository.NewPostgresRepository(dbPool)
	cacheRepo := repository.NewRedisCacheRepository(redisClient)

	// Instanciación de controladores con dependencias inyectadas
	authCtrl := controllers.NewAuthController(dbRepo)
	prodCtrl := controllers.NewProductoController(dbRepo, cacheRepo)
	redCtrl := controllers.NewRedController(dbRepo, cacheRepo)
	checkoutCtrl := controllers.NewCheckoutController(dbRepo, cacheRepo)

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

	// Registramos las rutas inyectando los controladores
	routes.SetupRoutes(app, authCtrl, prodCtrl, redCtrl, checkoutCtrl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Servidor Backend iniciado en http://localhost:%s\n", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}