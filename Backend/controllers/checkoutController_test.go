package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

// MockDatabaseRepository simula las operaciones de base de datos y captura lo que llega al repo.
type MockDatabaseRepository struct {
	mu sync.Mutex

	CrearPedidoCalls          []CrearPedidoCall
	CrearPedidoFunc            func(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error)
	ActualizarComisionesFunc  func(usuarioID int, totalPedido float64) error
}

// CrearPedidoCall registra los argumentos con que se invocó CrearPedido, para que
// los tests puedan afirmar que el controller pasó los valores esperados.
type CrearPedidoCall struct {
	UsuarioID     int
	EmailRecibido string
}

func (m *MockDatabaseRepository) BuscarUsuarioPorEmail(email string) (*models.Usuario, error) {
	return nil, nil
}
func (m *MockDatabaseRepository) ObtenerProductos() ([]models.Producto, error)        { return nil, nil }
func (m *MockDatabaseRepository) ObtenerProductoPorID(int) (*models.Producto, error)    { return nil, nil }
func (m *MockDatabaseRepository) ObtenerRed() (*models.Referido, error)                 { return nil, nil }
func (m *MockDatabaseRepository) CrearPedido(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
	m.mu.Lock()
	m.CrearPedidoCalls = append(m.CrearPedidoCalls, CrearPedidoCall{
		UsuarioID:     usuarioID,
		EmailRecibido: req.EmailCliente,
	})
	m.mu.Unlock()
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
	delCalled  bool
	lastDelKey string
	delKeys    []string
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
	m.delKeys = append(m.delKeys, key)
	return nil
}

// withAuthLocals setea c.Locals como si RequireAuth hubiera corrido antes.
func withAuthLocals(usuarioID int, email string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalUsuarioID, usuarioID)
		c.Locals(middleware.LocalEmail, email)
		c.Locals(middleware.LocalRol, "cliente")
		return c.Next()
	}
}

func newCheckoutApp(ctrl *CheckoutController, usuarioID int, email string) *fiber.App {
	app := fiber.New()
	// Simulamos que RequireAuth corrió antes.
	app.Use("/api", withAuthLocals(usuarioID, email))
	app.Post("/api/checkout", ctrl.Checkout)
	return app
}

func TestCheckoutController_Exitoso(t *testing.T) {
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
	ctrl := NewCheckoutController(mockDB, mockCache, "dev")
	app := newCheckoutApp(ctrl, 2, "cliente@upse.edu.ec")

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

	resp, err := app.Test(req, 5000)
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
	mockDB := &MockDatabaseRepository{
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
			return nil, errors.New("falla transaccional forzada para rollback")
		},
	}
	mockCache := &MockCacheRepository{}
	ctrl := NewCheckoutController(mockDB, mockCache, "dev")
	app := newCheckoutApp(ctrl, 2, "cliente@upse.edu.ec")

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

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error en petición de prueba: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Se esperaba código 500 por rollback de la transacción, se obtuvo %d", resp.StatusCode)
	}
}

// TestCheckoutController_EmailFalsoEnBody_SeSobrescribe verifica que aunque
// el body diga un email distinto al del token, el controller sobrescribe con
// el del token antes de llamar a CrearPedido. Así la fila en BD y las
// comisiones siempre corresponden al usuario autenticado.
func TestCheckoutController_EmailFalsoEnBody_SeSobrescribe(t *testing.T) {
	mockDB := &MockDatabaseRepository{
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
			return &models.Pedido{ID: 99, Total: 100, UsuarioID: usuarioID, EmailCliente: req.EmailCliente}, nil
		},
	}
	mockCache := &MockCacheRepository{}
	ctrl := NewCheckoutController(mockDB, mockCache, "dev")
	app := newCheckoutApp(ctrl, 7, "real@upse.edu.ec")

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Mallory",
		EmailCliente:  "atacante@example.com",
		MetodoPago:    "Tarjeta",
		Items: []models.CheckoutItem{
			{ProductoID: 1, Cantidad: 1, Precio: 100},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error al ejecutar petición: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperado 201, obtuvo %d", resp.StatusCode)
	}
	if len(mockDB.CrearPedidoCalls) != 1 {
		t.Fatalf("CrearPedido debió invocarse 1 vez, fue %d", len(mockDB.CrearPedidoCalls))
	}
	got := mockDB.CrearPedidoCalls[0]
	if got.UsuarioID != 7 {
		t.Errorf("usuarioID esperado 7 (del token), obtuvo %d", got.UsuarioID)
	}
	if got.EmailRecibido != "real@upse.edu.ec" {
		t.Errorf("email persistido esperado real@upse.edu.ec (del token), obtuvo %q", got.EmailRecibido)
	}
}

