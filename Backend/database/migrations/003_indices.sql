-- Migracion 003: indices declarados en el checklist U2-T3 que faltan
-- en init.sql. Se aplica una sola vez sobre el volumen pre-existente;
-- para volumenes nuevos, los indices se crearan en el primer
-- arranque al ejecutar init.sql (futura actualizacion del init.sql
-- para moverlos ahi).
--
-- Esta migracion es idempotente: usa IF NOT EXISTS.

-- 1. Indice sobre referidos.parent_id: clave para la reconstruccion
--    del arbol multinivel en repository.ObtenerRed, que recorre la
--    jerarquia con WHERE id IN (children of parent).
CREATE INDEX IF NOT EXISTS idx_referidos_parent_id ON referidos(parent_id);

-- 2. Indice sobre productos.categoria: soporte para filtros del
--    catalogo por categoria (Catalogo.tsx y futuras rutas /productos?categoria=...).
--    Utile como base cuando crezca el catalogo.
CREATE INDEX IF NOT EXISTS idx_productos_categoria ON productos(categoria);

SELECT 'migracion 003 aplicada' AS status;
