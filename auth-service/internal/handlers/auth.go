// Package handlers - auth.go expone los handlers HTTP de /auth/*.
//
// Endpoints (todos publicos, el gateway hace forward_auth para los que
// requieren token):
//
//	POST /auth/login    - body {email, password} -> 200 {token, ...} + cookie
//	POST /auth/logout   - limpia cookies
//	POST /auth/refresh  - emite un JWT nuevo a partir de uno vigente
//	POST /auth/validate - lee token (cookie o header), setea headers, 200/401
//	GET  /auth/me       - identidad del token (espejo de /api/me legacy)
package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/security"
)

// CookieName es el nombre del cookie httpOnly donde vive el JWT.
// Mismo nombre que Backend/cookieName (compatibilidad cross-service).
const CookieName = "auth_token"

// AuthHandler concentra las dependencias de los handlers /auth/*.
type AuthHandler struct {
	Cfg   *config.Config
	DB    *pgxpool.Pool
	Users *db.Repo
}

// NewAuthHandler crea un handler con sus deps.
func NewAuthHandler(cfg *config.Config, pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		Cfg:   cfg,
		DB:    pool,
		Users: &db.Repo{Pool: pool},
	}
}

// Login autentica con bcrypt y emite un JWT firmado.
//
// Cookies: setea 'auth_token' httpOnly (B2C). Si el email corresponde
// a un SaasUser vinculado, tambien setea 'saas_token' para futuras
// llamadas SaaS. Por ahora ambos derivan del mismo usuario (mismo
// sub) y la SaaS session es placeholder.
//
// Respuestas:
//   - 400 si body invalido o email/password vacios.
//   - 401 si usuario no existe o password incorrecta.
//   - 200 con {token, email, rol, expires} + cookies.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cuerpo de peticion invalido"})
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email y password son obligatorios"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, err := h.Users.BuscarUsuarioPorEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Credenciales incorrectas"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error interno del servidor"})
	}

	ok, err := security.VerifyPassword(user.Password, req.Password)
	if err != nil || !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Credenciales incorrectas"})
	}

	ttl := time.Duration(h.Cfg.JWTExpiresSecs) * time.Second
	token, err := security.SignB2CToken(h.Cfg.JWTSecret, user.ID, user.Email, user.Rol, ttl)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No se pudo firmar el token"})
	}

	secure := h.Cfg.AppEnv != "dev"
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,
		Value:    token,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "lax",
		Path:     "/",
		Expires:  time.Now().Add(ttl),
	})

	return c.JSON(fiber.Map{
		"token":   token,
		"email":   user.Email,
		"rol":     user.Rol,
		"expires": int(h.Cfg.JWTExpiresSecs),
	})
}

// Logout limpia la cookie. Idempotente.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   h.Cfg.AppEnv != "dev",
		SameSite: "lax",
		Path:     "/",
		MaxAge:   -1,
	})
	return c.JSON(fiber.Map{"status": "ok"})
}

// Validate lee el token (cookie o Authorization: Bearer) y si es valido:
//   - Setea headers X-User-Id, X-User-Email, X-User-Rol, X-Tenant-Id, X-Rol-Saas
//     (esto es lo que el gateway lee via copy_headers para pasar a
//     los backends upstream).
//   - Responde 200.
//
// Si invalido/ausente, responde 401.
func (h *AuthHandler) Validate(c *fiber.Ctx) error {
	token := extractToken(c)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token ausente"})
	}

	claims, err := security.ParseToken(h.Cfg.JWTSecret, token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token invalido o vencido"})
	}

	// Headers que el gateway va a copiar al upstream.
	c.Set("X-User-Id", claims.Subject)
	c.Set("X-User-Email", claims.Email)
	c.Set("X-User-Rol", claims.Rol)
	c.Set("X-Tenant-Id", itoa(claims.TenantID))
	c.Set("X-Rol-Saas", claims.RolSaas)

	return c.JSON(fiber.Map{"status": "ok"})
}

// Me devuelve la identidad del token (espejo del legacy /api/me).
// Header o cookie; si no, 401.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	token := extractToken(c)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token ausente"})
	}
	claims, err := security.ParseToken(h.Cfg.JWTSecret, token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token invalido o vencido"})
	}
	return c.JSON(fiber.Map{
		"sub":       claims.Subject,
		"email":     claims.Email,
		"rol":       claims.Rol,
		"tenant_id": claims.TenantID,
		"rol_saas":  claims.RolSaas,
	})
}

// Refresh emite un JWT nuevo a partir del actual (mismo secret, mismo
// TTL completo). Pensado para que el gateway o el cliente pidan uno
// fresco antes de que el actual venza.
//
// Por simplicidad, NO rotamos el secret y solo extendemos la expiracion;
// en una iteracion futura podemos agregar una denylist de tokens viejos.
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	token := extractToken(c)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token ausente"})
	}
	claims, err := security.ParseToken(h.Cfg.JWTSecret, token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token invalido o vencido"})
	}

	ttl := time.Duration(h.Cfg.JWTExpiresSecs) * time.Second
	signed, err := security.SignB2CToken(
		h.Cfg.JWTSecret,
		atoi(claims.Subject),
		claims.Email,
		claims.Rol,
		ttl,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No se pudo firmar el token"})
	}

	secure := h.Cfg.AppEnv != "dev"
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,
		Value:    signed,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "lax",
		Path:     "/",
		Expires:  time.Now().Add(ttl),
	})

	return c.JSON(fiber.Map{
		"token":   signed,
		"email":   claims.Email,
		"rol":     claims.Rol,
		"expires": int(h.Cfg.JWTExpiresSecs),
	})
}

// extractToken lee el token priorizando header Authorization: Bearer,
// cayendo a la cookie 'auth_token' (mismo patron que el middleware del Backend).
func extractToken(c *fiber.Ctx) string {
	header := c.Get(fiber.HeaderAuthorization)
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}
	return strings.TrimSpace(c.Cookies(CookieName))
}

// itoa convert int to string. Es la misma que en security.go (no la
// importamos para no aadir un modulo solo por esto).
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}

// atoi convierte string a int. Tolera strings con o sin signo.
func atoi(s string) int {
	if s == "" || s == "0" {
		return 0
	}
	n := 0
	neg := false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		return -n
	}
	return n
}
