// Package config gestiona la configuración de la aplicación y la conexión a la base de datos.
package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// DB es el pool global de conexiones a PostgreSQL gestionado por pgx.
var DB *pgxpool.Pool

// getEnv obtiene una variable de entorno o retorna un valor por defecto si no está definida.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// getEnvAny devuelve el valor de la primera variable de entorno definida de la
// lista, o el fallback si ninguna lo está. Sirve para aceptar tanto el nombre
// con prefijo de servicio (BACKEND_DB_HOST, el que publica docker-compose.yml y
// el que ya usa scripts/apply_migrations.sh) como el nombre corto historico
// (DB_HOST), sin romper a quien corra el binario fuera de compose.
func getEnvAny(fallback string, keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return fallback
}

// ConnectDB carga variables de entorno, configura el pool de pgx y valida la conexión.
func ConnectDB() (*pgxpool.Pool, error) {
	// Intentamos cargar el archivo .env si existe (sin fallar si las variables ya vienen del entorno)
	_ = godotenv.Load()

	// El compose define BACKEND_DB_* (mismo prefijo que las migraciones);
	// leer solo DB_HOST hacia que el binario cayera al default "localhost"
	// dentro del contenedor y el pool nunca conectara, dejando /api/health
	// en 503 y el gateway sin arrancar por depends_on: service_healthy.
	host := getEnvAny("localhost", "BACKEND_DB_HOST", "DB_HOST")
	port := getEnvAny("5432", "BACKEND_DB_PORT", "DB_PORT")
	user := getEnvAny("postgres", "BACKEND_DB_USER", "DB_USER")
	pass := getEnvAny("postgres", "BACKEND_DB_PASSWORD", "DB_PASSWORD")
	dbname := getEnvAny("multicatalogo", "BACKEND_DB_NAME", "DB_NAME")
	sslmode := getEnvAny("disable", "BACKEND_DB_SSLMODE", "DB_SSLMODE")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, dbname, sslmode)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error al parsear configuración de conexión a la BD: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error al crear el pool de conexiones: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}

	DB = pool
	log.Printf("✅ Conexión exitosa a PostgreSQL (%s:%s/%s)", host, port, dbname)
	return DB, nil
}

// CloseDB cierra ordenadamente el pool de conexiones.
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("🔌 Conexión a PostgreSQL cerrada")
	}
}
