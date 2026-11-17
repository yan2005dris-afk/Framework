-- Migración 002: tablas para el flujo de checkout transaccional (U2-T4).
--
-- El init.sql original contiene estas mismas tablas; este archivo existe
-- como migración explícita para aplicarse sobre volúmenes creados antes
-- de la introducción del checkout ACID en U2-T4, donde /api/checkout fallaba
-- con "relation 'pedidos' does not exist".
--
-- Es idempotente: usa IF NOT EXISTS y los CREATE INDEX también.

-- 1. Tabla pedidos (cabecera de la compra)
CREATE TABLE IF NOT EXISTS backend.pedidos (
    id SERIAL PRIMARY KEY,
    usuario_id INT NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    nombre_cliente VARCHAR(255) NOT NULL,
    email_cliente VARCHAR(255) NOT NULL,
    metodo_pago VARCHAR(50) NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    estado VARCHAR(50) NOT NULL DEFAULT 'completado',
    creado_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabla pedido_items (líneas del pedido)
CREATE TABLE IF NOT EXISTS backend.pedido_items (
    id SERIAL PRIMARY KEY,
    pedido_id INT NOT NULL REFERENCES backend.pedidos(id) ON DELETE CASCADE,
    producto_id INT NOT NULL REFERENCES backend.productos(id),
    cantidad INT NOT NULL,
    precio_unitario NUMERIC(10, 2) NOT NULL,
    subtotal NUMERIC(10, 2) NOT NULL
);

-- 3. Índices declarados en la checklist U2-T3
CREATE INDEX IF NOT EXISTS idx_pedidos_email ON backend.pedidos(email_cliente);
CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido ON backend.pedido_items(pedido_id);

SELECT 'migracion 002 aplicada' AS status;
