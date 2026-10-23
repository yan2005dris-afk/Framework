package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"multicatalogo-backend/auth"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// stubDBRepo implementa repository.DatabaseRepository devolviendo el usuario
// sólo cuando el email coincide con `knownEmail`.
type stubDBRepo struct {
	knownEmail string
	knownID    int
}

func (s *stubDBRepo) RegistrarUsuario(req models.RegisterRequest, pwd string) (*models.Usuario, error) { return nil, nil }
func (s *stubDBRepo) BuscarUsuarioPorEmail(email string) (*models.Usuario, error) {
	if email == s.knownEmail {
		return &models.Usuario{ID: s.knownID, Email: email, Rol: "admin"}, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "no")
}

// Resto de la interfaz: devolvemos errores neutros porque el middleware no
// los invoca en estos tests.
func (s *stubDBRepo) ObtenerProductos() ([]models.Producto, error)              { return nil, nil }
func (s *stubDBRepo) ObtenerProductoPorID(int) (*models.Producto, error)       { return nil, nil }
func (s *stubDBRepo) ObtenerProductosAdmin() ([]models.Producto, error) { return nil, nil }
func (s *stubDBRepo) CrearProducto(req models.ProductoRequest) (*models.Producto, error) { return nil, nil }
func (s *stubDBRepo) ActualizarProducto(int, models.ProductoRequest) (*models.Producto, error) { return nil, nil }
func (s *stubDBRepo) EliminarProductoLogico(int) error { return nil }
func (s *stubDBRepo) ObtenerRed() (*models.Referido, error)                     { return nil, nil }
func (s *stubDBRepo) CrearPedido(int, models.CheckoutRequest, string) (*models.Pedido, bool, error) {
	return nil, false, nil
}
func (s *stubDBRepo) ActualizarComisiones(int, float64) error { return nil }
func (s *stubDBRepo) ListarPedidosMios(int, int, int) ([]models.Pedido, error) { return nil, nil }
func (s *stubDBRepo) ListarPedidosAdmin(string, int, int) ([]models.Pedido, error) { return nil, nil }
func (s *stubDBRepo) ObtenerPedidoDetalle(int) (*models.PedidoConDetalle, error) { return nil, nil }
func (s *stubDBRepo) ActualizarEstadoPedido(int, string) error { return nil }
func (s *stubDBRepo) ObtenerReporteResumen(string, string) (*models.ReporteResumen, error) { return nil, nil }
func (s *stubDBRepo) ObtenerTopProductos(int) ([]models.TopProducto, error) { return nil, nil }
func (s *stubDBRepo) ObtenerTopReferidos(int) ([]models.TopReferido, error) { return nil, nil }
func (s *stubDBRepo) ObtenerComisionesPendientes() ([]models.DeudaDistribuidor, error) { return nil, nil }
func (s *stubDBRepo) RegistrarComisionesPedido(int, int, float64) ([]models.Comision, error) { return nil, nil }
func (s *stubDBRepo) ObtenerComisionesPorReferido(int) (*models.ResumenComisiones, error) { return nil, nil }
func (s *stubDBRepo) LiquidarComisiones(int) (*models.Liquidacion, error) { return nil, nil }
func (s *stubDBRepo) ListarLiquidaciones() ([]models.Liquidacion, error) { return nil, nil }
func (s *stubDBRepo) ListarLiquidacionesPorReferido(int) ([]models.Liquidacion, error) { return nil, nil }

func signToken(t *testing.T, secret []byte, email, rol string, expiresAt time.Time) string {
	t.Helper()
	claims := auth.AuthClaims{
		Email: email,
		Rol:   rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("error firmando token: %v", err)
	}
	return signed
}

func newAppProtected(dbRepo repository.DatabaseRepository) (*fiber.App, *int) {
	called := 0
	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTSecret: []byte("test-secret"), DBRepo: dbRepo}))
	app.Get("/protected", func(c *fiber.Ctx) error {
		called++
		email, _ := c.Locals(LocalEmail).(string)
		rol, _ := c.Locals(LocalRol).(string)
		uid, _ := c.Locals(LocalUsuarioID).(int)
		return c.JSON(fiber.Map{"email": email, "rol": rol, "uid": uid})
	})
	return app, &called
}

