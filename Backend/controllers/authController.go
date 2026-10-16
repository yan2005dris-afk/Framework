// Package controllers implementa los controladores de la aplicación con inyección de dependencias.
package controllers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"multicatalogo-backend/auth"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// CookieName es el nombre de la cookie httpOnly donde viaja el JWT.
// Vivir en una constante evita typos entre emisor (Login) y verificador (middleware).
const CookieName = "auth_token"

// AuthController maneja la autenticación con JWT real (HS256) en cookie httpOnly.
type AuthController struct {
	dbRepo         repository.DatabaseRepository
	jwtSecret      []byte
	expiresSeconds int64
	appEnv         string
}

// NewAuthController instancia un AuthController con dependencias inyectadas.
// jwtSecret se usa para firmar los tokens (mínimo 32 bytes recomendado en producción).
// appEnv controla el flag Secure de la cookie (en dev http://localhost requiere Secure=false).
func NewAuthController(dbRepo repository.DatabaseRepository, jwtSecret string, appEnv string) *AuthController {
	if len(jwtSecret) == 0 {
		jwtSecret = "dev-insecure-secret-please-replace-in-production"
	}
	if appEnv == "" {
		appEnv = "dev"
	}
	return &AuthController{
		dbRepo:         dbRepo,
		jwtSecret:      []byte(jwtSecret),
		expiresSeconds: 7200,
		appEnv:         appEnv,
	}
}

// NewAuthControllerWithTTL permite especificar el TTL del token (segundos).
func NewAuthControllerWithTTL(dbRepo repository.DatabaseRepository, jwtSecret string, appEnv string, expiresSeconds int64) *AuthController {
	c := NewAuthController(dbRepo, jwtSecret, appEnv)
	if expiresSeconds > 0 {
		c.expiresSeconds = expiresSeconds
	}
	return c
}

// Login autentica al usuario con bcrypt, firma el JWT y lo entrega en una
// cookie httpOnly `auth_token` (SameSite=Lax, Secure según APP_ENV). El body
// JSON no incluye el token — eso reduce la superficie de robo por XSS al
// evitar que el JS del frontend tenga acceso al valor.
//
// Respuestas:
//   - 400 BadRequest: cuerpo inválido o email/password vacíos.
//   - 401 Unauthorized: usuario no encontrado o password incorrecta.
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cuerpo de petición inválido",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email y password son obligatorios",
		})
	}

	user, err := ac.dbRepo.BuscarUsuarioPorEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Credenciales incorrectas",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error interno del servidor",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Mismo mensaje que ErrNotFound para no filtrar existencia de cuentas.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Credenciales incorrectas",
		})
	}

	now := time.Now()
	expAt := now.Add(time.Duration(ac.expiresSeconds) * time.Second)
	claims := auth.AuthClaims{
		Email: user.Email,
		Rol:   user.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   itoa(user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(ac.jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo generar el token",
		})
	}

	// Cookie httpOnly: el JS del frontend NO puede leer este valor.
	// SameSite=Lax permite que el browser la envíe en requests same-site (POST /api/checkout).
	// Secure=true en cualquier entorno != dev (HTTPS obligatorio para cookie Secure).
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,
		Value:    signed,
		HTTPOnly: true,
		Secure:   ac.appEnv != "dev",
		SameSite: "lax",
		Path:     "/",
		Expires:  expAt,
	})

	return c.JSON(fiber.Map{
		"email":   user.Email,
		"rol":     user.Rol,
		"expires": ac.expiresSeconds,
	})
}

// Logout limpia la cookie httpOnly. Endpoint público (no necesita RequireAuth):
// si no hay cookie, simplemente termina; si la hay, la sobrescribe con MaxAge=-1
// para que el browser la borre.
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   ac.appEnv != "dev",
		SameSite: "lax",
		Path:     "/",
		MaxAge:   -1,
	})
	return c.JSON(fiber.Map{"status": "ok"})
}

// Me devuelve la identidad del usuario autenticado extrayéndola de c.Locals
// (poblada por middleware.RequireAuth). Requiere estar protegido por el
// middleware; si no se setearon locals, devuelve 401.
//
// Es el endpoint que el frontend usa al montar la app para hidratar la sesión
// desde la cookie, sin depender de nada guardado en localStorage.
func (ac *AuthController) Me(c *fiber.Ctx) error {
	email, emailOk := c.Locals(middleware.LocalEmail).(string)
	rol, _ := c.Locals(middleware.LocalRol).(string)
	usuarioID, _ := c.Locals(middleware.LocalUsuarioID).(int)
	if !emailOk || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No autenticado",
		})
	}
	return c.JSON(fiber.Map{
		"id":    usuarioID,
		"email": email,
		"rol":   rol,
	})
}

// itoa evita importar strconv solo para una conversión ID -> string en el claim Subject.
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
