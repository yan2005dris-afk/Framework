-- Migración 008: Soporte para borrado lógico (activo) y stock en productos (Fase 2 backoffice).
-- Idempotente: añade columnas si no existen.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'productos' AND column_name = 'activo'
    ) THEN
        ALTER TABLE productos ADD COLUMN activo BOOLEAN NOT NULL DEFAULT true;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'productos' AND column_name = 'stock'
    ) THEN
        ALTER TABLE productos ADD COLUMN stock INT NOT NULL DEFAULT 100;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_productos_activo ON productos(activo);
CREATE INDEX IF NOT EXISTS idx_productos_categoria ON productos(categoria);

SELECT 'migracion 008 aplicada: productos activo y stock' AS status;
