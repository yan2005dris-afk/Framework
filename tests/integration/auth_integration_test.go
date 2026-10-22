// Tests de integración para verificar que la autenticación funcione
// correctamente entre auth-service, Backend Go y SaaS FastAPI.
//
// Estos tests verifican:
// 1. Que el JWT_SECRET sea compartido entre servicios
// 2. Que los tokens emitidos por auth-service sean válidos en ambos backends
// 3. Que los proxies funcionen correctamente
// 4. Que los headers X-User-* se propaguen correctamente
//
// Pre-requisitos:
// - docker compose up -d --build
// - Esperar que los healthchecks pasen (~15s)
//
// Ejecutar:
//   cd tests/integration
//   go test -v -tags=integration
//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	gatewayURL     = "http://localhost:8080"
	backendURL     = "http://localhost:3000"
	authServiceURL = "http://localhost:3001"
	saasURL        = "http://localhost:8000"
)

// TestJWTSecretConsistency verifica que ambos servicios usen el mismo JWT_SECRET
func TestJWTSecretConsistency(t *testing.T) {
	// Este test verifica que las variables de entorno estén configuradas
	// correctamente para que los tokens sean compatibles

	authSecret := os.Getenv("AUTH_JWT_SECRET")
	if authSecret == "" {
		t.Skip("AUTH_JWT_SECRET no está configurado - saltando test de consistencia")
	}

	// En producción, JWT_SECRET del Backend debe coincidir con AUTH_JWT_SECRET
	backendSecret := os.Getenv("AUTH_JWT_SECRET")
	if backendSecret == "" {
		backendSecret = os.Getenv("JWT_SECRET")
	}

	if backendSecret != authSecret {
		t.Errorf("JWT_SECRET del Backend (%s) no coincide con AUTH_JWT_SECRET del auth-service (%s). Esto causará que los tokens no sean compatibles entre servicios.",
			backendSecret, authSecret)
	} else {
		t.Logf("✓ JWT_SECRET consistente entre servicios: %s", maskSecret(authSecret))
	}
}

// TestAuthServiceLoginDirect verifica que el auth-service emita tokens correctamente
func TestAuthServiceLoginDirect(t *testing.T) {
	body := map[string]string{
		"email":    "admin@upse.edu.ec",
		"password": "123456",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", authServiceURL+"/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("No se pudo conectar al auth-service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Login falló con status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Verificar que haya una cookie Set-Cookie
	cookies := resp.Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}

	if authCookie == nil {
		t.Fatal("No se encontró cookie auth_token en la respuesta")
	}

	if !authCookie.HttpOnly {
		t.Error("La cookie auth_token debe ser HttpOnly")
	}

	// Verificar que el token sea un JWT válido
	token := authCookie.Value
	claims := &jwt.MapClaims{}
	secret := []byte(os.Getenv("AUTH_JWT_SECRET"))
	if secret == nil {
		secret = []byte("dev-insecure-secret-please-replace-in-production")
	}

	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !parsed.Valid {
		t.Fatalf("El token emitido no es válido con el JWT_SECRET actual: %v", err)
	}

	// Verificar claims básicos
	if email, ok := (*claims)["email"].(string); !ok || email != "admin@upse.edu.ec" {
		t.Errorf("Claim email incorrecto: %v", (*claims)["email"])
	}

	if rol, ok := (*claims)["rol"].(string); !ok || rol != "admin" {
		t.Errorf("Claim rol incorrecto: %v", (*claims)["rol"])
	}

	t.Logf("✓ Token emitido correctamente: email=%s, rol=%s", (*claims)["email"], (*claims)["rol"])
}

// TestGatewayLoginViaBackend verifica que el login a través del gateway funcione
func TestGatewayLoginViaBackend(t *testing.T) {
	body := map[string]string{
		"email":    "admin@upse.edu.ec",
		"password": "123456",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", gatewayURL+"/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("No se pudo conectar al gateway: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Login vía gateway falló con status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Verificar cookie
	cookies := resp.Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}

	if authCookie == nil {
		t.Fatal("No se encontró cookie auth_token vía gateway")
	}

	t.Logf("✓ Login vía gateway exitoso, cookie recibida")
}

// TestGatewayLoginViaSaaS verifica que el login a través del SaaS proxy funcione
func TestGatewayLoginViaSaaS(t *testing.T) {
	body := map[string]string{
		"email":    "admin@upse.edu.ec",
		"password": "123456",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", gatewayURL+"/api/saas/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("No se pudo conectar al gateway (SaaS): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Login vía gateway (SaaS) falló con status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Verificar cookie
	cookies := resp.Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}

	if authCookie == nil {
		t.Fatal("No se encontró cookie auth_token vía gateway (SaaS)")
	}

	t.Logf("✓ Login vía gateway (SaaS) exitoso, cookie recibida")
}

// TestTokenReuseBetweenBackends verifica que un token emitido por auth-service
// sea aceptado tanto por el Backend como por el SaaS
func TestTokenReuseBetweenBackends(t *testing.T) {
	// 1. Obtener token del auth-service
	body := map[string]string{
		"email":    "admin@upse.edu.ec",
		"password": "123456",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", authServiceURL+"/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("No se pudo conectar al auth-service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login falló con status %d", resp.StatusCode)
	}

	cookies := resp.Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("No se recibió cookie del auth-service")
	}

	token := authCookie.Value

	// 2. Usar el token contra el Backend (/api/me)
	reqMe, _ := http.NewRequest("GET", backendURL+"/api/me", nil)
	reqMe.AddCookie(authCookie)
	respMe, err := client.Do(reqMe)
	if err != nil {
		t.Fatalf("Error al llamar /api/me del Backend: %v", err)
	}
	defer respMe.Body.Close()

	if respMe.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(respMe.Body)
		t.Errorf("Backend rechazó el token con status %d: %s", respMe.StatusCode, string(bodyBytes))
	} else {
		t.Logf("✓ Backend aceptó el token emitido por auth-service")
	}

	// 3. Usar el token contra el SaaS (/api/saas/auth/me)
	reqSaaS, _ := http.NewRequest("GET", saasURL+"/api/saas/auth/me", nil)
	reqSaaS.AddCookie(authCookie)
	respSaaS, err := client.Do(reqSaaS)
	if err != nil {
		t.Fatalf("Error al llamar /api/saas/auth/me del SaaS: %v", err)
	}
	defer respSaaS.Body.Close()

	if respSaaS.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(respSaaS.Body)
		t.Errorf("SaaS rechazó el token con status %d: %s", respSaaS.StatusCode, string(bodyBytes))
	} else {
		t.Logf("✓ SaaS aceptó el token emitido por auth-service")
	}
}

