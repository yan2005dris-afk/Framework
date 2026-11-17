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
	liqCtrl *controllers.LiquidacionController,
	pedCtrl *controllers.PedidoController,
	repCtrl *controllers.ReporteController,
	dbRepo repository.DatabaseRepository,
	jwtSecret string,
) {
	api := app.Group("/api")

	// Estado del servidor (público; hace ping real a PG y Redis)
	api.Get("/health", healthCtrl.Health)

	// Middleware de proteccion con JWT. Lee el token desde el header
	// `Authorization: Bearer <jwt>` (override) o desde la cookie httpOnly
	// `auth_token` (caso normal del flujo SPA). Setea c.Locals con email,
	// rol y (si DBRepo no es nil) usuarioID.
	requireAuth := middleware.RequireAuth(middleware.AuthConfig{
		JWTSecret: []byte(jwtSecret),
		DBRepo:    dbRepo,
	})

	// Catálogo de Productos (público)
	api.Get("/productos", prodCtrl.GetProductos)
	api.Get("/productos/:id", prodCtrl.GetProductoByID)

	// Administración de Productos (Fase 2, protegido admin)
	api.Get("/admin/productos", requireAuth, prodCtrl.GetProductosAdmin)
	api.Post("/productos", requireAuth, prodCtrl.CrearProducto)
	api.Put("/productos/:id", requireAuth, prodCtrl.ActualizarProducto)
	api.Delete("/productos/:id", requireAuth, prodCtrl.EliminarProducto)

	// Autenticación y Registro (públicos)
	api.Post("/login", authCtrl.Login)
	api.Post("/logout", authCtrl.Logout)
	api.Post("/register", authCtrl.Register)

	// Hidratación de sesión desde cookie httpOnly (protegido).
	// El frontend lo llama al montar la app para saber si hay sesión activa
	// sin depender de localStorage.
	api.Get("/me", requireAuth, authCtrl.Me)

	// Red Multinivel (protegida)
	api.Get("/red", requireAuth, redCtrl.GetRed)

	// Checkout / Transacciones ACID (protegido)
	api.Post("/checkout", requireAuth, checkoutCtrl.Checkout)

	// Gestión de Pedidos (Fase 3)
	api.Get("/pedidos/mios", requireAuth, pedCtrl.ListarPedidosMios)
	api.Get("/pedidos", requireAuth, pedCtrl.ListarPedidosAdmin)
	api.Get("/pedidos/:id", requireAuth, pedCtrl.ObtenerPedidoPorID)
	api.Patch("/pedidos/:id/estado", requireAuth, pedCtrl.ActualizarEstado)

	// Ledger de Comisiones y Liquidaciones (Fase 4 y Fase 5)
	api.Get("/comisiones/mias", requireAuth, liqCtrl.ObtenerComisionesMias)
	api.Get("/liquidaciones/mias", requireAuth, liqCtrl.ListarLiquidacionesMias)
	api.Get("/liquidaciones", requireAuth, liqCtrl.ListarLiquidaciones)
	api.Post("/liquidaciones", requireAuth, liqCtrl.Liquidar)

	// Reportes y Dashboard de Negocio (Fase 6, admin)
	api.Get("/reportes/resumen", requireAuth, repCtrl.Resumen)
	api.Get("/reportes/top-productos", requireAuth, repCtrl.TopProductos)
	api.Get("/reportes/top-referidos", requireAuth, repCtrl.TopReferidos)
	api.Get("/reportes/comisiones-pendientes", requireAuth, repCtrl.ComisionesPendientes)
}

// SetupDefaultRoutes permite inicializar rutas instanciando los repositorios y controllers
// por defecto, útil para tests que no requieren BD real.
func SetupDefaultRoutes(app *fiber.App) {
	dbRepo := &repository.PostgresRepository{}
	cacheRepo := &repository.RedisCacheRepository{}

	// Tests sin auth-service: el proxy devuelve 502 BadGateway porque no hay
	// URL arriba; suficiente para validar que el handler existe y responde.
	authCtrl := controllers.NewAuthController("http://127.0.0.1:1", dbRepo, "dev")
	prodCtrl := controllers.NewProductoController(dbRepo, cacheRepo, "dev")
	redCtrl := controllers.NewRedController(dbRepo, cacheRepo, "dev")
	checkoutCtrl := controllers.NewCheckoutController(dbRepo, cacheRepo, "dev")
	healthCtrl := controllers.NewHealthController(nil, nil)
	liqCtrl := controllers.NewLiquidacionController(dbRepo, "dev")
	pedCtrl := controllers.NewPedidoController(dbRepo, "dev")
	repCtrl := controllers.NewReporteController(dbRepo, "dev")

	SetupRoutes(app, authCtrl, prodCtrl, redCtrl, checkoutCtrl, healthCtrl, liqCtrl, pedCtrl, repCtrl, dbRepo, "dev-secret")
}
