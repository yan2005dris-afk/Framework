// Package repository centraliza el acceso y persistencia de datos relacionales en PostgreSQL y caché en Redis.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	query := `SELECT id, email, password, rol FROM auth.users WHERE email = $1`
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

	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria, COALESCE(activo, true), COALESCE(stock, 0) 
	           FROM productos 
	           WHERE COALESCE(activo, true) = true 
	           ORDER BY id ASC`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar productos: %w", err)
	}
	defer rows.Close()

	productos := make([]models.Producto, 0)
	for rows.Next() {
		var p models.Producto
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Activo, &p.Stock); err != nil {
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

	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria, COALESCE(activo, true), COALESCE(stock, 0) 
	           FROM productos WHERE id = $1`
	var p models.Producto
	err = pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Activo, &p.Stock)
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
//
// idempotencyKey activa la idempotencia (U4-T3, S10). Vacia = comportamiento de
// siempre. Si el usuario ya uso esa key, NO crea un pedido nuevo: devuelve el que
// esa key creo, con replay en true para que el controller responda 200 y no 201.
func (r *PostgresRepository) CrearPedido(usuarioID int, req models.CheckoutRequest, idempotencyKey string) (*models.Pedido, bool, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Iniciar Transacción SQL
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("error al iniciar transacción: %w", err)
	}
	// Asegurar ROLLBACK en caso de salida con error o pánico antes del COMMIT
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 2. Replay de idempotencia. Va antes de calcular el total y antes de tocar
	//    la tabla pedidos: si la key ya se uso, todo lo que viene abajo es
	//    trabajo que se descarta.
	if idempotencyKey != "" {
		var pedidoIDExistente int
		errReplay := tx.QueryRow(ctx,
			`SELECT pedido_id FROM checkout_idempotencia WHERE usuario_id = $1 AND clave = $2`,
			usuarioID, idempotencyKey,
		).Scan(&pedidoIDExistente)
		if errReplay == nil {
			pedido, err := r.leerPedido(ctx, tx, pedidoIDExistente)
			if err != nil {
				return nil, false, err
			}
			return pedido, true, nil
		}
		if !errors.Is(errReplay, pgx.ErrNoRows) {
			return nil, false, fmt.Errorf("error al consultar la idempotencia del checkout: %w", errReplay)
		}
	}

	// 3. Calcular total del pedido
	var total float64
	for _, item := range req.Items {
		total += item.Precio * float64(item.Cantidad)
	}

	// 4. Insertar registro principal en tabla pedidos
	insertPedidoQuery := `
		INSERT INTO pedidos (usuario_id, nombre_cliente, email_cliente, metodo_pago, total, estado)
		VALUES ($1, $2, $3, $4, $5, 'completado')
		RETURNING id, creado_at
	`
	var pedidoID int
	var creadoAt time.Time
	err = tx.QueryRow(ctx, insertPedidoQuery, usuarioID, req.NombreCliente, req.EmailCliente, req.MetodoPago, total).Scan(&pedidoID, &creadoAt)
	if err != nil {
		return nil, false, fmt.Errorf("error al insertar pedido: %w", err)
	}

	// 5. Insertar cada item en tabla pedido_items dentro de la misma transacción
	insertItemQuery := `
		INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio_unitario, subtotal)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, item := range req.Items {
		subtotal := item.Precio * float64(item.Cantidad)
		_, err = tx.Exec(ctx, insertItemQuery, pedidoID, item.ProductoID, item.Cantidad, item.Precio, subtotal)
		if err != nil {
			return nil, false, fmt.Errorf("error al insertar item %d del pedido: %w", item.ProductoID, err)
		}
	}

	// 6. Reservar la key DENTRO de la misma transacción que crea el pedido. Si
	//    esta línea falla, el rollback se lleva el pedido con ella: nunca queda
	//    un pedido sin key, que es el estado que permite duplicar en el reintento.
	if idempotencyKey != "" {
		_, err = tx.Exec(ctx,
			`INSERT INTO checkout_idempotencia (usuario_id, clave, pedido_id) VALUES ($1, $2, $3)`,
			usuarioID, idempotencyKey, pedidoID,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			// 23505 = unique_violation: dos requests con la misma key pasaron a
			// la vez el SELECT del paso 2 y los dos llegaron al INSERT. Gana el
			// que.committeo primero; este hace rollback y devuelve su pedido.
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				_ = tx.Rollback(ctx)
				return r.replayPorClave(usuarioID, idempotencyKey)
			}
			return nil, false, fmt.Errorf("error al registrar la idempotencia del checkout: %w", err)
		}
	}

	// 7. Commit de la transacción
	if err = tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("error al confirmar transacción (commit): %w", err)
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

	return pedido, false, nil
}

// replayPorClave resuelve el caso de carrera: la key ya quedó asociada a otro
// pedido, asi que se devuelve ese. Corre por el pool y no por la transacción
// porque para entonces esta ya hizo rollback.
func (r *PostgresRepository) replayPorClave(usuarioID int, clave string) (*models.Pedido, bool, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var pedidoID int
	err = pool.QueryRow(ctx,
		`SELECT pedido_id FROM checkout_idempotencia WHERE usuario_id = $1 AND clave = $2`,
		usuarioID, clave,
	).Scan(&pedidoID)
	if err != nil {
		return nil, false, fmt.Errorf("error al resolver la carrera de idempotencia: %w", err)
	}
	pedido, err := r.leerPedido(ctx, pool, pedidoID)
	if err != nil {
		return nil, false, err
	}
	return pedido, true, nil
}

// consultor es lo mínimo para leer un pedido: sirve el pool y la transacción
// por igual, para no duplicar la consulta en replayPorClave y en CrearPedido.
type consultor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// leerPedido reconstruye el pedido desde la fila guardada. El replay tiene que
// devolver exactamente la misma respuesta que dio el request original, y eso
// solo se puede armar leyendo los datos que quedaron guardados.
func (r *PostgresRepository) leerPedido(ctx context.Context, q consultor, pedidoID int) (*models.Pedido, error) {
	var pedido models.Pedido
	var creadoAt time.Time
	err := q.QueryRow(ctx,
		`SELECT id, usuario_id, nombre_cliente, email_cliente, metodo_pago, total, estado, creado_at
		 FROM pedidos WHERE id = $1`,
		pedidoID,
	).Scan(&pedido.ID, &pedido.UsuarioID, &pedido.NombreCliente, &pedido.EmailCliente,
		&pedido.MetodoPago, &pedido.Total, &pedido.Estado, &creadoAt)
	if err != nil {
		return nil, fmt.Errorf("error al leer el pedido %d: %w", pedidoID, err)
	}
	pedido.CreadoAt = creadoAt.Format(time.RFC3339)
	return &pedido, nil
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

// Tasas estándar de comisión MLM por nivel (Nivel 1: 10%, Nivel 2: 5%, Nivel 3: 2%)
var TasasComision = map[int]float64{
	1: 0.10,
	2: 0.05,
	3: 0.02,
}

// RegistrarComisionesPedido calcula e inserta en el ledger de comisiones el desglose multinivel.
func (r *PostgresRepository) RegistrarComisionesPedido(pedidoID int, compradorID int, totalPedido float64) ([]models.Comision, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener la cadena ascendente de referidos (hasta 3 niveles) a partir del comprador
	// Si comprador no existe en referidos, asignamos al nodo raíz (0) como nivel 1
	rows, err := pool.Query(ctx, `
		WITH RECURSIVE upline AS (
			SELECT id, parent_id, 1 as nivel
			FROM referidos
			WHERE id = $1
			UNION ALL
			SELECT r.id, r.parent_id, u.nivel + 1
			FROM referidos r
			JOIN upline u ON r.id = u.parent_id
			WHERE u.nivel < 3
		)
		SELECT id, nivel FROM upline WHERE id != $1 OR id = 0;
	`, compradorID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar upline de referidos: %w", err)
	}
	defer rows.Close()

	type uplineNode struct {
		id    int
		nivel int
	}
	var upline []uplineNode
	for rows.Next() {
		var u uplineNode
		if err := rows.Scan(&u.id, &u.nivel); err == nil {
			upline = append(upline, u)
		}
	}

	// Si no hubo upline encontrado (ej. comprador no en referidos), atribuir al nodo 0
	if len(upline) == 0 {
		upline = append(upline, uplineNode{id: 0, nivel: 1})
	}

	var comisiones []models.Comision
	for _, node := range upline {
		tasa, ok := TasasComision[node.nivel]
		if !ok {
			tasa = 0.02
		}
		monto := totalPedido * tasa
		porcentaje := tasa * 100.0

		var id int
		var creadoEn time.Time
		err := pool.QueryRow(ctx, `
			INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto, estado)
			VALUES ($1, $2, $3, $4, $5, 'pendiente')
			RETURNING id, creado_en
		`, pedidoID, node.id, node.nivel, porcentaje, monto).Scan(&id, &creadoEn)
		if err != nil {
			// Si falla una inserción, loguear o continuar según caso
			continue
		}

		comisiones = append(comisiones, models.Comision{
			ID:         id,
			PedidoID:   pedidoID,
			ReferidoID: node.id,
			Nivel:      node.nivel,
			Porcentaje: porcentaje,
			Monto:      monto,
			Estado:     "pendiente",
			CreadoEn:   creadoEn.Format(time.RFC3339),
		})
	}

	return comisiones, nil
}

// ObtenerComisionesPorReferido consulta el historial y balance de comisiones devengadas.
func (r *PostgresRepository) ObtenerComisionesPorReferido(referidoID int) (*models.ResumenComisiones, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT id, pedido_id, referido_id, nivel, porcentaje, monto, estado, liquidacion_id, creado_en
		FROM comisiones
		WHERE referido_id = $1
		ORDER BY id DESC
	`, referidoID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar comisiones: %w", err)
	}
	defer rows.Close()

	resumen := &models.ResumenComisiones{
		ReferidoID: referidoID,
		Comisiones: make([]models.Comision, 0),
	}

	for rows.Next() {
		var c models.Comision
		var creadoEn time.Time
		if err := rows.Scan(&c.ID, &c.PedidoID, &c.ReferidoID, &c.Nivel, &c.Porcentaje, &c.Monto, &c.Estado, &c.LiquidacionID, &creadoEn); err != nil {
			return nil, err
		}
		c.CreadoEn = creadoEn.Format(time.RFC3339)
		if c.Estado == "pendiente" {
			resumen.TotalPendiente += c.Monto
		} else if c.Estado == "pagada" {
			resumen.TotalPagado += c.Monto
		}
		resumen.Comisiones = append(resumen.Comisiones, c)
	}

	return resumen, nil
}

