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
	// CrearPedido registra el pedido en una transacción ACID.
	//
	// idempotencyKey activa la idempotencia (U4-T3, S10): vacía es el
	// comportamiento de siempre, y una key ya usada por ese usuario devuelve
	// el pedido que creó, con replay en true (el controller responde 200 en vez
	// de 201 para que el cliente no lo tome por un pedido nuevo).
	CrearPedido(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error)
	ActualizarComisiones(usuarioID int, totalPedido float64) error
}

// CacheRepository define los métodos requeridos para interactuar con la capa de caché en memoria.
type CacheRepository interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttlSeconds int) error
	Del(key string) error
}