func TestRequireAuth_TokenValido_SeteaLocals(t *testing.T) {
	stub := &stubDBRepo{knownEmail: "admin@upse.edu.ec", knownID: 7}
	app, called := newAppProtected(stub)

	tok := signToken(t, []byte("test-secret"), "admin@upse.edu.ec", "admin", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test falló: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200, obtuvo %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var got map[string]interface{}
	_ = json.Unmarshal(body, &got)
	if got["email"] != "admin@upse.edu.ec" {
		t.Errorf("email en locals incorrecto: %v", got["email"])
	}
	if got["rol"] != "admin" {
		t.Errorf("rol en locals incorrecto: %v", got["rol"])
	}
	if got["uid"] == nil {
		t.Errorf("usuarioID en locals no fue seteado")
	}
	if *called != 1 {
		t.Errorf("next() debería haberse llamado 1 vez, fue %d", *called)
	}
}

func TestRequireAuth_SinHeader_Retorna401(t *testing.T) {
	app, called := newAppProtected(&stubDBRepo{})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
	if *called != 0 {
		t.Errorf("next() no debió invocarse")
	}
}

func TestRequireAuth_HeaderMalFormado_Retorna401(t *testing.T) {
	app, _ := newAppProtected(&stubDBRepo{})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NotBearer abc")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
}

func TestRequireAuth_TokenFirmaAdulterada_Retorna401(t *testing.T) {
	app, _ := newAppProtected(&stubDBRepo{})
	tok := signToken(t, []byte("OTHER-SECRET"), "admin@upse.edu.ec", "admin", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
}

func TestRequireAuth_TokenVencido_Retorna401(t *testing.T) {
	app, _ := newAppProtected(&stubDBRepo{})
	tok := signToken(t, []byte("test-secret"), "admin@upse.edu.ec", "admin", time.Now().Add(-1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuvo %d", resp.StatusCode)
	}
}

func TestRequireAdmin_ClienteEsRechazado(t *testing.T) {
	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTSecret: []byte("test-secret")}))
	app.Use(RequireAdmin())
	app.Get("/admin", func(c *fiber.Ctx) error { return c.SendString("ok") })

	tok := signToken(t, []byte("test-secret"), "cliente@upse.edu.ec", "cliente", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("esperado 403, obtuvo %d", resp.StatusCode)
	}
}

func TestRequireAdmin_AdminPasa(t *testing.T) {
	app := fiber.New()
	app.Use(RequireAuth(AuthConfig{JWTSecret: []byte("test-secret")}))
	app.Use(RequireAdmin())
	app.Get("/admin", func(c *fiber.Ctx) error { return c.SendString("ok") })

	tok := signToken(t, []byte("test-secret"), "admin@upse.edu.ec", "admin", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("esperado 200, obtuvo %d - body=%s", resp.StatusCode, string(body))
	}
}

// TestRequireAuth_DesdeCookie_Valida verifica que el middleware lee el JWT
// desde la cookie httpOnly cuando no se envia header Authorization.
// Esto es el camino normal del flujo SPA: el browser adjunta la cookie
// automaticamente y el backend la acepta sin que el JS vea nunca el token.
func TestRequireAuth_DesdeCookie_Valida(t *testing.T) {
	stub := &stubDBRepo{knownEmail: "admin@upse.edu.ec", knownID: 7}
	app, called := newAppProtected(stub)

	tok := signToken(t, []byte("test-secret"), "admin@upse.edu.ec", "admin", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// Sólo cookie, sin header Authorization.
	req.AddCookie(&http.Cookie{Name: CookieName, Value: tok})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test falló: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("esperado 200 desde cookie, obtuvo %d - body=%s", resp.StatusCode, string(body))
	}
	if *called != 1 {
		t.Errorf("next() debió invocarse 1 vez, fue %d", *called)
	}
}

// TestRequireAuth_HeaderGanaSobreCookie verifica que cuando hay ambos (header
// y cookie), el header Authorization tiene precedencia. El backend prefiere el
// header porque permite a clientes externos (curl, tests) usar Bearer explícito.
func TestRequireAuth_HeaderGanaSobreCookie(t *testing.T) {
	stub := &stubDBRepo{knownEmail: "admin@upse.edu.ec", knownID: 7}
	app, _ := newAppProtected(stub)

	// Header con un secret válido.
	tokHeader := signToken(t, []byte("test-secret"), "admin@upse.edu.ec", "admin", time.Now().Add(1*time.Hour))
	// Cookie con un secret distinto (debe ser ignorada por el middleware).
	tokCookie := signToken(t, []byte("OTHER-SECRET"), "attacker@example.com", "admin", time.Now().Add(1*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokHeader)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: tokCookie})

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("esperado 200, obtuvo %d - body=%s", resp.StatusCode, string(body))
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "admin@upse.edu.ec") {
		t.Errorf("se esperaba email del token del header; body=%s", string(body))
	}
	if strings.Contains(string(body), "attacker@example.com") {
		t.Errorf("el header no debió ser anulado por la cookie; body=%s", string(body))
	}
}


