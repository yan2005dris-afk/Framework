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

	// Autenticación (pública)
	api.Post("/login", authCtrl.Login)

	// Catálogo de Productos (público)
	api.Get("/productos", prodCtrl.GetProductos)
	api.Get("/productos/:id", prodCtrl.GetProductoByID)

	// Rutas protegidas con JWT (Authorization: Bearer <token>).
	// El middleware parsea el token, setea c.Locals y, si DBRepo no es nil,
	// hidrata el usuarioID desde la BD para que el checkout pueda usarlo
	// como identidad (ignora emailCliente del body).
	auth := middleware.RequireAuth(middleware.AuthConfig{
		JWTSecret: []byte(jwtSecret),
		DBRepo:    dbRepo,
	})

	// Red Multinivel (protegida)
	api.Get("/red", auth, redCtrl.GetRed)

	// Checkout / Transacciones ACID (protegido)
	api.Post("/checkout", auth, checkoutCtrl.Checkout)
}

// SetupDefaultRoutes permite inicializar rutas instanciando los repositorios y controllers
// por defecto, útil para tests que no requieren BD real.
func SetupDefaultRoutes(app *fiber.App) {
	dbRepo := &repository.PostgresRepository{}
	cacheRepo := &repository.RedisCacheRepository{}

	authCtrl := controllers.NewAuthController(dbRepo, "dev-secret")
	prodCtrl := controllers.NewProductoController(dbRepo, cacheRepo, "dev")
	redCtrl := controllers.NewRedController(dbRepo, cacheRepo, "dev")
	checkoutCtrl := controllers.NewCheckoutController(dbRepo, cacheRepo, "dev")
	healthCtrl := controllers.NewHealthController(nil, nil)

	SetupRoutes(app, authCtrl, prodCtrl, redCtrl, checkoutCtrl, healthCtrl, dbRepo, "dev-secret")
}
