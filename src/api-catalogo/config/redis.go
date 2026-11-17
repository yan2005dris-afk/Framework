// Package config gestiona la configuración de la aplicación y la conexión a Redis.
package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient es el cliente global de Redis.
var RedisClient *redis.Client

// ConnectRedis inicializa el cliente de Redis y valida la conexión.
func ConnectRedis() (*redis.Client, error) {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("no se pudo conectar a Redis: %w", err)
	}

	log.Printf("✅ Conexión exitosa a Redis (%s:%s)", host, port)
	return RedisClient, nil
}

// CloseRedis cierra ordenadamente la conexión a Redis.
func CloseRedis() {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			log.Printf("⚠️ Error al cerrar conexión a Redis: %v", err)
		} else {
			log.Println("🔌 Conexión a Redis cerrada")
		}
	}
}
