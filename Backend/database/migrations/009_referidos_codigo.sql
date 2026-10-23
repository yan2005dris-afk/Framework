-- Migración 009: Columna código de referido en referidos y FK opcional en auth.users (Fase 1).
-- Idempotente.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'referidos' AND column_name = 'codigo'
    ) THEN
        ALTER TABLE referidos ADD COLUMN codigo VARCHAR(50) UNIQUE;
    END IF;
END $$;

-- Generar códigos para los nodos semilla si aún no tienen
UPDATE referidos SET codigo = 'TU' WHERE id = 0 AND (codigo IS NULL OR codigo = '');
UPDATE referidos SET codigo = 'ANA' WHERE id = 1 AND (codigo IS NULL OR codigo = '');
UPDATE referidos SET codigo = 'LUIS' WHERE id = 2 AND (codigo IS NULL OR codigo = '');
UPDATE referidos SET codigo = 'MARTA' WHERE id = 3 AND (codigo IS NULL OR codigo = '');
UPDATE referidos SET codigo = 'CARLOS' WHERE id = 4 AND (codigo IS NULL OR codigo = '');
UPDATE referidos SET codigo = 'SOFIA' WHERE id = 5 AND (codigo IS NULL OR codigo = '');
UPDATE referidos SET codigo = 'MARCO' WHERE id = 6 AND (codigo IS NULL OR codigo = '');
UPDATE referidos SET codigo = 'DIANA' WHERE id = 7 AND (codigo IS NULL OR codigo = '');

CREATE INDEX IF NOT EXISTS idx_referidos_codigo ON referidos(codigo);

SELECT 'migracion 009 aplicada: codigo de referido' AS status;
