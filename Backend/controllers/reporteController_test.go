package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

type MockReporteRepo struct {
	MockDatabaseRepository
	ResumenFunc func(desde, hasta string) (*models.ReporteResumen, error)
	TopProdFunc func(limit int) ([]models.TopProducto, error)
}

func (m *MockReporteRepo) ObtenerReporteResumen(desde, hasta string) (*models.ReporteResumen, error) {
	if m.ResumenFunc != nil {
		return m.ResumenFunc(desde, hasta)
	}
	return &models.ReporteResumen{
		TotalVendido:         5420.50,
		NumeroPedidos:        45,
		TicketPromedio:       120.45,
		ComisionesDevengadas: 650.00,
	}, nil
}

func (m *MockReporteRepo) ObtenerTopProductos(limit int) ([]models.TopProducto, error) {
	if m.TopProdFunc != nil {
		return m.TopProdFunc(limit)
	}
	return []models.TopProducto{
		{ProductoID: 1, Nombre: "Serum Facial", TotalVentas: 2300.0, Cantidad: 50},
	}, nil
}

func TestReporteResumen_Admin(t *testing.T) {
	app := fiber.New()
	repo := &MockReporteRepo{}
	ctrl := NewReporteController(repo, "dev")

	app.Get("/reportes/resumen", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "admin")
		return ctrl.Resumen(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reportes/resumen?desde=2026-01-01&hasta=2026-12-31", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("esperado 200 OK, recibido %d", resp.StatusCode)
	}

	var res models.ReporteResumen
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("error decodificando json: %v", err)
	}

	if res.TotalVendido != 5420.50 || res.NumeroPedidos != 45 {
		t.Errorf("datos inesperados en reporte resumen: %+v", res)
	}
}

func TestReporteResumen_ForbiddenCliente(t *testing.T) {
	app := fiber.New()
	repo := &MockReporteRepo{}
	ctrl := NewReporteController(repo, "dev")

	app.Get("/reportes/resumen", func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalRol, "cliente")
		return ctrl.Resumen(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reportes/resumen", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("esperado 403 Forbidden para cliente, recibido %d", resp.StatusCode)
	}
}
