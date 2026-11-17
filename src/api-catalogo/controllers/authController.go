// Package controllers implementa los controladores de la aplicación con inyección de dependencias.
package controllers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"golang.org/x/crypto/bcrypt"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"

	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
)

// CookieName se mantiene por compat con tests/AuthController legacy. El valor
// real de la cookie lo emite el auth-service; este controller solo lo reenvia.
const CookieName = "auth_token"

// AuthController es un proxy HTTP transparente al auth-service.
//
// Post-GA5 el Backend ya no firma JWTs ni compara bcrypts: lo unico que
// hacia con esos secretos (los emite y los valida) vive ahora en
// auth-service:3001. Mantenemos los endpoints locales /api/login,
// /api/logout, /api/refresh, /api/me en el Backend por dos razones:
//  1. Compatibilidad: el Frontend (y sus tests) sigue llamando /api/login.
//     Cambiar la URL romeria el SPA.
//  2. Defense in depth: cuando el gateway funciona correctamente, este
//     proxy no se usa porque el gateway enruta /api/auth/* directo al
//     auth-service. Pero si por alguna razon un request llega al Backend
//     sin pasar por el gateway, este proxy es una segunda capa que sigue
//     autenticando contra el auth-service.
//
// El middleware.RequireAuth del Backend sigue corriendo como tercera
// capa: si llegan requests con cookie auth_token valida, valida local
// (defensa) sin necesidad de round-trip al auth-service en cada request.
type AuthController struct {
	authServiceURL string
	httpClient     *http.Client
	dbRepo         repository.DatabaseRepository
	appEnv         string
}

// NewAuthController instancia el proxy.
// authServiceURL: URL base del auth-service (e.g. "http://auth-service:3001").
func NewAuthController(authServiceURL string, dbRepo repository.DatabaseRepository, appEnv string) *AuthController {
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:3001"
	}
	if appEnv == "" {
		appEnv = "dev"
	}
	return &AuthController{
		authServiceURL: strings.TrimRight(authServiceURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		dbRepo: dbRepo,
		appEnv: appEnv,
	}
}

// Login hace proxy a POST /auth/login del auth-service.
// Forwarda el body tal cual + cookie/Authorization, y devuelve el response
// (status + headers, incluido Set-Cookie) al cliente.
func (ac *AuthController) Login(c *fiber.Ctx) error {
	return ac.proxy(c, http.MethodPost, "/auth/login")
}

// Logout hace proxy a POST /auth/logout.
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	return ac.proxy(c, http.MethodPost, "/auth/logout")
}

// Me devuelve la identidad del request actual desde c.Locals (poblada por
// middleware.RequireAuth). Esto evita round-trip al auth-service cuando
// la sesion ya fue validada por el middleware local (defensa en profundidad).
// Si los locals no estan poblados (caso patologico en que alguien llama
// /api/me sin RequireAuth antes), devuelve 401.
func (ac *AuthController) Me(c *fiber.Ctx) error {
	email, emailOk := c.Locals(middleware.LocalEmail).(string)
	rol, _ := c.Locals(middleware.LocalRol).(string)
	usuarioID, _ := c.Locals(middleware.LocalUsuarioID).(int)
	if !emailOk || email == "" {
		return models.APIError{Status: fiber.StatusUnauthorized, Error: "No autenticado"}.Write(c)
	}
	return c.JSON(fiber.Map{
		"id":    usuarioID,
		"email": email,
		"rol":   rol,
	})
}

// proxy reenvia el request al auth-service y devuelve el response tal cual.
//
// Headers que forwarda:
//   - Content-Type
//   - Authorization (token JWT directo, caso poco comun)
//   - Cookie (caso normal: cookie httpOnly del browser)
//
// Headers de respuesta: copia todos los del upstream (incluyendo Set-Cookie).
func (ac *AuthController) proxy(c *fiber.Ctx, method, path string) error {
	target := ac.authServiceURL + path

	req, err := http.NewRequest(method, target, bytes.NewReader(c.Body()))
	if err != nil {
		return models.APIError{
			Status:  fiber.StatusInternalServerError,
			Error:   "No se pudo construir el request al auth-service",
			Detalle: err.Error(),
		}.Write(c)
	}

	if ct := c.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth := c.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if cookie := c.Get("Cookie"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return models.APIError{
			Status:  fiber.StatusBadGateway,
			Error:   "auth-service no responde",
			Detalle: err.Error(),
		}.Write(c)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.APIError{
			Status:  fiber.StatusBadGateway,
			Error:   "No se pudo leer respuesta del auth-service",
			Detalle: err.Error(),
		}.Write(c)
	}

	// Copia headers de respuesta, en particular Set-Cookie.
	for k, vs := range resp.Header {
		for _, v := range vs {
			c.Set(k, v)
		}
	}

	return c.Status(resp.StatusCode).Send(body)
}

// itoa evita importar strconv en sitios donde solo necesitamos Subject string.
// Ya esta en middleware/auth.go y aqui para no acoplar paquetes.
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

// FormatDuration se usa en algunos puntos de logs. Mantenido porque su signature
// aparece en tests; lo dejo aunque hoy no se use directamente.
var _ = fmt.Sprintf

// Register registra un nuevo usuario en auth.users y le asigna su nodo en referidos.
func (ac *AuthController) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return models.APIError{Status: fiber.StatusBadRequest, Error: "Cuerpo de petición inválido"}.Write(c)
	}

	if req.Nombre == "" || req.Email == "" || req.Password == "" {
		return models.APIError{
			Status:  fiber.StatusBadRequest,
			Error:   "Validación fallida",
			Detalle: "Los campos nombre, email y password son obligatorios",
		}.Write(c)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.APIError{Status: fiber.StatusInternalServerError, Error: "Error al procesar contraseña"}.Write(c)
	}

	usuario, err := ac.dbRepo.RegistrarUsuario(req, string(hashedPassword))
	if err != nil {
		if strings.Contains(err.Error(), "email ya registrado") {
			return models.APIError{Status: fiber.StatusConflict, Error: "El correo electrónico ya se encuentra registrado"}.Write(c)
		}
		if strings.Contains(err.Error(), "código de referido inválido") {
			return models.APIError{Status: fiber.StatusBadRequest, Error: "El código de referido especificado no existe"}.Write(c)
		}
		return models.APIError{Status: fiber.StatusInternalServerError, Error: err.Error()}.Write(c)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"mensaje": "Usuario registrado exitosamente",
		"usuario": usuario,
	})
}
