package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

// dbWithUser es un repo que devuelve un usuario con un hash bcrypt pre-generado.
type dbWithUser struct {
	email  string
	id     int
	hash   string
	notFound bool
}

func (d *dbWithUser) BuscarUsuarioPorEmail(email string) (*models.Usuario, error) {
	if d.notFound || email != d.email {
		return nil, fiber.NewError(fiber.StatusNotFound, "no encontrado")
	}
	return &models.Usuario{ID: d.id, Email: d.email, Password: d.hash, Rol: "admin"}, nil
}
func (d *dbWithUser) ObtenerProductos() ([]models.Producto, error)                       { return nil, nil }
func (d *dbWithUser) ObtenerProductoPorID(int) (*models.Producto, error)                  { return nil, nil }
func (d *dbWithUser) ObtenerRed() (*models.Referido, error)                               { return nil, nil }
func (d *dbWithUser) CrearPedido(int, models.CheckoutRequest) (*models.Pedido, error)    { return nil, nil }
func (d *dbWithUser) ActualizarComisiones(int, float64) error                             { return nil }

// hashPassword genera un hash bcrypt determinístico para tests (cost 4 para velocidad).
func hashPassword(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), 4)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(h)
}

// TestAuthController_Login_SetteaCookie verifica que /api/login configura
// la cookie httpOnly con el JWT firmado cuando las credenciales son válidas.
// Que el body JSON NO incluya el token tambien es parte del contrato.
func TestAuthController_Login_SetteaCookie(t *testing.T) {
	repo := &dbWithUser{email: "admin@upse.edu.ec", id: 42, hash: hashPassword(t, "123456")}
	ctrl := NewAuthController(repo, "test-secret", "dev")
	app := fiber.New()
	app.Post("/api/login", ctrl.Login)

	body, _ := json.Marshal(models.LoginRequest{Email: "admin@upse.edu.ec", Password: "123456"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("esperado 200, obtuvo %d - body=%s", resp.StatusCode, string(respBody))
	}

	// 1. La cookie debe estar presente y marcada httpOnly.
	cookies := resp.Cookies()
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CookieName {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatalf("no se encontró cookie %s en respuesta: %v", CookieName, cookies)
	}
	if !cookie.HttpOnly {
		t.Errorf("cookie %s debe ser httpOnly", CookieName)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite esperado Lax, obtuvo %v", cookie.SameSite)
	}
	if cookie.Secure {
		t.Errorf("en dev la cookie NO debe ser Secure (http://localhost), obtuvo true")
	}
	if cookie.Value == "" {
		t.Errorf("cookie no puede estar vacía")
	}
	if cookie.Path != "/" {
		t.Errorf("cookie Path esperado '/', obtuvo %q", cookie.Path)
	}

	// 2. El body no debe incluir el token.
	respBody, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(respBody), `"token"`) {
		t.Errorf("body NO debe contener 'token' (debe ir solo en cookie): %s", string(respBody))
	}
	if !strings.Contains(string(respBody), `"email":"admin@upse.edu.ec"`) {
		t.Errorf("body debe incluir email del usuario: %s", string(respBody))
	}
	if !strings.Contains(string(respBody), `"rol":"admin"`) {
		t.Errorf("body debe incluir rol: %s", string(respBody))
	}
}

