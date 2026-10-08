// Package models define las estructuras de datos utilizadas en la aplicación.
package models

// LoginRequest define el cuerpo esperado para la autenticación de usuarios.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Usuario representa la entidad de usuario almacenada en la base de datos.
type Usuario struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Rol      string `json:"rol"`
}

// Producto define la estructura completa de un artículo en el catálogo.
type Producto struct {
	ID          int      `json:"id"`
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion"`
	Precio      float64  `json:"precio"`
	Categoria   string   `json:"categoria"`
	Img         string   `json:"img"`
	Galeria     []string `json:"galeria"`
}

// Referido define la estructura jerárquica de un miembro en la red multinivel.
type Referido struct {
	ID     int        `json:"id"`
	Nombre string     `json:"nombre"`
	Nivel  int        `json:"nivel"`
	Ventas float64    `json:"ventas"`
	Hijos  []Referido `json:"hijos,omitempty"`
}

// ReferidoRow modela un registro plano de la tabla referidos para reconstrucción de jerarquía.
type ReferidoRow struct {
	ID       int
	Nombre   string
	Nivel    int
	Ventas   float64
	ParentID *int
}