// Package repository centraliza el acceso y persistencia de datos relacionales en PostgreSQL y caché en Redis.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
)

var (
	// ErrNotFound indica que el registro solicitado no existe en la base de datos.
	ErrNotFound = errors.New("registro no encontrado")
	// ErrNoDBConnection indica que el pool de conexiones no está inicializado.
	ErrNoDBConnection = errors.New("conexión a base de datos no disponible")
	// ErrCantidadInvalida indica que un item del checkout tiene cantidad <= 0 o faltante.
	ErrCantidadInvalida = errors.New("cantidad debe ser mayor que cero")
)

// PostgresRepository implementa DatabaseRepository usando un pool de conexiones pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository crea una nueva instancia de PostgresRepository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// getPool retorna el pool inyectado o el global como fallback.
func (r *PostgresRepository) getPool() (*pgxpool.Pool, error) {
	if r != nil && r.pool != nil {
		return r.pool, nil
	}
	if config.DB != nil {
		return config.DB, nil
	}
	return nil, ErrNoDBConnection
}

// BuscarUsuarioPorEmail trae al usuario (incluido el hash bcrypt del password) por email.
// La verificación del password se hace en el controller con bcrypt.CompareHashAndPassword.
func (r *PostgresRepository) BuscarUsuarioPorEmail(email string) (*models.Usuario, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, email, password, rol FROM usuarios WHERE email = $1`
	var user models.Usuario
	err = pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Rol)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error al consultar usuario: %w", err)
	}

	return &user, nil
}

// ObtenerProductos ejecuta SELECT sobre la tabla productos y retorna el catálogo completo.
func (r *PostgresRepository) ObtenerProductos() ([]models.Producto, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos ORDER BY id ASC`
	rows, err := pool.Query(ctx, query)
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
func (r *PostgresRepository) ObtenerProductoPorID(id int) (*models.Producto, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos WHERE id = $1`
	var p models.Producto
	err = pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria)
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
func (r *PostgresRepository) ObtenerRed() (*models.Referido, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY nivel ASC, id ASC`
	rows, err := pool.Query(ctx, query)
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

// CrearPedido ejecuta una transacción SQL ACID en PostgreSQL para registrar un pedido y sus items.
func (r *PostgresRepository) CrearPedido(usuarioID int, req models.CheckoutRequest) (*models.Pedido, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Iniciar Transacción SQL
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción: %w", err)
	}
	// Asegurar ROLLBACK en caso de salida con error o pánico antes del COMMIT
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 2. Calcular total del pedido
	var total float64
	for _, item := range req.Items {
		total += item.Precio * float64(item.Cantidad)
	}

	// 3. Insertar registro principal en tabla pedidos
	insertPedidoQuery := `
		INSERT INTO pedidos (usuario_id, nombre_cliente, email_cliente, metodo_pago, total, estado)
		VALUES ($1, $2, $3, $4, $5, 'completado')
		RETURNING id, creado_at
	`
	var pedidoID int
	var creadoAt time.Time
	err = tx.QueryRow(ctx, insertPedidoQuery, usuarioID, req.NombreCliente, req.EmailCliente, req.MetodoPago, total).Scan(&pedidoID, &creadoAt)
	if err != nil {
		return nil, fmt.Errorf("error al insertar pedido: %w", err)
	}

	// 4. Insertar cada item en tabla pedido_items dentro de la misma transacción
	insertItemQuery := `
		INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio_unitario, subtotal)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, item := range req.Items {
		subtotal := item.Precio * float64(item.Cantidad)
		_, err = tx.Exec(ctx, insertItemQuery, pedidoID, item.ProductoID, item.Cantidad, item.Precio, subtotal)
		if err != nil {
			return nil, fmt.Errorf("error al insertar item %d del pedido: %w", item.ProductoID, err)
		}
	}

	// 5. Commit de la transacción
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción (commit): %w", err)
	}

	pedido := &models.Pedido{
		ID:            pedidoID,
		UsuarioID:     usuarioID,
		NombreCliente: req.NombreCliente,
		EmailCliente:  req.EmailCliente,
		MetodoPago:    req.MetodoPago,
		Total:         total,
		Estado:        "completado",
		CreadoAt:      creadoAt.Format(time.RFC3339),
	}

	return pedido, nil
}

// ActualizarComisiones actualiza los saldos/ventas de la red ascendente de forma recursiva o jerárquica.
func (r *PostgresRepository) ActualizarComisiones(usuarioID int, totalPedido float64) error {
	pool, err := r.getPool()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Actualizamos las ventas de la red sumando la venta a los nodos principales (raíz y nodo comprador)
	// Para la simulación didáctica de la red multinivel en la base de datos, incrementamos las ventas:
	// 1. Nodo comprador si existe en referidos (id = usuarioID)
	// 2. Nodo raíz (id = 0 'Tú') que consolida las ventas globales de la red
	query := `
		UPDATE referidos 
		SET ventas = ventas + $1 
		WHERE id IN (0, $2)
	`
	_, err = pool.Exec(ctx, query, totalPedido, usuarioID)
	if err != nil {
		return fmt.Errorf("error al actualizar comisiones de referidos: %w", err)
	}

	return nil
}

// RedisCacheRepository implementa CacheRepository usando el cliente go-redis.
type RedisCacheRepository struct {
	client *redis.Client
}

// NewRedisCacheRepository crea una nueva instancia de RedisCacheRepository.
func NewRedisCacheRepository(client *redis.Client) *RedisCacheRepository {
	return &RedisCacheRepository{client: client}
}

// getClient retorna el cliente inyectado o el global como fallback.
func (c *RedisCacheRepository) getClient() *redis.Client {
	if c != nil && c.client != nil {
		return c.client
	}
	return config.RedisClient
}

// Get obtiene un valor de la caché.
func (c *RedisCacheRepository) Get(key string) ([]byte, error) {
	client := c.getClient()
	if client == nil {
		return nil, errors.New("cliente de redis no disponible")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return client.Get(ctx, key).Bytes()
}

// Set guarda un valor con expiración TTL en la caché.
func (c *RedisCacheRepository) Set(key string, value []byte, ttlSeconds int) error {
	client := c.getClient()
	if client == nil {
		return errors.New("cliente de redis no disponible")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
}

// Del elimina una o más claves de la caché.
func (c *RedisCacheRepository) Del(key string) error {
	client := c.getClient()
	if client == nil {
		return errors.New("cliente de redis no disponible")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return client.Del(ctx, key).Err()
}

// Funciones de compatibilidad hacia atrás delegadas a la implementación por defecto

var defaultRepo = &PostgresRepository{}

// BuscarUsuarioPorEmail (wrapper de compatibilidad) delega en la implementación por defecto.
func BuscarUsuarioPorEmail(email string) (*models.Usuario, error) {
	return defaultRepo.BuscarUsuarioPorEmail(email)
}

func ObtenerProductos() ([]models.Producto, error) {
	return defaultRepo.ObtenerProductos()
}

func ObtenerProductoPorID(id int) (*models.Producto, error) {
	return defaultRepo.ObtenerProductoPorID(id)
}

func ObtenerRed() (*models.Referido, error) {
	return defaultRepo.ObtenerRed()
}
