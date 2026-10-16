// Package auth contiene los tipos compartidos entre los controllers y el middleware
// de autenticación. Evita ciclos de import.
//
// AuthClaims es la estructura de claims que el backend firma con HS256 y que el
// middleware RequireAuth parsea. Vive aquí para que tanto controllers (firmador)
// como middleware (verificador) puedan usarla sin dependencias circulares.
package auth

import "github.com/golang-jwt/jwt/v5"

// AuthClaims representa los claims del JWT firmado por el backend.
//   - Subject: UsuarioID (string porque jwt usa string para sub).
//   - Email y Rol: para que RequireAuth setee c.Locals sin volver a la BD.
type AuthClaims struct {
	Email string `json:"email"`
	Rol   string `json:"rol"`
	jwt.RegisteredClaims
}
