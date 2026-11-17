package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

type MockProdRepo struct {
	MockDatabaseRepository
	CrearProductoFunc      func(req models.ProductoRequest) (*models.Producto, error)
	ActualizarProductoFunc func(id int, req models.ProductoRequest) (*models.Producto, error)
	EliminarProductoFunc   func(id int) error
	ObtenerAdminFunc       func(string) ([]models.Producto, error)
}

func (m *MockProdRepo) CrearProducto(req models.ProductoRequest) (*models.Producto, error) {
	if m.CrearProductoFunc != nil {
		return m.CrearProductoFunc(req)
	}
	return &models.Producto{
		ID:          99,
		Nombre:      req.Nombre,
		Descripcion: req.Descripcion,
		Precio:      req.Precio,
		Categoria:   req.Categoria,
		Activo:      true,
		Stock:       100,
	}, nil
}

func (m *MockProdRepo) ActualizarProducto(id int, req models.ProductoRequest) (*models.Producto, error) {
	if m.ActualizarProductoFunc != nil {
		return m.ActualizarProductoFunc(id, req)
	}
	return &models.Producto{
		ID:          id,
		Nombre:      req.Nombre,
		Descripcion: req.Descripcion,
		Precio:      req.Precio,
		Categoria:   req.Categoria,
		Activo:      true,
		Stock:       100,
	}, nil
}

func (m *MockProdRepo) EliminarProductoLogico(id int) error {
	if m.EliminarProductoFunc != nil {
		return m.EliminarProductoFunc(id)
	}
	return nil
}

func (m *MockProdRepo) ObtenerProductosAdmin(q string) ([]models.Producto, error) {
	if m.ObtenerAdminFunc != nil {
		return m.ObtenerAdminFunc(q)
	}
	return []models.Producto{
		{ID: 1, Nombre: "Prod 1", Activo: true},
		{ID: 2, Nombre: "Prod 2 Inactivo", Activo: false},
	}, nil
}

func TestCrearProducto_RequiereAdmin(t *testing.T) {
	app := fiber.New()
	repo := &MockProdRepo{}
	ctrl := NewProductoController(repo, nil, "dev")

	app.Post("/productos", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "cliente")
		return ctrl.CrearProducto(c)
	})

	body := []byte(`{"nombre": "Crema", "precio": 25.5, "categoria": "Cuidado"}`)
	req := httptest.NewRequest(http.MethodPost, "/productos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("esperado 403 para cliente, recibido %d", resp.StatusCode)
	}
}

func TestCrearProducto_ExitosoAdmin(t *testing.T) {
	app := fiber.New()
	repo := &MockProdRepo{}
	cache := &MockCacheRepository{}
	ctrl := NewProductoController(repo, cache, "dev")

	app.Post("/productos", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "admin")
		return ctrl.CrearProducto(c)
	})

	body := []byte(`{"nombre": "Serum Facial", "precio": 45.0, "categoria": "Rostro"}`)
	req := httptest.NewRequest(http.MethodPost, "/productos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("esperado 201 Created, recibido %d", resp.StatusCode)
	}

	var p models.Producto
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("error al decodificar: %v", err)
	}

	if p.Nombre != "Serum Facial" || p.ID != 99 {
		t.Errorf("producto devuelto no coincide: %+v", p)
	}

	if !cache.delCalled || cache.lastDelKey != "productos:all" {
		t.Errorf("debe invalidar caché 'productos:all' tras crear")
	}
}

func TestEliminarProducto_Logico(t *testing.T) {
	app := fiber.New()
	repo := &MockProdRepo{}
	cache := &MockCacheRepository{}
	ctrl := NewProductoController(repo, cache, "dev")

	app.Delete("/productos/:id", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "admin")
		return ctrl.EliminarProducto(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/productos/99", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("esperado 200 OK, recibido %d", resp.StatusCode)
	}

	if !cache.delCalled || cache.lastDelKey != "productos:all" {
		t.Errorf("debe invalidar caché tras eliminación lógica")
	}
}

func TestGetProductosAdmin_ConFiltroBusqueda(t *testing.T) {
	app := fiber.New()
	busquedaRecibida := ""
	repo := &MockProdRepo{
		ObtenerAdminFunc: func(q string) ([]models.Producto, error) {
			busquedaRecibida = q
			return []models.Producto{
				{ID: 1, Nombre: "Serum Rejuvenecedor", Activo: true},
			}, nil
		},
	}
	ctrl := NewProductoController(repo, nil, "dev")

	app.Get("/admin/productos", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "admin")
		return ctrl.GetProductosAdmin(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/productos?q=Serum", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("esperado 200 OK, recibido %d", resp.StatusCode)
	}

	if busquedaRecibida != "Serum" {
		t.Errorf("esperado parametro de busqueda 'Serum', recibido '%s'", busquedaRecibida)
	}
}
