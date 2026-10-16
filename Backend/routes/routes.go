// Package routes registra y enlaza los endpoints con los controladores correspondientes usando DI.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/repository"
)

// SetupRoutes configura las rutas HTTP del servidor Fiber con dependencias inyectadas.
func SetupRoutes(
	app *fiber.App,
	authCtrl *controllers.AuthController,
	prodCtrl *controllers.ProductoController,
	redCtrl *controllers.RedController,
	checkoutCtrl *controllers.CheckoutController,
) {
	api := app.Group("/api")

	// Estado del servidor (endpoint package-level, sin estado mutable)
	api.Get("/health", controllers.GetHealth)

	// Autenticación
	api.Post("/login", authCtrl.Login)

	// Catálogo de Productos
	api.Get("/productos", prodCtrl.GetProductos)
	api.Get("/productos/:id", prodCtrl.GetProductoByID)

	// Red Multinivel
	api.Get("/red", redCtrl.GetRed)

	// Checkout / Transacciones ACID (Fase 4)
	api.Post("/checkout", checkoutCtrl.Checkout)
}

// SetupDefaultRoutes permite inicializar rutas instanciando los repositorios por defecto.
// Usa un secreto de JWT placeholder; en producción llamar a SetupRoutes directamente
// pasando dependencias reales y el JWT_SECRET de env.
func SetupDefaultRoutes(app *fiber.App) {
	dbRepo := &repository.PostgresRepository{}
	cacheRepo := &repository.RedisCacheRepository{}

	authCtrl := controllers.NewAuthController(dbRepo, "dev-secret")
	prodCtrl := controllers.NewProductoController(dbRepo, cacheRepo)
	redCtrl := controllers.NewRedController(dbRepo, cacheRepo)
	checkoutCtrl := controllers.NewCheckoutController(dbRepo, cacheRepo)

	SetupRoutes(app, authCtrl, prodCtrl, redCtrl, checkoutCtrl)
}
