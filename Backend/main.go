package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"multicatalogo-backend/config"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/requestlogger"
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

	// Variables de entorno de seguridad (U2-T4 + GA5).
	// Post-GA5 el Backend NO firma JWTs localmente, pero igual valida
	// tokens via middleware.RequireAuth (defensa en profundidad), asi que
	// sigue necesitando el secret. AUTH_JWT_SECRET y JWT_SECRET son
	// equivalentes; preferimos AUTH_JWT_SECRET por consistencia con el
	// auth-service.
	jwtSecret := os.Getenv("AUTH_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	// CORS_ORIGINS lista separada por comas de orígenes permitidos.
	// El default es el dev server de Vite; en producción hay que setearlo
	// explícitamente. AllowCredentials=true es obligatorio para que las
	// cookies httpOnly viajen desde el frontend.
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173"
	}

	// AUTH_SERVICE_URL: URL del auth-service para que los handlers
	// /api/login, /api/logout, /api/refresh proxyen. En compose es
	// http://auth-service:3001 (resoluble por DNS Docker). En dev
	// nativo, http://localhost:3001.
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:3001"
	}

	// Instanciación de controladores con dependencias inyectadas
	authCtrl := controllers.NewAuthController(authServiceURL)
	prodCtrl := controllers.NewProductoController(dbRepo, cacheRepo, appEnv)
	redCtrl := controllers.NewRedController(dbRepo, cacheRepo, appEnv)
	checkoutCtrl := controllers.NewCheckoutController(dbRepo, cacheRepo, appEnv)
	healthCtrl := controllers.NewHealthController(dbPool, redisClient)
	liqCtrl := controllers.NewLiquidacionController(dbRepo, appEnv)

	app := fiber.New()

	// Log estructurado en JSON (#21): una linea por request. Los campos
	// comunes con el SaaS (ts en UTC con sufijo Z, level, service) son los que
	// permiten agregar los logs de los dos servicios en una sola linea de
	// tiempo.
	app.Use(requestlogger.New(os.Stdout))

	// CORS con credenciales habilitadas (necesario para que la cookie httpOnly
	// viaje en el cross-puerto dev 5173 → 3000). En producción no se acepta
	// "*" como origen credencial.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))

	// Registramos las rutas inyectando los controladores
	routes.SetupRoutes(app, authCtrl, prodCtrl, redCtrl, checkoutCtrl, healthCtrl, liqCtrl, dbRepo, jwtSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("🚀 Servidor Backend iniciado en http://localhost:%s (APP_ENV=%s)\n", port, appEnv)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