// TestCheckoutController_CantidadCero_Retorna400 verifica que un item con
// cantidad 0 no llega a la transacción; mapea ErrCantidadInvalida a 400.
func TestCheckoutController_CantidadCero_Retorna400(t *testing.T) {
	mockDB := &MockDatabaseRepository{}
	mockCache := &MockCacheRepository{}
	ctrl := NewCheckoutController(mockDB, mockCache, "dev")
	app := newCheckoutApp(ctrl, 2, "cliente@upse.edu.ec")

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Test",
		EmailCliente:  "cliente@upse.edu.ec",
		MetodoPago:    "Efectivo",
		Items: []models.CheckoutItem{
			{ProductoID: 1, Cantidad: 0, Precio: 50},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtuvo %d", resp.StatusCode)
	}
	if len(mockDB.CrearPedidoCalls) != 0 {
		t.Errorf("CrearPedido NO debió invocarse con cantidad 0 (se invocó %d veces)", len(mockDB.CrearPedidoCalls))
	}
}

// TestCheckoutController_CantidadNegativa_Retorna400 cubre el límite negativo.
func TestCheckoutController_CantidadNegativa_Retorna400(t *testing.T) {
	mockDB := &MockDatabaseRepository{}
	mockCache := &MockCacheRepository{}
	ctrl := NewCheckoutController(mockDB, mockCache, "dev")
	app := newCheckoutApp(ctrl, 2, "cliente@upse.edu.ec")

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Test",
		EmailCliente:  "cliente@upse.edu.ec",
		MetodoPago:    "Efectivo",
		Items: []models.CheckoutItem{
			{ProductoID: 1, Cantidad: -1, Precio: 50},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtuvo %d", resp.StatusCode)
	}
}

// TestCheckoutController_ProductosYRedCacheInvalidada verifica que el checkout
// invalida 'productos:all', 'red:tree' y la clave por email tras una compra exitosa.
func TestCheckoutController_ProductosYRedCacheInvalidada(t *testing.T) {
	mockDB := &MockDatabaseRepository{
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
			return &models.Pedido{ID: 100, Total: 50, UsuarioID: usuarioID}, nil
		},
	}
	mockCache := &MockCacheRepository{}
	ctrl := NewCheckoutController(mockDB, mockCache, "dev")
	app := newCheckoutApp(ctrl, 2, "cliente@upse.edu.ec")

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Buyer",
		EmailCliente:  "body@upse.edu.ec",
		MetodoPago:    "Tarjeta",
		Items: []models.CheckoutItem{
			{ProductoID: 1, Cantidad: 1, Precio: 50},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperado 201, obtuvo %d", resp.StatusCode)
	}

	// La goroutine puede tardar unos ms. Damos margen consultando luego.
	for i := 0; i < 50 && len(mockCache.delKeys) < 3; i++ {
		// spin corto
		for j := 0; j < 100000; j++ {
		}
	}
	have := map[string]bool{}
	for _, k := range mockCache.delKeys {
		have[k] = true
	}
	for _, want := range []string{"productos:all", "red:tree", "red:tree:cliente@upse.edu.ec"} {
		if !have[want] {
			t.Errorf("se esperaba Del(%q) tras checkout, no se encontró. delKeys=%v", want, mockCache.delKeys)
		}
	}
}
