-- Migración 006: idempotencia en el checkout (U4-T3, S10).
--
-- El service worker nunca reintenta solo un POST, pero el cliente sí: si el
-- primer intento LLEGÓ al servidor y solo se perdió la respuesta, el reintento
-- crea un segundo pedido con el mismo contenido. Esta tabla es lo que hace que
-- ese reintento sea seguro.
--
-- El par (usuario_id, clave) es la clave primaria y no la clave sola: dos
-- usuarios distintos pueden generar el mismo UUID en teoria practica, y un
-- cliente no debe poder leer el pedido de otro por adivinar una key.
--
-- La FK a pedidos tiene ON DELETE CASCADE: si se borra el pedido, su registro
-- de idempotencia se va con el, y la key queda libre para reusarse.
--
-- Es idempotente: IF NOT EXISTS en tabla e índices.

CREATE TABLE IF NOT EXISTS checkout_idempotencia (
    usuario_id INT NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    clave VARCHAR(255) NOT NULL,
    pedido_id INT NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    creado_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (usuario_id, clave)
);

-- Respalda el borrado en cascada cuando se elimina un pedido: sin este índice
-- Postgres tiene que revisar toda la tabla por cada pedido borrado.
CREATE INDEX IF NOT EXISTS idx_checkout_idempotencia_pedido ON checkout_idempotencia (pedido_id);

SELECT 'migracion 006 aplicada' AS status;