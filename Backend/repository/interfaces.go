// Package repository define las abstracciones e interfaces para persistencia y caché.
package repository

import (
	"multicatalogo-backend/models"
)

// DatabaseRepository define los métodos requeridos para operaciones en la base de datos relacional.
type DatabaseRepository interface {
	// BuscarUsuarioPorEmail trae al usuario (incluido el hash bcrypt del password) por email.
	BuscarUsuarioPorEmail(email string) (*models.Usuario, error)
	RegistrarUsuario(req models.RegisterRequest, hashedPassword string) (*models.Usuario, error)
	ObtenerProductos() ([]models.Producto, error)
	ObtenerProductosAdmin() ([]models.Producto, error)
	ObtenerProductoPorID(id int) (*models.Producto, error)
	CrearProducto(req models.ProductoRequest) (*models.Producto, error)
	ActualizarProducto(id int, req models.ProductoRequest) (*models.Producto, error)
	EliminarProductoLogico(id int) error
	ObtenerRed() (*models.Referido, error)
	// CrearPedido registra el pedido en una transacción ACID.
	CrearPedido(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error)
	ListarPedidosMios(usuarioID int, limit, offset int) ([]models.Pedido, error)
	ListarPedidosAdmin(estado string, limit, offset int) ([]models.Pedido, error)
	ObtenerPedidoDetalle(id int) (*models.PedidoConDetalle, error)
	ActualizarEstadoPedido(id int, nuevoEstado string) error
	ActualizarComisiones(usuarioID int, totalPedido float64) error

	// Métodos para Ledger de Comisiones y Liquidaciones
	RegistrarComisionesPedido(pedidoID int, compradorID int, totalPedido float64) ([]models.Comision, error)
	ObtenerComisionesPorReferido(referidoID int) (*models.ResumenComisiones, error)
	LiquidarComisiones(referidoID int) (*models.Liquidacion, error)
	ListarLiquidaciones() ([]models.Liquidacion, error)
	ListarLiquidacionesPorReferido(referidoID int) ([]models.Liquidacion, error)

	// Métodos de Reportes y KPIs (Fase 6)
	ObtenerReporteResumen(desde, hasta string) (*models.ReporteResumen, error)
	ObtenerTopProductos(limit int) ([]models.TopProducto, error)
	ObtenerTopReferidos(limit int) ([]models.TopReferido, error)
	ObtenerComisionesPendientes() ([]models.DeudaDistribuidor, error)
}

// CacheRepository define los métodos requeridos para interactuar con la capa de caché en memoria.
type CacheRepository interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttlSeconds int) error
	Del(key string) error
}
