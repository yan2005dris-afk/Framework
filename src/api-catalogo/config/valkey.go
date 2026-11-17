// Package config gestiona la configuración de la aplicación y la conexión Singleton a Valkey/Redis.
package config

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	valkeyClient *redis.Client
	valkeyOnce   sync.Once
	valkeyErr    error
)

// GetValkeyClient devuelve la instancia Singleton thread-safe del cliente Valkey.
func GetValkeyClient() (*redis.Client, error) {
	valkeyOnce.Do(func() {
		host := getEnv("VALKEY_HOST", getEnv("REDIS_HOST", "localhost"))
		port := getEnv("VALKEY_PORT", getEnv("REDIS_PORT", "6379"))
		password := getEnv("VALKEY_PASSWORD", getEnv("REDIS_PASSWORD", ""))
		db := 0

		valkeyClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: password,
			DB:       db,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := valkeyClient.Ping(ctx).Err(); err != nil {
			valkeyErr = fmt.Errorf("no se pudo conectar a Valkey (%s:%s): %w", host, port, err)
			log.Printf("⚠️ %v", valkeyErr)
			return
		}

		log.Printf("✅ Conexión Singleton exitosa a Valkey (%s:%s)", host, port)
	})

	return valkeyClient, valkeyErr
}

// CloseValkey cierra ordenadamente la conexión Singleton de Valkey.
func CloseValkey() {
	if valkeyClient != nil {
		if err := valkeyClient.Close(); err != nil {
			log.Printf("⚠️ Error al cerrar conexión a Valkey: %v", err)
		} else {
			log.Println("🔌 Conexión Singleton a Valkey cerrada")
		}
	}
}
