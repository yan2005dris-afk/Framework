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
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv:         getEnv("APP_ENV", "dev"),
		JWTSecret:      []byte(getEnv("AUTH_JWT_SECRET", "dev-insecure-secret-please-replace-in-production")),
		JWTExpiresSecs: 7200,

		DBHost:     getEnv("AUTH_DB_HOST", "localhost"),
		DBPort:     getEnv("AUTH_DB_PORT", "5436"),
		DBUser:     getEnv("AUTH_DB_USER", "postgres"),
		DBPassword: getEnv("AUTH_DB_PASSWORD", "postgres"),
		DBName:     getEnv("AUTH_DB_NAME", "multicatalogo_auth"),
	}
}

// DSN para pgxpool (autenticacion por password).
func (c *Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
