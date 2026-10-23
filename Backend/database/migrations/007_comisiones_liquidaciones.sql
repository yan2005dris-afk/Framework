-- Migración 007: Ledger de comisiones y liquidaciones a distribuidores (Fase 4 y Fase 5).
--
-- 1. Tabla comisiones: registra cada comisión devengada por venta multinivel (niveles 1, 2 y 3).
-- Estados: 'pendiente', 'pagada', 'revertida'.
--
-- 2. Tabla liquidaciones: consolida pagos de comisiones pendientes por distribuidor/referido.
-- Estados: 'procesada', 'completada'.

CREATE TABLE IF NOT EXISTS liquidaciones (
    id SERIAL PRIMARY KEY,
    referido_id INT NOT NULL REFERENCES referidos(id),
    total NUMERIC(10, 2) NOT NULL,
    estado VARCHAR(50) NOT NULL DEFAULT 'completada',
    creado_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pagada_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS comisiones (
    id SERIAL PRIMARY KEY,
    pedido_id INT NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    referido_id INT NOT NULL REFERENCES referidos(id),
    nivel INT NOT NULL,
    porcentaje NUMERIC(5, 2) NOT NULL,
    monto NUMERIC(10, 2) NOT NULL,
    estado VARCHAR(50) NOT NULL DEFAULT 'pendiente',
    liquidacion_id INT REFERENCES liquidaciones(id) ON DELETE SET NULL,
    creado_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices de auditoría y rendimiento
CREATE INDEX IF NOT EXISTS idx_comisiones_referido_estado ON comisiones(referido_id, estado);
CREATE INDEX IF NOT EXISTS idx_comisiones_pedido ON comisiones(pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_liquidacion ON comisiones(liquidacion_id);
CREATE INDEX IF NOT EXISTS idx_liquidaciones_referido ON liquidaciones(referido_id);

SELECT 'migracion 007 aplicada: comisiones y liquidaciones' AS status;
