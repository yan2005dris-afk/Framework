// Package db envuelve el acceso al schema 'auth' de la BD compartida.
// Convenciones:
//   - BuscarUsuarioPorEmail: el repo del auth-service NUNCA compara
//     password en SQL; devuelve el hash y el caller lo verifica con bcrypt.
//   - Los errores 'no encontrado' se devuelven como ErrNotFound
//     (sentinel; el handler mapea a 401).
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"auth-service/internal/models"
)

// ErrNotFound se devuelve cuando la query no encuentra el recurso.
var ErrNotFound = errors.New("registro no encontrado")

type Repo struct {
	Pool PoolLike
}

// PoolLike es lo minimo que el repo necesita del pool. db.DB lo satisface
// (tiene Pool como pgxpool.Pool), y los tests pueden inyectar un mock.
type PoolLike interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// BuscarUsuarioPorEmail retorna el Usuario (incluido el hash bcrypt) por email.
func (r *Repo) BuscarUsuarioPorEmail(ctx context.Context, email string) (*models.Usuario, error) {
	const q = `SELECT id, email, password, rol FROM auth.users WHERE email = $1`
	var u models.Usuario
	err := r.Pool.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.Password, &u.Rol)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query users: %w", err)
	}
	return &u, nil
}