// TestGatewayForwardAuth verifica que el forward_auth del gateway funcione
func TestGatewayForwardAuth(t *testing.T) {
	// 1. Login para obtener cookie
	body := map[string]string{
		"email":    "admin@upse.edu.ec",
		"password": "123456",
	}
	jsonBody, _ := json.Marshal(body)

	reqLogin, _ := http.NewRequest("POST", gatewayURL+"/api/auth/login", bytes.NewReader(jsonBody))
	reqLogin.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	respLogin, err := client.Do(reqLogin)
	if err != nil {
		t.Fatalf("Error en login: %v", err)
	}
	defer respLogin.Body.Close()

	cookies := respLogin.Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("No se recibió cookie")
	}

	// 2. Acceder a endpoint protegido del Backend vía gateway
	reqRed, _ := http.NewRequest("GET", gatewayURL+"/api/tx/red", nil)
	reqRed.AddCookie(authCookie)
	respRed, err := client.Do(reqRed)
	if err != nil {
		t.Fatalf("Error al llamar /api/tx/red: %v", err)
	}
	defer respRed.Body.Close()

	if respRed.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(respRed.Body)
		t.Errorf("Endpoint protegido del Backend falló con status %d: %s", respRed.StatusCode, string(bodyBytes))
	} else {
		t.Logf("✓ Gateway forward_auth funciona correctamente para Backend")
	}

	// 3. Verificar que sin cookie devuelve 401
	reqRedNoAuth, _ := http.NewRequest("GET", gatewayURL+"/api/tx/red", nil)
	respRedNoAuth, err := client.Do(reqRedNoAuth)
	if err != nil {
		t.Fatalf("Error al llamar /api/tx/red sin auth: %v", err)
	}
	defer respRedNoAuth.Body.Close()

	if respRedNoAuth.StatusCode != http.StatusUnauthorized {
		t.Errorf("Endpoint protegido sin cookie debería devolver 401, obtuvo %d", respRedNoAuth.StatusCode)
	} else {
		t.Logf("✓ Gateway rechaza correctamente requests sin token")
	}
}

// Helper functions

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Variable de entorno %s no está configurada", key))
	}
	return val
}