// LiquidarComisiones liquida en una transacción ACID todas las comisiones pendientes de un referido.
func (r *PostgresRepository) LiquidarComisiones(referidoID int) (*models.Liquidacion, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción de liquidación: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Bloquear y sumar comisiones pendientes con FOR UPDATE para evitar doble cobro
	var totalPendiente float64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(monto), 0)
		FROM comisiones
		WHERE referido_id = $1 AND estado = 'pendiente'
		FOR UPDATE
	`, referidoID).Scan(&totalPendiente)
	if err != nil {
		return nil, fmt.Errorf("error al calcular monto a liquidar: %w", err)
	}

	if totalPendiente <= 0 {
		return nil, errors.New("no hay comisiones pendientes para liquidar")
	}

	// 2. Crear registro de liquidación
	var liq models.Liquidacion
	var creadoEn, pagadaEn time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO liquidaciones (referido_id, total, estado, creado_en, pagada_en)
		VALUES ($1, $2, 'completada', NOW(), NOW())
		RETURNING id, referido_id, total, estado, creado_en, pagada_en
	`, referidoID, totalPendiente).Scan(&liq.ID, &liq.ReferidoID, &liq.Total, &liq.Estado, &creadoEn, &pagadaEn)
	if err != nil {
		return nil, fmt.Errorf("error al registrar liquidación: %w", err)
	}
	liq.CreadoEn = creadoEn.Format(time.RFC3339)
	liq.PagadaEn = pagadaEn.Format(time.RFC3339)

	// 3. Actualizar comisiones asociadas a 'pagada' con el id de liquidación
	_, err = tx.Exec(ctx, `
		UPDATE comisiones
		SET estado = 'pagada', liquidacion_id = $1
		WHERE referido_id = $2 AND estado = 'pendiente'
	`, liq.ID, referidoID)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar estado de comisiones: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error al confirmar liquidación: %w", err)
	}

	return &liq, nil
}

