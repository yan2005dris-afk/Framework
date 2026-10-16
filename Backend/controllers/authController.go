// Package controllers implementa los controladores de la aplicación con inyección de dependencias.
package controllers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"multicatalogo-backend/auth"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// AuthController maneja la autenticación de usuarios con JWT real (HS256) y verificación bcrypt.
type AuthController struct {
	dbRepo         repository.DatabaseRepository
	jwtSecret      []byte
	expiresSeconds int64
}

// NewAuthController instancia un AuthController con dependencias inyectadas.
// jwtSecret se usa para firmar los tokens (mínimo 32 bytes recomendado).
// expiresSeconds define la vigencia del token (default 7200s = 2h si es <= 0).
func NewAuthController(dbRepo repository.DatabaseRepository, jwtSecret string) *AuthController {
	if len(jwtSecret) == 0 {
		jwtSecret = "dev-insecure-secret-please-replace-in-production"
	}
	expires := int64(7200)
	return &AuthController{
		dbRepo:         dbRepo,
		jwtSecret:      []byte(jwtSecret),
		expiresSeconds: expires,
	}
}

// NewAuthControllerWithTTL instancia un AuthController con TTL de token explícito (segundos).
func NewAuthControllerWithTTL(dbRepo repository.DatabaseRepository, jwtSecret string, expiresSeconds int64) *AuthController {
	c := NewAuthController(dbRepo, jwtSecret)
	if expiresSeconds > 0 {
		c.expiresSeconds = expiresSeconds
	}
	return c
}

// Login autentica al usuario con bcrypt y devuelve un JWT firmado (HS256).
//
// Valida que el body sea {email, password} no vacíos, busca el usuario por
// email, compara el password contra el hash bcrypt almacenado y firma un
// token con los claims estándar {sub: usuarioID, email, rol, iat, exp}.
//
// Respuestas:
//   - 400 BadRequest: cuerpo inválido o email/password vacíos.
//   - 401 Unauthorized: usuario no encontrado o password incorrecto.
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
	claims := auth.AuthClaims{
		Email: user.Email,
		Rol:   user.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   itoa(user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ac.expiresSeconds) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(ac.jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo generar el token",
		})
	}

	return c.JSON(fiber.Map{
		"token":   signed,
		"email":   user.Email,
		"rol":     user.Rol,
		"expires": ac.expiresSeconds,
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
