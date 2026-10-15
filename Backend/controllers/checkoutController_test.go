package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
)

// MockDatabaseRepository simula las operaciones de base de datos.
type MockDatabaseRepository struct {
	CrearPedidoFunc         func(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error)
	ActualizarComisionesFunc func(usuarioID int, totalPedido float64) error
}

func (m *MockDatabaseRepository) BuscarUsuarioPorCredenciales(email, password string) (*models.Usuario, error) {
	return nil, nil
}
func (m *MockDatabaseRepository) ObtenerProductos() ([]models.Producto, error) {
	return nil, nil
}
func (m *MockDatabaseRepository) ObtenerProductoPorID(id int) (*models.Producto, error) {
	return nil, nil
}
func (m *MockDatabaseRepository) ObtenerRed() (*models.Referido, error) {
	return nil, nil
}
func (m *MockDatabaseRepository) CrearPedido(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
	if m.CrearPedidoFunc != nil {
		return m.CrearPedidoFunc(usuarioID, req)
	}
	return &models.Pedido{ID: 101, Total: 100}, nil
}
func (m *MockDatabaseRepository) ActualizarComisiones(usuarioID int, totalPedido float64) error {
	if m.ActualizarComisionesFunc != nil {
		return m.ActualizarComisionesFunc(usuarioID, totalPedido)
	}
	return nil
}

// MockCacheRepository simula las operaciones de Redis.
type MockCacheRepository struct {
	delCalled bool
	lastDelKey string
}

func (m *MockCacheRepository) Get(key string) ([]byte, error) {
	return nil, errors.New("miss")
}
func (m *MockCacheRepository) Set(key string, value []byte, ttl int) error {
	return nil
}
func (m *MockCacheRepository) Del(key string) error {
	m.delCalled = true
	m.lastDelKey = key
	return nil
}

func TestCheckoutController_Exitoso(t *testing.T) {
	app := fiber.New()
	mockDB := &MockDatabaseRepository{
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
			return &models.Pedido{
				ID:            42,
				UsuarioID:     usuarioID,
				NombreCliente: req.NombreCliente,
				EmailCliente:  req.EmailCliente,
				Total:         75.5,
				Estado:        "completado",
			}, nil
		},
	}
	mockCache := &MockCacheRepository{}
	ctrl := NewCheckoutController(mockDB, mockCache)

	app.Post("/api/checkout", ctrl.Checkout)

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Estudiante UPSE",
		EmailCliente:  "estudiante@upse.edu.ec",
		MetodoPago:    "Tarjeta de crédito",
		Items: []models.CheckoutItem{
			{ProductoID: 1, Cantidad: 1, Precio: 45.0},
			{ProductoID: 2, Cantidad: 1, Precio: 30.5},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error al ejecutar petición de prueba: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Se esperaba código 201, se obtuvo %d", resp.StatusCode)
	}

	var res models.CheckoutResponse
	_ = json.NewDecoder(resp.Body).Decode(&res)
	if res.PedidoID != 42 {
		t.Errorf("Se esperaba PedidoID 42, se obtuvo %d", res.PedidoID)
	}
}

func TestCheckoutController_FallaTransaccion(t *testing.T) {
	app := fiber.New()
	mockDB := &MockDatabaseRepository{
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
			return nil, errors.New("falla transaccional forzada para rollback")
		},
	}
	mockCache := &MockCacheRepository{}
	ctrl := NewCheckoutController(mockDB, mockCache)

	app.Post("/api/checkout", ctrl.Checkout)

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Test Fail",
		EmailCliente:  "fail@upse.edu.ec",
		MetodoPago:    "Efectivo contra entrega",
		Items: []models.CheckoutItem{
			{ProductoID: 99999, Cantidad: 1, Precio: 10.0},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición de prueba: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Se esperaba código 500 por rollback de la transacción, se obtuvo %d", resp.StatusCode)
	}
}
