package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

type MockPedidoRepo struct {
	MockDatabaseRepository
	ListarMiosFunc       func(usuarioID int, limit, offset int) ([]models.Pedido, error)
	ListarAdminFunc      func(estado string, limit, offset int) ([]models.Pedido, error)
	ObtenerDetalleFunc   func(id int) (*models.PedidoConDetalle, error)
	ActualizarEstadoFunc func(id int, estado string) error
}

func (m *MockPedidoRepo) ListarPedidosMios(usuarioID int, limit, offset int) ([]models.Pedido, error) {
	if m.ListarMiosFunc != nil {
		return m.ListarMiosFunc(usuarioID, limit, offset)
	}
	return []models.Pedido{
		{ID: 10, UsuarioID: usuarioID, Total: 120.0, Estado: "completado"},
	}, nil
}

func (m *MockPedidoRepo) ListarPedidosAdmin(estado string, limit, offset int) ([]models.Pedido, error) {
	if m.ListarAdminFunc != nil {
		return m.ListarAdminFunc(estado, limit, offset)
	}
	return []models.Pedido{
		{ID: 10, UsuarioID: 1, Total: 120.0, Estado: "completado"},
		{ID: 11, UsuarioID: 2, Total: 80.0, Estado: "pendiente"},
	}, nil
}

func (m *MockPedidoRepo) ObtenerPedidoDetalle(id int) (*models.PedidoConDetalle, error) {
	if m.ObtenerDetalleFunc != nil {
		return m.ObtenerDetalleFunc(id)
	}
	return &models.PedidoConDetalle{
		Pedido: models.Pedido{ID: id, UsuarioID: 5, Total: 99.0, Estado: "pendiente"},
		Items: []models.PedidoDetalleItem{
			{ID: 1, PedidoID: id, ProductoID: 2, NombreProducto: "Serum", Cantidad: 1, Subtotal: 99.0},
		},
	}, nil
}

func (m *MockPedidoRepo) ActualizarEstadoPedido(id int, estado string) error {
	if m.ActualizarEstadoFunc != nil {
		return m.ActualizarEstadoFunc(id, estado)
	}
	return nil
}

func TestListarPedidosMios_Exitoso(t *testing.T) {
	app := fiber.New()
	repo := &MockPedidoRepo{}
	ctrl := NewPedidoController(repo, "dev")

	app.Get("/pedidos/mios", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalUsuarioID, 42)
		return ctrl.ListarPedidosMios(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/pedidos/mios?page=1&limit=10", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("esperado 200 OK, recibido %d", resp.StatusCode)
	}
}

func TestActualizarEstadoPedido_Admin(t *testing.T) {
	app := fiber.New()
	repo := &MockPedidoRepo{}
	ctrl := NewPedidoController(repo, "dev")

	app.Patch("/pedidos/:id/estado", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "admin")
		return ctrl.ActualizarEstado(c)
	})

	body := []byte(`{"estado": "cancelado"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pedidos/10/estado", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("esperado 200 OK, recibido %d", resp.StatusCode)
	}
}

func TestObtenerPedidoPorID_ForbiddenAjenos(t *testing.T) {
	app := fiber.New()
	repo := &MockPedidoRepo{}
	ctrl := NewPedidoController(repo, "dev")

	app.Get("/pedidos/:id", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalUsuarioID, 99) // Distinto de usuarioID=5 del mock
		c.Locals(middleware.LocalRol, "cliente")
		return ctrl.ObtenerPedidoPorID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/pedidos/10", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("esperado 403 Forbidden para usuario ajeno, recibido %d", resp.StatusCode)
	}
}
