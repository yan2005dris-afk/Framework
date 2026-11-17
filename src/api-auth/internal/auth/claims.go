// Package auth contiene los tipos compartidos entre los handlers y el
// codigo de firma/verificacion de JWT. Vive aqui (en lugar del package
// 'auth' del Backend) porque el backend ya no firma JWTs; el auth-service
// es el unico responsable de emision y validacion.
//
// AuthClaims es la estructura de claims del JWT firmado con HS256. Los
// claims `sub`, `email`, `rol`, `tenant_id` y `rol_saas` se setean al firmar;
// el verificador los lee y los pone en headers HTTP que el gateway pasa
// a los backends upstream.
package auth

import "github.com/golang-jwt/jwt/v5"

// AuthClaims es la estructura de claims compartida por los dos tipos
// de tokens que emite el auth-service.
//
//   - sub       : usuarioID (string porque JWT RegisteredClaims lo pide asi).
//   - email     : identidad del usuario.
//   - rol       : rol B2C ('admin' o 'cliente') si el token es B2C; vacio si es SaaS.
//   - tenant_id : id del tenant SaaS si el token es SaaS; 0 si es B2C.
//   - rol_saas  : rol SaaS ('superadmin', 'admin', 'member') si el token es SaaS.
//
// El resto viene de jwt.RegisteredClaims (iat, exp, etc).
type AuthClaims struct {
	Email    string `json:"email"`
	Rol      string `json:"rol,omitempty"`
	TenantID int    `json:"tenant_id,omitempty"`
	RolSaas  string `json:"rol_saas,omitempty"`
	jwt.RegisteredClaims
}
