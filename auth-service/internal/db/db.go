// Package db envuelve el pool de conexiones Postgres del auth-service.
// Convencion: creacion lazy (no abrimos conexion en New, sino en el
// primer Ping). Permite que los tests no necesiten una BD real.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"auth-service/internal/config"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(cfg *config.Config) *DB {
	return &DB{}
}

// Init crea el pool. Retorna error si la BD no responde.
func (d *DB) Init(cfg *config.Config) error {
	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		return fmt.Errorf("pgxpool.New: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	d.Pool = pool
	return nil
}

// Ping es lo que usa el healthcheck. Timeout corto para no bloquear
// al gateway / load balancer.
func (d *DB) Ping(ctx context.Context) error {
	if d.Pool == nil {
		return fmt.Errorf("auth db pool not initialised")
	}
	return d.Pool.Ping(ctx)
}
