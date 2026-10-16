// Package repository define las abstracciones e interfaces para persistencia y caché.
package repository

import (
	"multicatalogo-backend/models"
)

// DatabaseRepository define los métodos requeridos para operaciones en la base de datos relacional.
type DatabaseRepository interface {
	// BuscarUsuarioPorEmail trae al usuario (incluido el hash bcrypt del password) por email.
	// La verificación del password se hace con bcrypt.CompareHashAndPassword en el controller,
	// nunca en SQL.
	BuscarUsuarioPorEmail(email string) (*models.Usuario, error)
	ObtenerProductos() ([]models.Producto, error)
	ObtenerProductoPorID(id int) (*models.Producto, error)
	ObtenerRed() (*models.Referido, error)
	CrearPedido(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error)
	ActualizarComisiones(usuarioID int, totalPedido float64) error
}

// CacheRepository define los métodos requeridos para interactuar con la capa de caché en memoria.
type CacheRepository interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttlSeconds int) error
	Del(key string) error
}
