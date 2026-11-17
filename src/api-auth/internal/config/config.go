// Package config gestiona la configuracion del auth-service via
// variables de entorno (cargadas desde .env si existe).
//
// Convenciones:
//   - Prefijo AUTH_ para todo dato sensible (secrets, DSNs).
//   - jwt_secret: la misma clave del JWT_SECRET del Backend permite
//     tokens compatibles (mismo secret). Distinta clave permite
//     tokens separados (aqui por defecto).
package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	AppEnv         string
	JWTSecret      []byte
	JWTExpiresSecs int64

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSchema   string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv:         getEnv("APP_ENV", "dev"),
		JWTSecret:      []byte(getEnv("AUTH_JWT_SECRET", "dev-insecure-secret-please-replace-in-production")),
		JWTExpiresSecs: 7200,

		// BD unica compartida con Backend y SaaS. El auth-service escribe
		// solo en el schema 'auth'.
		DBHost:     getEnv("AUTH_DB_HOST", "localhost"),
		DBPort:     getEnv("AUTH_DB_PORT", "5432"),
		DBUser:     getEnv("AUTH_DB_USER", "postgres"),
		DBPassword: getEnv("AUTH_DB_PASSWORD", "postgres"),
		DBName:     getEnv("AUTH_DB_NAME", "multicatalogo"),
		// Schema que el auth-service posee y donde escribe. Los demas
		// schemas (backend, saas) son read-only para este servicio.
		DBSchema: getEnv("AUTH_DB_SCHEMA", "auth"),
	}
}

// DSN para pgxpool (autenticacion por password).
// El schema 'auth' se aplica por query (search_path) o por SET en cada
// sesion — ver internal/db/db.go.
func (c *Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

// Schema que el servicio posee y donde escribe. Usado para setear
// `search_path` por conexion o prefijar queries.
func (c *Config) Schema() string {
	return c.DBSchema
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
