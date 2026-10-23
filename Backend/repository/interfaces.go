// Package repository define las abstracciones e interfaces para persistencia y caché.
package repository

import (
	"multicatalogo-backend/models"
)

// DatabaseRepository define los métodos requeridos para operaciones en la base de datos relacional.
type DatabaseRepository interface {
	// BuscarUsuarioPorEmail trae al usuario (incluido el hash bcrypt del password) por email.
	BuscarUsuarioPorEmail(email string) (*models.Usuario, error)
	ObtenerProductos() ([]models.Producto, error)
	ObtenerProductosAdmin() ([]models.Producto, error)
	ObtenerProductoPorID(id int) (*models.Producto, error)
	CrearProducto(req models.ProductoRequest) (*models.Producto, error)
	ActualizarProducto(id int, req models.ProductoRequest) (*models.Producto, error)
	EliminarProductoLogico(id int) error
	ObtenerRed() (*models.Referido, error)
	// CrearPedido registra el pedido en una transacción ACID.
	CrearPedido(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error)
	ActualizarComisiones(usuarioID int, totalPedido float64) error

	// Métodos para Ledger de Comisiones y Liquidaciones
	RegistrarComisionesPedido(pedidoID int, compradorID int, totalPedido float64) ([]models.Comision, error)
	ObtenerComisionesPorReferido(referidoID int) (*models.ResumenComisiones, error)
	LiquidarComisiones(referidoID int) (*models.Liquidacion, error)
	ListarLiquidaciones() ([]models.Liquidacion, error)
	ListarLiquidacionesPorReferido(referidoID int) ([]models.Liquidacion, error)
}

// CacheRepository define los métodos requeridos para interactuar con la capa de caché en memoria.
type CacheRepository interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttlSeconds int) error
	Del(key string) error
}
