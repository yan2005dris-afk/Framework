// Package models define las estructuras de datos utilizadas en la aplicación.
package models

import "github.com/gofiber/fiber/v2"

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
	Activo      bool     `json:"activo"`
	Stock       int      `json:"stock"`
}

// ProductoRequest representa el payload para crear o editar un producto por admin.
type ProductoRequest struct {
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion"`
	Precio      float64  `json:"precio"`
	Categoria   string   `json:"categoria"`
	Img         string   `json:"img"`
	Galeria     []string `json:"galeria"`
	Stock       *int     `json:"stock,omitempty"`
	Activo      *bool    `json:"activo,omitempty"`
}

// APIError estandariza los mensajes de error HTTP de la API.
//
// Status es el codigo HTTP a retornar y NO se serializa en el body
// porque el cliente ya lo recibe en la response line. Solo Error y
// Detalle (este ultimo opcional) llegan al JSON, manteniendo el shape
// esperado por el frontend (`{"error": "...", "detalle": "..."}`).
//
// Use el metodo Write para responder consistentemente desde un controller:
//
//	return models.APIError{Status: 400, Error: "Cuerpo invalido"}.Write(c)
type APIError struct {
	Status  int
	Error   string
	Detalle interface{}
}

// Write serializa el error como JSON con el codigo HTTP correspondiente.
// Si Detalle es nil o string vacio se omite del body (regla de omitempty
// manejada manualmente para mantener el shape `{error, detalle?}` estable
// independientemente del tipo de Detalle).
func (e APIError) Write(c *fiber.Ctx) error {
	body := fiber.Map{"error": e.Error}
	if e.Detalle != nil {
		switch v := e.Detalle.(type) {
		case string:
			if v != "" {
				body["detalle"] = v
			}
		default:
			body["detalle"] = v
		}
	}
	return c.Status(e.Status).JSON(body)
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

// Comision representa una entrada en el ledger de comisiones devengadas.
type Comision struct {
	ID            int      `json:"id"`
	PedidoID      int      `json:"pedidoId"`
	ReferidoID    int      `json:"referidoId"`
	Nivel         int      `json:"nivel"`
	Porcentaje    float64  `json:"porcentaje"`
	Monto         float64  `json:"monto"`
	Estado        string   `json:"estado"`
	LiquidacionID *int     `json:"liquidacionId,omitempty"`
	CreadoEn      string   `json:"creadoEn"`
}

// ResumenComisiones resume los montos pendientes y pagados de un distribuidor.
type ResumenComisiones struct {
	ReferidoID      int        `json:"referidoId"`
	TotalPendiente  float64    `json:"totalPendiente"`
	TotalPagado     float64    `json:"totalPagado"`
	Comisiones      []Comision `json:"comisiones"`
}

// Liquidacion representa la consolidación y pago de comisiones pendientes.
type Liquidacion struct {
	ID         int     `json:"id"`
	ReferidoID int     `json:"referidoId"`
	Total      float64 `json:"total"`
	Estado     string  `json:"estado"`
	CreadoEn   string  `json:"creadoEn"`
	PagadaEn   string  `json:"pagadaEn"`
}

// LiquidarRequest define el payload para liquidar las comisiones de un referido.
type LiquidarRequest struct {
	ReferidoID int `json:"referidoId"`
}

// PedidoDetalleItem representa el desglose de un producto dentro del detalle de un pedido.
type PedidoDetalleItem struct {
	ID             int     `json:"id"`
	PedidoID       int     `json:"pedidoId"`
	ProductoID     int     `json:"productoId"`
	NombreProducto string  `json:"nombreProducto"`
	ImgProducto    string  `json:"imgProducto"`
	Cantidad       int     `json:"cantidad"`
	PrecioUnitario float64 `json:"precioUnitario"`
	Subtotal       float64 `json:"subtotal"`
}

// PedidoConDetalle representa la cabecera completa con todos sus items relacionados.
type PedidoConDetalle struct {
	Pedido
	Items []PedidoDetalleItem `json:"items"`
}

// ActualizarEstadoPedidoRequest define el payload para transicionar el estado de un pedido.
type ActualizarEstadoPedidoRequest struct {
	Estado string `json:"estado"`
}
