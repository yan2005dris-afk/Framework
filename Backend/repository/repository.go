// Package repository centraliza el acceso y persistencia de datos relacionales en PostgreSQL.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
)

var (
	// ErrNotFound indica que el registro solicitado no existe en la base de datos.
	ErrNotFound = errors.New("registro no encontrado")
	// ErrNoDBConnection indica que el pool de conexiones no está inicializado.
	ErrNoDBConnection = errors.New("conexión a base de datos no disponible")
)

// BuscarUsuarioPorCredenciales consulta la tabla usuarios validando email y contraseña.
func BuscarUsuarioPorCredenciales(email, password string) (*models.Usuario, error) {
	if config.DB == nil {
		return nil, ErrNoDBConnection
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1 AND password = $2`
	var user models.Usuario
	err := config.DB.QueryRow(ctx, query, email, password).Scan(&user.ID, &user.Email, &user.Password, &user.Rol)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error al consultar usuario: %w", err)
	}

	return &user, nil
}

// ObtenerProductos ejecuta SELECT sobre la tabla productos y retorna el catálogo completo.
func ObtenerProductos() ([]models.Producto, error) {
	if config.DB == nil {
		return nil, ErrNoDBConnection
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos ORDER BY id ASC`
	rows, err := config.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar productos: %w", err)
	}
	defer rows.Close()

	productos := make([]models.Producto, 0)
	for rows.Next() {
		var p models.Producto
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria); err != nil {
			return nil, fmt.Errorf("error al escanear producto: %w", err)
		}
		if p.Galeria == nil {
			p.Galeria = []string{}
		}
		productos = append(productos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar productos: %w", err)
	}

	return productos, nil
}

// ObtenerProductoPorID consulta un producto específico por su identificador numérico.
func ObtenerProductoPorID(id int) (*models.Producto, error) {
	if config.DB == nil {
		return nil, ErrNoDBConnection
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos WHERE id = $1`
	var p models.Producto
	err := config.DB.QueryRow(ctx, query, id).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error al consultar producto por id: %w", err)
	}

	if p.Galeria == nil {
		p.Galeria = []string{}
	}

	return &p, nil
}

// ObtenerRed consulta la tabla referidos y reconstruye la jerarquía multinivel en memoria.
func ObtenerRed() (*models.Referido, error) {
	if config.DB == nil {
		return nil, ErrNoDBConnection
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY nivel ASC, id ASC`
	rows, err := config.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar referidos: %w", err)
	}
	defer rows.Close()

	childrenMap := make(map[int][]models.ReferidoRow)
	var rootRow *models.ReferidoRow

	for rows.Next() {
		var row models.ReferidoRow
		if err := rows.Scan(&row.ID, &row.Nombre, &row.Nivel, &row.Ventas, &row.ParentID); err != nil {
			return nil, fmt.Errorf("error al escanear referido: %w", err)
		}

		if row.ParentID == nil {
			rowCopy := row
			rootRow = &rowCopy
		} else {
			childrenMap[*row.ParentID] = append(childrenMap[*row.ParentID], row)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar referidos: %w", err)
	}

	if rootRow == nil {
		return nil, ErrNotFound
	}

	var buildNode func(r models.ReferidoRow) models.Referido
	buildNode = func(r models.ReferidoRow) models.Referido {
		node := models.Referido{
			ID:     r.ID,
			Nombre: r.Nombre,
			Nivel:  r.Nivel,
			Ventas: r.Ventas,
		}
		children := childrenMap[r.ID]
		if len(children) > 0 {
			node.Hijos = make([]models.Referido, 0, len(children))
			for _, child := range children {
				node.Hijos = append(node.Hijos, buildNode(child))
			}
		}
		return node
	}

	root := buildNode(*rootRow)
	return &root, nil
}
