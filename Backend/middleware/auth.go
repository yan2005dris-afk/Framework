// Package middleware contiene los middlewares HTTP de la aplicación.
// Las capas (RequireAuth, RequireAdmin) cuelgan de net/http vía fiber.
package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/repository"
)

// Local keys usados para propagar la identidad verificada en c.Locals.
const (
	LocalEmail     = "email"
	LocalRol       = "rol"
	LocalUsuarioID = "usuarioID"
)

// AuthConfig agrupa las dependencias que RequireAuth necesita para identificar al usuario.
type AuthConfig struct {
	JWTSecret []byte
	DBRepo    repository.DatabaseRepository // opcional, solo si se quiere hidratar el usuarioID desde BD
}

// RequireAuth devuelve un fiber.Handler que verifica el header Authorization: Bearer <jwt>,
// parsea el token HS256 con el secreto provisto y setea c.Locals con la identidad.
//
// Si el header falta o el token es inválido/vence, responde 401 con cuerpo JSON.
// Si DBRepo está seteado y el email del token no se encuentra en BD, también 401.
func RequireAuth(cfg AuthConfig) fiber.Handler {
	if len(cfg.JWTSecret) == 0 {
		// No bloquear el wiring si se pasa vacío: se rechaza cada request.
		cfg.JWTSecret = []byte("__unset__")
	}
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			return unauthorized(c, "Token ausente")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return unauthorized(c, "Esquema de autorización inválido")
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims := &controllers.AuthClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de firma inesperado")
			}
			return cfg.JWTSecret, nil
		})
		if err != nil {
			return unauthorized(c, "Token inválido o vencido")
		}

		c.Locals(LocalEmail, claims.Email)
		c.Locals(LocalRol, claims.Rol)

		if cfg.DBRepo != nil {
			user, err := cfg.DBRepo.BuscarUsuarioPorEmail(claims.Email)
			if err != nil {
				return unauthorized(c, "Usuario del token no existe")
			}
			c.Locals(LocalUsuarioID, user.ID)
		}

		return c.Next()
	}
}

// RequireAdmin exige rol=admin además de un token válido.
// Útil para endpoints administrativos; RequireAuth debe correr antes.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rol, _ := c.Locals(LocalRol).(string)
		if rol != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Se requiere rol administrador",
			})
		}
		return c.Next()
	}
}

func unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": msg,
	})
}
