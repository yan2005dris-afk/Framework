package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

// fakeAuthService es un httptest.Server que mockea el auth-service.
// Devuelve lo que configuramos en cada test, y graba las requests que recibe.
type fakeAuthService struct {
	server   *httptest.Server
	lastReq  *http.Request
	lastBody []byte
	// Respuesta programable:
	respStatus  int
	respBody    string
	respCookies []*http.Cookie
}

func newFakeAuthService(t *testing.T) *fakeAuthService {
	f := &fakeAuthService{respStatus: http.StatusOK, respBody: `{"ok":true}`}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.lastReq = r
		f.lastBody = body
		// Devolver la respuesta programable con las cookies programables.
		for _, c := range f.respCookies {
			http.SetCookie(w, c)
		}
		w.WriteHeader(f.respStatus)
		_, _ = w.Write([]byte(f.respBody))
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func TestAuthController_Login_ProxyaYDevuelveCookie(t *testing.T) {
	fake := newFakeAuthService(t)
	fake.respStatus = http.StatusOK
	fake.respBody = `{"email":"admin@upse.edu.ec","rol":"admin","expires":7200}`
	fake.respCookies = []*http.Cookie{{
		Name:     CookieName,
		Value:    "fake-jwt-signed",
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		MaxAge:   7200,
	}}

	ctrl := NewAuthController(fake.server.URL, nil, "dev")
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
		t.Fatalf("esperado 200, obtuvo %d", resp.StatusCode)
	}

	// 1. La cookie del auth-service llego al cliente.
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == CookieName {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatalf("no se encontró cookie %s en respuesta: %v", CookieName, resp.Cookies())
	}
	if cookie.Value != "fake-jwt-signed" {
		t.Errorf("cookie Value esperado 'fake-jwt-signed', obtuvo %q", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Errorf("cookie debe ser httpOnly")
	}

	// 2. El body que devolvió el auth-service fue forwardeado integro.
	respBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(respBody), `"email":"admin@upse.edu.ec"`) {
		t.Errorf("body esperado contiene email; obtuvo: %s", string(respBody))
	}

	// 3. El request llego al fake con el body original.
	if fake.lastReq == nil {
		t.Fatalf("auth-service no recibio request")
	}
	if string(fake.lastBody) != string(body) {
		t.Errorf("body original no fue forwardeado integro")
	}
	if ct := fake.lastReq.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type esperado 'application/json', obtuvo %q", ct)
	}
}

func TestAuthController_Login_400CuandoCuerpoFalta(t *testing.T) {
	fake := newFakeAuthService(t)
	ctrl := NewAuthController(fake.server.URL, nil, "dev")
	app := fiber.New()
	app.Post("/api/login", ctrl.Login)

	// Body vacio -> el fake recibe "{}" o vacio igual, y el auth-service
	// en produccion devolveria 400. Aqui el fake devuelve 200, asi que
	// el proxy termina 200. El test importante es que la cookie no se
	// propague cuando el upstream falla (próximo test).
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200 (forwarded del fake), obtuvo %d", resp.StatusCode)
	}
	for _, c := range resp.Cookies() {
		if c.Name == CookieName && c.Value != "" {
			t.Errorf("cookie de sesion no debe setearse si la respuesta es 200 sin Set-Cookie upstream")
		}
	}
}

func TestAuthController_Login_NoPropagaCookieSiUpstreamFalla(t *testing.T) {
	fake := newFakeAuthService(t)
	fake.respStatus = http.StatusUnauthorized
	fake.respBody = `{"error":"Credenciales incorrectas"}`
	fake.respCookies = nil
	ctrl := NewAuthController(fake.server.URL, nil, "dev")
	app := fiber.New()
	app.Post("/api/login", ctrl.Login)

	body, _ := json.Marshal(models.LoginRequest{Email: "admin@upse.edu.ec", Password: "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
	respBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(respBody), "Credenciales incorrectas") {
		t.Errorf("body esperado contiene mensaje del upstream, obtuvo: %s", string(respBody))
	}
	for _, c := range resp.Cookies() {
		if c.Name == CookieName {
			t.Errorf("no debe haber cookie de sesion cuando upstream devuelve 401")
		}
	}
}

func TestAuthController_Logout_ProxyaYDevuelveMaxAgeNegativo(t *testing.T) {
	fake := newFakeAuthService(t)
	fake.respStatus = http.StatusOK
	fake.respBody = `{"status":"ok"}`
	fake.respCookies = []*http.Cookie{{
		Name:   CookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}}

	ctrl := NewAuthController(fake.server.URL, nil, "dev")
	app := fiber.New()
	app.Post("/api/logout", ctrl.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200, obtuvo %d", resp.StatusCode)
	}
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("se esperaba Set-Cookie de limpieza")
	}
	for _, c := range cookies {
		if c.Name == CookieName && c.MaxAge >= 0 {
			t.Errorf("cookie de logout debe tener MaxAge<0, obtuvo %d", c.MaxAge)
		}
	}
}

func TestAuthController_Me_ConLocals(t *testing.T) {
	ctrl := NewAuthController("http://127.0.0.1:1", nil, "dev") // Nunca se llama en Me().
	app := fiber.New()
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

func TestAuthController_Me_SinLocals(t *testing.T) {
	ctrl := NewAuthController("http://127.0.0.1:1", nil, "dev")
	app := fiber.New()
	app.Get("/api/me", ctrl.Me)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
}

func TestAuthController_Login_502SiUpstreamNoResponde(t *testing.T) {
	// URL de un puerto que no escucha.
	ctrl := NewAuthController("http://127.0.0.1:1", nil, "dev")
	app := fiber.New()
	app.Post("/api/login", ctrl.Login)

	body, _ := json.Marshal(models.LoginRequest{Email: "x@y.z", Password: "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("esperado 502 si auth-service no responde, obtuvo %d", resp.StatusCode)
	}
}
