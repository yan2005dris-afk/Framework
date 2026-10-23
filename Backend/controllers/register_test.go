package controllers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
)

type MockAuthRepo struct {
	MockDatabaseRepository
	RegistrarFunc func(req models.RegisterRequest, pwd string) (*models.Usuario, error)
}

func (m *MockAuthRepo) RegistrarUsuario(req models.RegisterRequest, pwd string) (*models.Usuario, error) {
	if m.RegistrarFunc != nil {
		return m.RegistrarFunc(req, pwd)
	}
	return &models.Usuario{
		ID:    50,
		Email: req.Email,
		Rol:   "cliente",
	}, nil
}

func TestRegister_CamposObligatorios(t *testing.T) {
	app := fiber.New()
	repo := &MockAuthRepo{}
	ctrl := NewAuthController("http://127.0.0.1:1", repo, "dev")
	app.Post("/api/register", ctrl.Register)

	body := []byte(`{"nombre": "", "email": "nuevo@test.com", "password": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("esperado 400 por campos vacios, recibido %d", resp.StatusCode)
	}
}

func TestRegister_Exitoso(t *testing.T) {
	app := fiber.New()
	repo := &MockAuthRepo{}
	ctrl := NewAuthController("http://127.0.0.1:1", repo, "dev")
	app.Post("/api/register", ctrl.Register)

	body := []byte(`{"nombre": "Carlos Perez", "email": "carlos@test.com", "password": "securepassword", "ref": "LUIS"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("esperado 201 Created, recibido %d", resp.StatusCode)
	}
}

func TestRegister_EmailDuplicado(t *testing.T) {
	app := fiber.New()
	repo := &MockAuthRepo{
		RegistrarFunc: func(req models.RegisterRequest, pwd string) (*models.Usuario, error) {
			return nil, errors.New("email ya registrado")
		},
	}
	ctrl := NewAuthController("http://127.0.0.1:1", repo, "dev")
	app.Post("/api/register", ctrl.Register)

	body := []byte(`{"nombre": "Carlos", "email": "duplicado@test.com", "password": "password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error en app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("esperado 409 Conflict, recibido %d", resp.StatusCode)
	}
}
