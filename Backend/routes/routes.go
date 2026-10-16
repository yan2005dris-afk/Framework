// Package routes registra y enlaza los endpoints con los controladores correspondientes usando DI.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/repository"
)

// SetupRoutes configura las rutas HTTP del servidor Fiber con dependencias inyectadas.
// jwtSecret se usa para firmar y verificar el JWT en las rutas protegidas.
func SetupRoutes(
	app *fiber.App,
	authCtrl *controllers.AuthController,
	prodCtrl *controllers.ProductoController,
	redCtrl *controllers.RedController,
	checkoutCtrl *controllers.CheckoutController,
	healthCtrl *controllers.HealthController,
	dbRepo repository.DatabaseRepository,
	jwtSecret string,
) {
	api := app.Group("/api")

	// Estado del servidor (público; hace ping real a PG y Redis)
	api.Get("/health", healthCtrl.Health)

	// Catálogo de Productos (público)
	api.Get("/productos", prodCtrl.GetProductos)
	api.Get("/productos/:id", prodCtrl.GetProductoByID)

	// Autenticación (pública)
	api.Post("/login", authCtrl.Login)
	api.Post("/logout", authCtrl.Logout)

	// Middleware de proteccion con JWT. Lee el token desde el header
	// `Authorization: Bearer <jwt>` (override) o desde la cookie httpOnly
	// `auth_token` (caso normal del flujo SPA). Setea c.Locals con email,
	// rol y (si DBRepo no es nil) usuarioID.
	requireAuth := middleware.RequireAuth(middleware.AuthConfig{
		JWTSecret: []byte(jwtSecret),
		DBRepo:    dbRepo,
	})

	// Hidratación de sesión desde cookie httpOnly (protegido).
	// El frontend lo llama al montar la app para saber si hay sesión activa
	// sin depender de localStorage.
	api.Get("/me", requireAuth, authCtrl.Me)

	// Red Multinivel (protegida)
	api.Get("/red", requireAuth, redCtrl.GetRed)

	// Checkout / Transacciones ACID (protegido)
	api.Post("/checkout", requireAuth, checkoutCtrl.Checkout)
}

// SetupDefaultRoutes permite inicializar rutas instanciando los repositorios y controllers
// por defecto, útil para tests que no requieren BD real.
func SetupDefaultRoutes(app *fiber.App) {
	dbRepo := &repository.PostgresRepository{}
	cacheRepo := &repository.RedisCacheRepository{}

	// Tests sin auth-service: el proxy devuelve 502 BadGateway porque no hay
	// URL arriba; suficiente para validar que el handler existe y responde.
	authCtrl := controllers.NewAuthController("http://127.0.0.1:1")
	prodCtrl := controllers.NewProductoController(dbRepo, cacheRepo, "dev")
	redCtrl := controllers.NewRedController(dbRepo, cacheRepo, "dev")
	checkoutCtrl := controllers.NewCheckoutController(dbRepo, cacheRepo, "dev")
	healthCtrl := controllers.NewHealthController(nil, nil)

	SetupRoutes(app, authCtrl, prodCtrl, redCtrl, checkoutCtrl, healthCtrl, dbRepo, "dev-secret")
}
