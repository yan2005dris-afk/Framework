// auth-service entrypoint.
//
// Servicio de autenticación extraido del Backend Go/Fiber (U3-T2).
// Centraliza emision y validacion de JWTs para que Backend y SaaS no
// dupliquen la logica de bcrypt/JWT/cookies. El gateway enruta sus
// requests a este servicio via forward_auth (/auth/validate).
//
// Estado actual (GA1+GA2):
//   - go.mod + deps
//   - config desde env
//   - pool de Postgres en :5436 (compartida via schema 'auth')
//   - /auth/health con ping a la BD
//   - /auth/login, /auth/logout, /auth/refresh, /auth/validate, /auth/me
//
// GA3+ crea el schema 'auth' y el seed inicial (ver migrations/).
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/handlers"
)

func main() {
	cfg := config.Load()

	// Pool PG. Init lazy: si falla, el servicio arranca igual y
	// /auth/health devuelve "degraded"; GA2+ agrega Init con retry.
	database := db.New(cfg)
	if err := database.Init(cfg); err != nil {
		log.Printf("⚠️  auth-db no responde (continuamos en degraded): %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:               "multicatalogo-auth",
		DisableStartupMessage: true,
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
	}))

	// Handlers
	hHealth := &handlers.HealthHandler{DB: database}
	app.Get("/auth/health", hHealth.Health)

	if database.Pool != nil {
		hAuth := handlers.NewAuthHandler(cfg, database.Pool)
		app.Post("/auth/login", hAuth.Login)
		app.Post("/auth/logout", hAuth.Logout)
		app.Post("/auth/refresh", hAuth.Refresh)
		app.Post("/auth/validate", hAuth.Validate)
		// El gateway llama a /auth/validate con GET (ver el forward_auth del
		// Caddyfile): con POST, Caddy consumia el body del request original y
		// el checkout quedaba en 502. GET y POST quedan registrados para que
		// el endpoint no dependa de como lo invoque cada cliente.
		app.Get("/auth/validate", hAuth.Validate)
		app.Get("/auth/me", hAuth.Me)
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "multicatalogo-auth",
			"docs":    "/auth/health",
		})
	})

	log.Printf("🚀 auth-service en :3001 (APP_ENV=%s)", cfg.AppEnv)
	log.Fatal(app.Listen(":3001"))
}