// ListarLiquidaciones devuelve el historial global de liquidaciones realizadas (Admin).
func (r *PostgresRepository) ListarLiquidaciones() ([]models.Liquidacion, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT id, referido_id, total, estado, creado_en, pagada_en
		FROM liquidaciones
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("error al listar liquidaciones: %w", err)
	}
	defer rows.Close()

	var liquidaciones []models.Liquidacion
	for rows.Next() {
		var l models.Liquidacion
		var cEn, pEn time.Time
		if err := rows.Scan(&l.ID, &l.ReferidoID, &l.Total, &l.Estado, &cEn, &pEn); err != nil {
			return nil, err
		}
		l.CreadoEn = cEn.Format(time.RFC3339)
		l.PagadaEn = pEn.Format(time.RFC3339)
		liquidaciones = append(liquidaciones, l)
	}

	return liquidaciones, nil
}

// ListarLiquidacionesPorReferido devuelve las liquidaciones correspondientes a un distribuidor específico.
func (r *PostgresRepository) ListarLiquidacionesPorReferido(referidoID int) ([]models.Liquidacion, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT id, referido_id, total, estado, creado_en, pagada_en
		FROM liquidaciones
		WHERE referido_id = $1
		ORDER BY id DESC
	`, referidoID)
	if err != nil {
		return nil, fmt.Errorf("error al listar liquidaciones de referido: %w", err)
	}
	defer rows.Close()

	var liquidaciones []models.Liquidacion
	for rows.Next() {
		var l models.Liquidacion
		var cEn, pEn time.Time
		if err := rows.Scan(&l.ID, &l.ReferidoID, &l.Total, &l.Estado, &cEn, &pEn); err != nil {
			return nil, err
		}
		l.CreadoEn = cEn.Format(time.RFC3339)
		l.PagadaEn = pEn.Format(time.RFC3339)
		liquidaciones = append(liquidaciones, l)
	}

	return liquidaciones, nil
}

// ObtenerProductosAdmin retorna todos los productos incluidos los inactivos (Backoffice Admin).
func (r *PostgresRepository) ObtenerProductosAdmin() ([]models.Producto, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, descripcion, precio, categoria, img, galeria, COALESCE(activo, true), COALESCE(stock, 0) 
	           FROM productos ORDER BY id ASC`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar productos admin: %w", err)
	}
	defer rows.Close()

	productos := make([]models.Producto, 0)
	for rows.Next() {
		var p models.Producto
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Activo, &p.Stock); err != nil {
			return nil, fmt.Errorf("error al escanear producto: %w", err)
		}
		if p.Galeria == nil {
			p.Galeria = []string{}
		}
		productos = append(productos, p)
	}
	return productos, nil
}