// TestAuthController_Login_PasswordIncorrecta retorna 401 sin filtrar
// existencia de la cuenta.
func TestAuthController_Login_PasswordIncorrecta(t *testing.T) {
	repo := &dbWithUser{email: "admin@upse.edu.ec", id: 42, hash: hashPassword(t, "123456")}
	ctrl := NewAuthController(repo, "test-secret", "dev")
	app := fiber.New()
	app.Post("/api/login", ctrl.Login)

	body, _ := json.Marshal(models.LoginRequest{Email: "admin@upse.edu.ec", Password: "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
	for _, c := range resp.Cookies() {
		if c.Name == CookieName {
			t.Errorf("NO debe haber cookie de sesión cuando auth falla")
		}
	}
}

// TestAuthController_Logout_LimpiaCookie verifica que POST /api/logout setea
// la cookie con MaxAge=0 (o negativo) y body 'status':'ok'.
func TestAuthController_Logout_LimpiaCookie(t *testing.T) {
	repo := &dbWithUser{email: "admin@upse.edu.ec", id: 42, hash: hashPassword(t, "123456")}
	ctrl := NewAuthController(repo, "test-secret", "dev")
	app := fiber.New()
	app.Post("/api/logout", ctrl.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200, obtuvo %d", resp.StatusCode)
	}

	cookies := resp.Cookies()
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CookieName {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatalf("se esperaba Set-Cookie de limpieza")
	}
	if cookie.MaxAge >= 0 {
		t.Errorf("cookie de logout debe tener MaxAge<0, obtuvo %d", cookie.MaxAge)
	}
}

// TestAuthController_Me_ConLocals retorna 200 con la identidad del token.
func TestAuthController_Me_ConLocals(t *testing.T) {
	repo := &dbWithUser{email: "admin@upse.edu.ec", id: 42, hash: "x"}
	ctrl := NewAuthController(repo, "test-secret", "dev")
	app := fiber.New()
	// Simulamos RequireAuth corriendo antes.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalEmail, "admin@upse.edu.ec")
		c.Locals(middleware.LocalRol, "admin")
		c.Locals(middleware.LocalUsuarioID, 42)
		return c.Next()
	})
	app.Get("/api/me", ctrl.Me)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200, obtuvo %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"id":42`) {
		t.Errorf("body debe incluir id: %s", string(body))
	}
	if !strings.Contains(string(body), `"email":"admin@upse.edu.ec"`) {
		t.Errorf("body debe incluir email: %s", string(body))
	}
	if !strings.Contains(string(body), `"rol":"admin"`) {
		t.Errorf("body debe incluir rol: %s", string(body))
	}
}

// TestAuthController_Me_SinLocals retorna 401 (defensivo si alguien llama Me sin middleware).
func TestAuthController_Me_SinLocals(t *testing.T) {
	repo := &dbWithUser{}
	ctrl := NewAuthController(repo, "test-secret", "dev")
	app := fiber.New()
	app.Get("/api/me", ctrl.Me)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
}

// TestAuthController_Login_RepoSinConexion_NoFiltraDetalle verifica que errores
// 500 internos no se filtran al body en producción (APP_ENV=prod).
func TestAuthController_Login_RepoSinConexion_NoFiltraDetalle(t *testing.T) {
	repo := &dbWithBrokenLookup{}
	ctrl := NewAuthController(repo, "test-secret", "prod")
	app := fiber.New()
	app.Post("/api/login", ctrl.Login)

	body, _ := json.Marshal(models.LoginRequest{Email: "any@example.com", Password: "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("esperado 500, obtuvo %d", resp.StatusCode)
	}
	respBody, _ := io.ReadAll(resp.Body)
	// En prod no debe filtrarse el detalle del error.
	if strings.Contains(string(respBody), "connection refused") {
		t.Errorf("en prod el body NO debe filtrar detalles internos: %s", string(respBody))
	}
}

type dbWithBrokenLookup struct{}

func (d *dbWithBrokenLookup) BuscarUsuarioPorEmail(string) (*models.Usuario, error) {
	return nil, errors.New("connection refused: db down at /var/run/postgresql")
}
func (d *dbWithBrokenLookup) ObtenerProductos() ([]models.Producto, error)                     { return nil, nil }
func (d *dbWithBrokenLookup) ObtenerProductoPorID(int) (*models.Producto, error)                { return nil, nil }
func (d *dbWithBrokenLookup) ObtenerRed() (*models.Referido, error)                             { return nil, nil }
func (d *dbWithBrokenLookup) CrearPedido(int, models.CheckoutRequest) (*models.Pedido, error)  { return nil, nil }
func (d *dbWithBrokenLookup) ActualizarComisiones(int, float64) error                           { return nil }

// silence unused-time warnings
var _ = time.Now
