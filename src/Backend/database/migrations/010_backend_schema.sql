-- Migración 010: Crear schema 'backend' y migrar todas las tablas desde 'public'.
--
-- Contexto: la documentación (docs/architecture.md) establece que el Backend
-- debe poseer el schema 'backend.*', pero las tablas siempre vivieron en
-- 'public' por defecto. Esta migración corrige ese acoplamiento implícito.
--
-- Qué hace, en orden:
--   1. Crea el schema 'backend' si no existe.
--   2. Mueve las 7 tablas del Backend a backend.*.
--   3. Reapunta las FK entre tablas (ya que ALTER TABLE SET SCHEMA no
--      actualiza las FK automáticamente para tablas que también se mueven).
--   4. Mueve los índices asociados.
--
-- Es idempotente: se puede aplicar N veces sin error.
--
-- Nota: auth.users NO se mueve — pertenece al auth-service (schema 'auth').
-- Las FK hacia auth.users(id) se mantienen cruzando schemas, lo cual es
-- correcto en PostgreSQL.

-- 1. Crear el schema
CREATE SCHEMA IF NOT EXISTS backend;

-- 2. Mover tablas a backend.* (idempotente: verifica existencia primero)
DO $migrate$
DECLARE
    tbl text;
    old_schema text := 'public';
    new_schema text := 'backend';
BEGIN
    -- Lista de tablas del Backend
    FOREACH tbl IN ARRAY ['productos', 'referidos', 'pedidos', 'pedido_items',
                           'checkout_idempotencia', 'comisiones', 'liquidaciones']
    LOOP
        IF to_regclass(quote_ident(old_schema) || '.' || quote_ident(tbl)) IS NOT NULL THEN
            EXECUTE format('ALTER TABLE %I.%I SET SCHEMA %I',
                           old_schema, tbl, new_schema);
            RAISE NOTICE 'tabla % migrada a backend.%', tbl, tbl;
        ELSE
            -- Si ya está en backend, no hacer nada
            IF to_regclass(quote_ident(new_schema) || '.' || quote_ident(tbl)) IS NOT NULL THEN
                RAISE NOTICE 'tabla % ya está en backend.%', tbl, tbl;
            ELSE
                RAISE NOTICE 'tabla % no encontrada en public ni backend, saltando', tbl;
            END IF;
        END IF;
    END LOOP;
END
$migrate$;

-- 3. Reapuntar las FK entre tablas backend.*
--    ALTER TABLE ... SET SCHEMA ya reapunta las FK internamente en PG 15+,
--    pero para versiones anteriores o seguridad, las reapuntamos explícitamente.

-- pedidos.usuario_id → auth.users(id) (cross-schema, ya correcto)
ALTER TABLE backend.pedidos DROP CONSTRAINT IF EXISTS pedidos_usuario_id_fkey;
ALTER TABLE backend.pedidos
    ADD CONSTRAINT pedidos_usuario_id_fkey
    FOREIGN KEY (usuario_id) REFERENCES auth.users(id) ON DELETE CASCADE;

-- pedido_items.pedido_id → backend.pedidos(id)
ALTER TABLE backend.pedido_items DROP CONSTRAINT IF EXISTS pedido_items_pedido_id_fkey;
ALTER TABLE backend.pedido_items
    ADD CONSTRAINT pedido_items_pedido_id_fkey
    FOREIGN KEY (pedido_id) REFERENCES backend.pedidos(id) ON DELETE CASCADE;

-- pedido_items.producto_id → backend.productos(id)
ALTER TABLE backend.pedido_items DROP CONSTRAINT IF EXISTS pedido_items_producto_id_fkey;
ALTER TABLE backend.pedido_items
    ADD CONSTRAINT pedido_items_producto_id_fkey
    FOREIGN KEY (producto_id) REFERENCES backend.productos(id);

-- referidos.parent_id → backend.referidos(id)
ALTER TABLE backend.referidos DROP CONSTRAINT IF EXISTS referidos_parent_id_fkey;
ALTER TABLE backend.referidos
    ADD CONSTRAINT referidos_parent_id_fkey
    FOREIGN KEY (parent_id) REFERENCES backend.referidos(id) ON DELETE CASCADE;

-- checkout_idempotencia.usuario_id → auth.users(id) (cross-schema)
ALTER TABLE backend.checkout_idempotencia DROP CONSTRAINT IF EXISTS checkout_idempotencia_usuario_id_fkey;
ALTER TABLE backend.checkout_idempotencia
    ADD CONSTRAINT checkout_idempotencia_usuario_id_fkey
    FOREIGN KEY (usuario_id) REFERENCES auth.users(id) ON DELETE CASCADE;

-- checkout_idempotencia.pedido_id → backend.pedidos(id)
ALTER TABLE backend.checkout_idempotencia DROP CONSTRAINT IF EXISTS checkout_idempotencia_pedido_id_fkey;
ALTER TABLE backend.checkout_idempotencia
    ADD CONSTRAINT checkout_idempotencia_pedido_id_fkey
    FOREIGN KEY (pedido_id) REFERENCES backend.pedidos(id) ON DELETE CASCADE;

-- comisiones.pedido_id → backend.pedidos(id)
ALTER TABLE backend.comisiones DROP CONSTRAINT IF EXISTS comisiones_pedido_id_fkey;
ALTER TABLE backend.comisiones
    ADD CONSTRAINT comisiones_pedido_id_fkey
    FOREIGN KEY (pedido_id) REFERENCES backend.pedidos(id) ON DELETE CASCADE;

-- comisiones.referido_id → backend.referidos(id)
ALTER TABLE backend.comisiones DROP CONSTRAINT IF EXISTS comisiones_referido_id_fkey;
ALTER TABLE backend.comisiones
    ADD CONSTRAINT comisiones_referido_id_fkey
    FOREIGN KEY (referido_id) REFERENCES backend.referidos(id);

-- comisiones.liquidacion_id → backend.liquidaciones(id)
ALTER TABLE backend.comisiones DROP CONSTRAINT IF EXISTS comisiones_liquidacion_id_fkey;
ALTER TABLE backend.comisiones
    ADD CONSTRAINT comisiones_liquidacion_id_fkey
    FOREIGN KEY (liquidacion_id) REFERENCES backend.liquidaciones(id) ON DELETE SET NULL;

-- liquidaciones.referido_id → backend.referidos(id)
ALTER TABLE backend.liquidaciones DROP CONSTRAINT IF EXISTS liquidaciones_referido_id_fkey;
ALTER TABLE backend.liquidaciones
    ADD CONSTRAINT liquidaciones_referido_id_fkey
    FOREIGN KEY (referido_id) REFERENCES backend.referidos(id);

SELECT 'migracion 010 aplicada: tablas movidas a backend.*' AS status;