// CrearProducto inserta un nuevo artículo en la tabla productos.
func (r *PostgresRepository) CrearProducto(req models.ProductoRequest) (*models.Producto, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	galeria := req.Galeria
	if galeria == nil {
		galeria = []string{}
	}
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	stock := 100
	if req.Stock != nil {
		stock = *req.Stock
	}

	var p models.Producto
	query := `
		INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, activo, stock)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, nombre, descripcion, precio, categoria, img, galeria, activo, stock
	`
	err = pool.QueryRow(ctx, query, req.Nombre, req.Descripcion, req.Precio, req.Categoria, req.Img, galeria, activo, stock).
		Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Activo, &p.Stock)
	if err != nil {
		return nil, fmt.Errorf("error al insertar producto: %w", err)
	}
	return &p, nil
}

// ActualizarProducto permite edición parcial o total de los campos del producto.
func (r *PostgresRepository) ActualizarProducto(id int, req models.ProductoRequest) (*models.Producto, error) {
	pool, err := r.getPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p, err := r.ObtenerProductoPorID(id)
	if err != nil {
		return nil, err
	}

	if req.Nombre != "" {
		p.Nombre = req.Nombre
	}
	if req.Descripcion != "" {
		p.Descripcion = req.Descripcion
	}
	if req.Precio > 0 {
		p.Precio = req.Precio
	}
	if req.Categoria != "" {
		p.Categoria = req.Categoria
	}
	if req.Img != "" {
		p.Img = req.Img
	}
	if req.Galeria != nil {
		p.Galeria = req.Galeria
	}
	if req.Activo != nil {
		p.Activo = *req.Activo
	}
	if req.Stock != nil {
		p.Stock = *req.Stock
	}

	query := `
		UPDATE productos 
		SET nombre = $1, descripcion = $2, precio = $3, categoria = $4, img = $5, galeria = $6, activo = $7, stock = $8
		WHERE id = $9
		RETURNING id, nombre, descripcion, precio, categoria, img, galeria, activo, stock
	`
	err = pool.QueryRow(ctx, query, p.Nombre, p.Descripcion, p.Precio, p.Categoria, p.Img, p.Galeria, p.Activo, p.Stock, id).
		Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Activo, &p.Stock)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar producto: %w", err)
	}
	return p, nil
}

// EliminarProductoLogico marca el producto como inactivo (activo = false) sin borrar la fila física.
func (r *PostgresRepository) EliminarProductoLogico(id int) error {
	pool, err := r.getPool()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := pool.Exec(ctx, `UPDATE productos SET activo = false WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error al desactivar producto: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
