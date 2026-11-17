// Package security concentra las primitivas criptograficas del auth-service:
// bcrypt para verificar passwords y HS256 para firmar/verificar JWTs.
package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"auth-service/internal/auth"
)

// BcryptCost es el cost factor para hash y verify. 10 es estandar (Backend).
const BcryptCost = 10

// HashPassword devuelve el hash bcrypt de plain. Si plain esta vacio, falla.
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("plain password vacio")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}
	return string(h), nil
}

// VerifyPassword compara plain contra hash. true si matchean, false si no.
// Cualquier error distinto a ErrMismatchedHashAndPassword se propaga.
func VerifyPassword(hash, plain string) (bool, error) {
	if hash == "" || plain == "" {
		return false, errors.New("hash o plain vacio")
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, fmt.Errorf("bcrypt.CompareHashAndPassword: %w", err)
}

// SignB2CToken firma un JWT B2C (claims {sub, email, rol}). HS256 con secret.
func SignB2CToken(secret []byte, userID int, email, rol string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := auth.AuthClaims{
		Email: email,
		Rol:   rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// SignSaaSToken firma un JWT SaaS (claims {sub, email, tenant_id, rol_saas}).
func SignSaaSToken(secret []byte, userID int, email, tenantSlug, rolSaas string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := auth.AuthClaims{
		Email:    email,
		TenantID: tenantIDFromSlug(tenantSlug),
		RolSaas:  rolSaas,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// ParseToken verifica la firma HS256 y devuelve los claims, o un error
// si la firma es invalida o el token esta vencido.
func ParseToken(secret []byte, signed string) (*auth.AuthClaims, error) {
	claims := &auth.AuthClaims{}
	_, err := jwt.ParseWithClaims(signed, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo de firma inesperado: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	return claims, nil
}

// itoa convierte int a string sin importar strconv solo para Subject.
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

// tenantIDFromSlug: placeholder. La implementacion real consultara
// auth.saas_tenants por slug para resolver el ID. Para GA2 dejamos el ID
// dummy 1 cuando el slug es 'acme' (el seed), y 0 para vacio (sin tenant).
//
// Cuando implementemos el SaaS user lookup, esta funcion abrira su propia
// query a la BD; por ahora es un stub que devuelve 1 fijo para que el
// flujo basico funcione end-to-end con un solo tenant demo.
func tenantIDFromSlug(slug string) int {
	if slug == "" {
		return 0
	}
	return 1
}
