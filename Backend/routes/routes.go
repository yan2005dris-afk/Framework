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

	// Estado del servidor
	api.Get("/health", controllers.GetHealth)

	// Autenticación
	if authCtrl != nil {
		api.Post("/login", authCtrl.Login)
	} else {
		api.Post("/login", controllers.Login)
	}

	// Catálogo de Productos
	if prodCtrl != nil {
		api.Get("/productos", prodCtrl.GetProductos)
		api.Get("/productos/:id", prodCtrl.GetProductoByID)
	} else {
		api.Get("/productos", controllers.GetProductos)
		api.Get("/productos/:id", controllers.GetProductoByID)
	}

	// Red Multinivel
	if redCtrl != nil {
		api.Get("/red", redCtrl.GetRed)
	} else {
		api.Get("/red", controllers.GetRed)
	}

	// Checkout / Transacciones ACID (Fase 4)
	if checkoutCtrl != nil {
		api.Post("/checkout", checkoutCtrl.Checkout)
	}
}

// SetupDefaultRoutes permite inicializar rutas instanciando los repositorios por defecto.
func SetupDefaultRoutes(app *fiber.App) {
	dbRepo := &repository.PostgresRepository{}
	cacheRepo := &repository.RedisCacheRepository{}

	authCtrl := controllers.NewAuthController(dbRepo)
	prodCtrl := controllers.NewProductoController(dbRepo, cacheRepo)
	redCtrl := controllers.NewRedController(dbRepo, cacheRepo)
	checkoutCtrl := controllers.NewCheckoutController(dbRepo, cacheRepo)

	SetupRoutes(app, authCtrl, prodCtrl, redCtrl, checkoutCtrl)
}
