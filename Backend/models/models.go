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

// APIError estandariza los mensajes de error HTTP de la API.
type APIError struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
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

// CheckoutItem representa un artículo en la solicitud de checkout.
type CheckoutItem struct {
	ProductoID int     `json:"productoId"`
	Cantidad   int     `json:"cantidad"`
	Precio     float64 `json:"precio"`
}

// CheckoutRequest define el cuerpo esperado para procesar una compra desde el frontend.
type CheckoutRequest struct {
	NombreCliente string         `json:"nombreCliente"`
	EmailCliente  string         `json:"emailCliente"`
	Direccion     string         `json:"direccion,omitempty"`
	Ciudad        string         `json:"ciudad,omitempty"`
	Telefono      string         `json:"telefono,omitempty"`
	MetodoPago    string         `json:"metodoPago"`
	Items         []CheckoutItem `json:"items"`
}

// CheckoutResponse representa la respuesta exitosa tras procesar el pedido.
type CheckoutResponse struct {
	PedidoID int     `json:"pedidoId"`
	Numero   string  `json:"numero"`
	Total    float64 `json:"total"`
	Mensaje  string  `json:"mensaje"`
}

// Pedido representa la cabecera de una compra registrada en base de datos.
type Pedido struct {
	ID            int       `json:"id"`
	UsuarioID     int       `json:"usuarioId"`
	NombreCliente string    `json:"nombreCliente"`
	EmailCliente  string    `json:"emailCliente"`
	MetodoPago    string    `json:"metodoPago"`
	Total         float64   `json:"total"`
	Estado        string    `json:"estado"`
	CreadoAt      string    `json:"creadoAt"`
}

// PedidoItem representa un item asociado a un pedido en base de datos.
type PedidoItem struct {
	ID             int     `json:"id"`
	PedidoID       int     `json:"pedidoId"`
	ProductoID     int     `json:"productoId"`
	Cantidad       int     `json:"cantidad"`
	PrecioUnitario float64 `json:"precioUnitario"`
	Subtotal       float64 `json:"subtotal"`
}
