package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

// MockDatabaseRepository simula las operaciones de base de datos y captura lo que llega al repo.
type MockDatabaseRepository struct {
	mu sync.Mutex

	CrearPedidoCalls         []CrearPedidoCall
	CrearPedidoFunc           func(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error)
	ActualizarComisionesFunc func(usuarioID int, totalPedido float64) error
}

// CrearPedidoCall registra los argumentos con que se invocó CrearPedido, para que
// los tests puedan afirmar que el controller pasó los valores esperados.
type CrearPedidoCall struct {
	UsuarioID      int
	EmailRecibido  string
	IdempotencyKey string
}

func (m *MockDatabaseRepository) BuscarUsuarioPorEmail(email string) (*models.Usuario, error) {
	return nil, nil
}
func (m *MockDatabaseRepository) ObtenerProductos() ([]models.Producto, error)        { return nil, nil }
func (m *MockDatabaseRepository) ObtenerProductoPorID(int) (*models.Producto, error)    { return nil, nil }
func (m *MockDatabaseRepository) ObtenerProductosAdmin() ([]models.Producto, error) { return nil, nil }
func (m *MockDatabaseRepository) CrearProducto(req models.ProductoRequest) (*models.Producto, error) { return nil, nil }
func (m *MockDatabaseRepository) ActualizarProducto(int, models.ProductoRequest) (*models.Producto, error) { return nil, nil }
func (m *MockDatabaseRepository) EliminarProductoLogico(int) error { return nil }
func (m *MockDatabaseRepository) ObtenerRed() (*models.Referido, error)                 { return nil, nil }
func (m *MockDatabaseRepository) CrearPedido(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error) {
	m.mu.Lock()
	m.CrearPedidoCalls = append(m.CrearPedidoCalls, CrearPedidoCall{
		UsuarioID:      usuarioID,
		EmailRecibido:  req.EmailCliente,
		IdempotencyKey: idempotencyKey,
	})
	m.mu.Unlock()
	if m.CrearPedidoFunc != nil {
		return m.CrearPedidoFunc(usuarioID, req, idempotencyKey)
	}
	return &models.Pedido{ID: 101, Total: 100}, false, nil
}
func (m *MockDatabaseRepository) ActualizarComisiones(usuarioID int, totalPedido float64) error {
	if m.ActualizarComisionesFunc != nil {
		return m.ActualizarComisionesFunc(usuarioID, totalPedido)
	}
	return nil
}
func (m *MockDatabaseRepository) RegistrarComisionesPedido(int, int, float64) ([]models.Comision, error) { return nil, nil }
func (m *MockDatabaseRepository) ObtenerComisionesPorReferido(int) (*models.ResumenComisiones, error) { return nil, nil }
func (m *MockDatabaseRepository) LiquidarComisiones(int) (*models.Liquidacion, error) { return nil, nil }
func (m *MockDatabaseRepository) ListarLiquidaciones() ([]models.Liquidacion, error) { return nil, nil }
func (m *MockDatabaseRepository) ListarLiquidacionesPorReferido(int) ([]models.Liquidacion, error) { return nil, nil }
func (m *MockDatabaseRepository) ListarPedidosMios(int, int, int) ([]models.Pedido, error) { return nil, nil }
func (m *MockDatabaseRepository) ListarPedidosAdmin(string, int, int) ([]models.Pedido, error) { return nil, nil }
func (m *MockDatabaseRepository) ObtenerPedidoDetalle(int) (*models.PedidoConDetalle, error) { return nil, nil }
func (m *MockDatabaseRepository) ActualizarEstadoPedido(int, string) error { return nil }

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
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error) {
			return &models.Pedido{
				ID:            42,
				UsuarioID:     usuarioID,
				NombreCliente: req.NombreCliente,
				EmailCliente:  req.EmailCliente,
				Total:         75.5,
				Estado:        "completado",
			}, false, nil
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
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error) {
			return nil, false, errors.New("falla transaccional forzada para rollback")
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
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error) {
			return &models.Pedido{ID: 99, Total: 100, UsuarioID: usuarioID, EmailCliente: req.EmailCliente}, false, nil
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
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error) {
			return &models.Pedido{ID: 100, Total: 50, UsuarioID: usuarioID}, false, nil
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

// ---------------------------------------------------------------------------
// U4-T3 · S10: idempotencia del checkout
//
// El service worker nunca reintenta solo un POST, pero el cliente sí. Si el
// primer intento llegó al servidor y solo se perdió la respuesta, el reintento
// crearía un segundo pedido. Estos tests fijan el contrato que lo evita.
// ---------------------------------------------------------------------------

// El header `Idempotency-Key` tiene que llegar al repositorio tal cual: es la
// clave con la que el repo decide entre crear el pedido y hacer replay.
func TestCheckoutController_ReenviaLaIdempotencyKeyAlRepositorio(t *testing.T) {
	mockDB := &MockDatabaseRepository{}
	ctrl := NewCheckoutController(mockDB, &MockCacheRepository{}, "dev")
	app := newCheckoutApp(ctrl, 3, "clave@upse.edu.ec")

	const key = "9f2c1a7e-5b31-4c88-9a0d-6e1f2b3c4d5e"
	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Estudiante UPSE",
		EmailCliente:  "estudiante@upse.edu.ec",
		MetodoPago:    "Tarjeta de crédito",
		Items:         []models.CheckoutItem{{ProductoID: 1, Cantidad: 1, Precio: 45.0}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error al ejecutar petición de prueba: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Se esperaba 201 con la key nueva, se obtuvo %d", resp.StatusCode)
	}
	if len(mockDB.CrearPedidoCalls) != 1 {
		t.Fatalf("Se esperaba 1 llamada a CrearPedido, hubo %d", len(mockDB.CrearPedidoCalls))
	}
	if got := mockDB.CrearPedidoCalls[0].IdempotencyKey; got != key {
		t.Errorf("La key que llegó al repo = %q, se esperaba %q", got, key)
	}
}

// Sin el header, la key va vacía y el checkout se comporta como siempre. Es lo
// que evita que un cliente viejo, o uno que no es una PWA, se rompa al desplegar.
func TestCheckoutController_SinHeaderLaKeyVaVacia(t *testing.T) {
	mockDB := &MockDatabaseRepository{}
	ctrl := NewCheckoutController(mockDB, &MockCacheRepository{}, "dev")
	app := newCheckoutApp(ctrl, 3, "sinho@upse.edu.ec")

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Estudiante UPSE",
		EmailCliente:  "sinho@upse.edu.ec",
		MetodoPago:    "Tarjeta de crédito",
		Items:         []models.CheckoutItem{{ProductoID: 1, Cantidad: 1, Precio: 45.0}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error al ejecutar petición de prueba: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Se esperaba 201 sin header, se obtuvo %d", resp.StatusCode)
	}
	if got := mockDB.CrearPedidoCalls[0].IdempotencyKey; got != "" {
		t.Errorf("Sin header la key debe ir vacía, llegó %q", got)
	}
}

// Un replay (la key ya se usó, el repo devuelve el pedido original) tiene que:
//   - responder 200 en vez de 201, para que el cliente no lo tome por un pedido nuevo;
//   - NO volver a pedir las comisiones. Si las pidiera, el reintento duplicaría
//     también las ventas de la red, que es el mismo bug un paso más abajo y
//     bastante más difícil de detectar que un pedido repetido.
func TestCheckoutController_ReplayResponde200YCobraUnaSolaComision(t *testing.T) {
	// Canal con buffer: si el controller no llamara, la goroutine no se
	// cuelga esperando a que alguien lea.
	comisiones := make(chan float64, 1)

	mockDB := &MockDatabaseRepository{
		CrearPedidoFunc: func(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error) {
			return &models.Pedido{ID: 42, Total: 75.5, UsuarioID: usuarioID}, true, nil
		},
		ActualizarComisionesFunc: func(_ int, totalPedido float64) error {
			comisiones <- totalPedido
			return nil
		},
	}
	ctrl := NewCheckoutController(mockDB, &MockCacheRepository{}, "dev")
	app := newCheckoutApp(ctrl, 3, "replay@upse.edu.ec")

	body, _ := json.Marshal(models.CheckoutRequest{
		NombreCliente: "Estudiante UPSE",
		EmailCliente:  "replay@upse.edu.ec",
		MetodoPago:    "Tarjeta de crédito",
		Items:         []models.CheckoutItem{{ProductoID: 1, Cantidad: 1, Precio: 75.5}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "9f2c1a7e-5b31-4c88-9a0d-6e1f2b3c4d5e")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Error al ejecutar petición de prueba: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Un replay debe responder 200, se obtuvo %d", resp.StatusCode)
	}

	var respuesta models.CheckoutResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		t.Fatalf("No se pudo leer la respuesta: %v", err)
	}
	if respuesta.PedidoID != 42 {
		t.Errorf("El replay debe devolver el pedido original (42), devolvió %d", respuesta.PedidoID)
	}

	// El cobro de comisiones pasa por una goroutine. Le damos una ventana chica:
	// si en 100ms no apareció, es que el controller no lo llamó.
	select {
	case total := <-comisiones:
		t.Errorf("Un replay no debe volver a cobrar comisiones, cobró %v", total)
	case <-time.After(100 * time.Millisecond):
	}
}
