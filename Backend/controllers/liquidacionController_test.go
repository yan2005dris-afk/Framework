package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

type MockLiquidacionRepo struct {
	MockDatabaseRepository
	ObtenerComisionesFunc func(referidoID int) (*models.ResumenComisiones, error)
	LiquidarFunc          func(referidoID int) (*models.Liquidacion, error)
	ListarLiquidacionesFn func() ([]models.Liquidacion, error)
}

func (m *MockLiquidacionRepo) ObtenerComisionesPorReferido(id int) (*models.ResumenComisiones, error) {
	if m.ObtenerComisionesFunc != nil {
		return m.ObtenerComisionesFunc(id)
	}
	return &models.ResumenComisiones{
		ReferidoID:     id,
		TotalPendiente: 150.0,
		TotalPagado:    50.0,
		Comisiones:     []models.Comision{},
	}, nil
}

func (m *MockLiquidacionRepo) LiquidarComisiones(id int) (*models.Liquidacion, error) {
	if m.LiquidarFunc != nil {
		return m.LiquidarFunc(id)
	}
	return &models.Liquidacion{
		ID:         1,
		ReferidoID: id,
		Total:      150.0,
		Estado:     "completada",
	}, nil
}

func (m *MockLiquidacionRepo) ListarLiquidaciones() ([]models.Liquidacion, error) {
	if m.ListarLiquidacionesFn != nil {
		return m.ListarLiquidacionesFn()
	}
	return []models.Liquidacion{
		{ID: 1, ReferidoID: 2, Total: 150.0, Estado: "completada"},
	}, nil
}

func TestObtenerComisionesMias_SinAuth(t *testing.T) {
	app := fiber.New()
	repo := &MockLiquidacionRepo{}
	ctrl := NewLiquidacionController(repo, "dev")

	app.Get("/comisiones/mias", ctrl.ObtenerComisionesMias)

	req := httptest.NewRequest(http.MethodGet, "/comisiones/mias", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("esperado status 401, recibido %d", resp.StatusCode)
	}
}

func TestObtenerComisionesMias_Exitoso(t *testing.T) {
	app := fiber.New()
	repo := &MockLiquidacionRepo{}
	ctrl := NewLiquidacionController(repo, "dev")

	app.Get("/comisiones/mias", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalUsuarioID, 2)
		return ctrl.ObtenerComisionesMias(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/comisiones/mias", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("esperado status 200, recibido %d", resp.StatusCode)
	}

	var res models.ResumenComisiones
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("error al decodificar respuesta: %v", err)
	}

	if res.ReferidoID != 2 || res.TotalPendiente != 150.0 {
		t.Errorf("datos devueltos no coinciden con mock: %+v", res)
	}
}

func TestLiquidar_RequiereAdmin(t *testing.T) {
	app := fiber.New()
	repo := &MockLiquidacionRepo{}
	ctrl := NewLiquidacionController(repo, "dev")

	app.Post("/liquidaciones", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "cliente")
		return ctrl.Liquidar(c)
	})

	body := []byte(`{"referidoId": 2}`)
	req := httptest.NewRequest(http.MethodPost, "/liquidaciones", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("esperado status 403 para rol cliente, recibido %d", resp.StatusCode)
	}
}

func TestLiquidar_ExitosoAdmin(t *testing.T) {
	app := fiber.New()
	repo := &MockLiquidacionRepo{}
	ctrl := NewLiquidacionController(repo, "dev")

	app.Post("/liquidaciones", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "admin")
		return ctrl.Liquidar(c)
	})

	body := []byte(`{"referidoId": 2}`)
	req := httptest.NewRequest(http.MethodPost, "/liquidaciones", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("esperado status 201, recibido %d", resp.StatusCode)
	}
}

func TestLiquidar_SinSaldo(t *testing.T) {
	app := fiber.New()
	repo := &MockLiquidacionRepo{
		LiquidarFunc: func(id int) (*models.Liquidacion, error) {
			return nil, errors.New("no hay comisiones pendientes para liquidar")
		},
	}
	ctrl := NewLiquidacionController(repo, "dev")

	app.Post("/liquidaciones", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "admin")
		return ctrl.Liquidar(c)
	})

	body := []byte(`{"referidoId": 2}`)
	req := httptest.NewRequest(http.MethodPost, "/liquidaciones", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("esperado status 400 por saldo en 0, recibido %d", resp.StatusCode)
	}
}
