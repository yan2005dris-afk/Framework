-- Inicialización del esquema y datos para multicatalogo
--
-- NOTA: La tabla 'usuarios' ha sido migrada al schema 'auth' gestionado por
-- el auth-service. Este script ya no crea la tabla public.usuarios ni los seeds
-- de usuarios. El Backend ahora consulta auth.users para autenticación.

-- 1. Tabla productos
CREATE TABLE IF NOT EXISTS productos (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    descripcion TEXT NOT NULL,
    precio NUMERIC(10, 2) NOT NULL,
    categoria VARCHAR(100) NOT NULL,
    img TEXT NOT NULL,
    galeria TEXT[] NOT NULL DEFAULT '{}'
);

-- 3. Tabla referidos
CREATE TABLE IF NOT EXISTS referidos (
    id INT PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    nivel INT NOT NULL,
    ventas NUMERIC(10, 2) NOT NULL,
    parent_id INT REFERENCES referidos(id) ON DELETE CASCADE
);

-- 4. Tabla pedidos
-- usuario_id NO lleva REFERENCES aqui: este script corre como
-- docker-entrypoint-initdb.d, antes de que el auth-service cree auth.users.
-- La FK hacia auth.users(id) la agrega la migracion
-- database/migrations/003_migrate_to_auth_schema.sql, que corre en el
-- entrypoint del Backend (ver Backend/scripts/apply_migrations.sh).
CREATE TABLE IF NOT EXISTS pedidos (
    id SERIAL PRIMARY KEY,
    usuario_id INT NOT NULL,
    nombre_cliente VARCHAR(255) NOT NULL,
    email_cliente VARCHAR(255) NOT NULL,
    metodo_pago VARCHAR(50) NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    estado VARCHAR(50) NOT NULL DEFAULT 'completado',
    creado_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Tabla pedido_items
CREATE TABLE IF NOT EXISTS pedido_items (
    id SERIAL PRIMARY KEY,
    pedido_id INT NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    producto_id INT NOT NULL REFERENCES productos(id),
    cantidad INT NOT NULL,
    precio_unitario NUMERIC(10, 2) NOT NULL,
    subtotal NUMERIC(10, 2) NOT NULL
);

-- Los datos de demostracion (catalogo y arbol de referidos) NO viven aca:
-- estan en database/migrations/005_datos_demo.sql, que el entrypoint del
-- Backend aplica en cada arranque. Asi un volumen nuevo y uno preexistente
-- terminan con el mismo catalogo, sin depender de initdb.
