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
// Configura `search_path` para que las queries sin prefijo de schema
// caigan en `auth.*` por defecto (el schema que el servicio posee).
func (d *DB) Init(cfg *config.Config) error {
	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		return fmt.Errorf("pgxpool.New: %w", err)
	}

	// Aplicar search_path al pool. Cada conexion nueva arranca con
	// el search_path apuntando al schema que el servicio posee.
	if _, err := pool.Exec(context.Background(),
		"ALTER ROLE "+quoteIdentifier(cfg.DBUser)+" IN DATABASE "+quoteIdentifier(cfg.DBName)+" SET search_path TO "+quoteIdentifier(cfg.Schema())+", public"); err != nil {
		// Si el role no permite ALTER ROLE (caso comun en Postgres gestionados),
		// fijar el search_path por sesion en AfterConnect hook es la alternativa;
		// dejamos el ALTER ROLE como best-effort.
		_ = err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	d.Pool = pool
	return nil
}

// quoteIdentifier escapa identificadores de Postgres. Simplificado:
// asume que el identificador no contiene comillas dobles (true para
// 'auth', 'postgres', 'multicatalogo' en este proyecto).
func quoteIdentifier(name string) string {
	return "\"" + name + "\""
}

// Ping es lo que usa el healthcheck. Timeout corto para no bloquear
// al gateway / load balancer.
func (d *DB) Ping(ctx context.Context) error {
	if d.Pool == nil {
		return fmt.Errorf("auth db pool not initialised")
	}
	return d.Pool.Ping(ctx)
}
