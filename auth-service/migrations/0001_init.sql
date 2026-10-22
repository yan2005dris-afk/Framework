-- Migracion 0001: schema 'auth' para el auth-service.
--
-- El auth-service es dueno del schema 'auth' dentro de la BD compartida
-- 'multicatalogo' (decision arquitectonica L7). Este archivo crea:
--   - el schema 'auth' si no existe
--   - la tabla 'users' (login B2C espejo de backend.usuarios)
--   - seeds con los 2 usuarios demo del Backend (passwords bcrypt
--     cost 10, mismos hashes que init.sql del Backend)
--
-- Convenciones del schema:
--   - 'users' es B2C: cualquier persona con email+password.
--   - 'saas_users' (futuro) es B2B: ربط بين users y tenants para
--     login SaaS multi-tenant.
--   - Las queries del auth-service usan search_path = 'auth' (ver
--     internal/db/db.go).

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id          SERIAL PRIMARY KEY,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    rol         VARCHAR(50) NOT NULL DEFAULT 'cliente',
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_auth_users_email ON auth.users (email);

-- Seed de los 2 usuarios demo del Backend (mismos hashes bcrypt cost 10
-- precomputados y verificados con bcrypt.CompareHashAndPassword).
INSERT INTO auth.users (id, email, password, rol) VALUES
    (1, 'admin@upse.edu.ec',   '$2a$10$wYdwoYJp2ubHFZYw73enietm48txFvoAjiHZHgBsZaKvT2vpf3Mji', 'admin'),
    (2, 'cliente@upse.edu.ec', '$2a$10$7J26b3yIZQ5XC.aiRhqHsOyibUpKimgaC3mKqgrrZQbqqrnbFD2Ua', 'cliente')
ON CONFLICT (id) DO UPDATE SET
    email    = EXCLUDED.email,
    password = EXCLUDED.password,
    rol      = EXCLUDED.rol;

SELECT setval('auth.users_id_seq', COALESCE((SELECT MAX(id) FROM auth.users), 1));

-- Verificacion rapida (no falla si hay rows; si falla, algo se rompio).
SELECT
    (SELECT count(*) FROM auth.users WHERE email = 'admin@upse.edu.ec')   AS admin_count,
    (SELECT count(*) FROM auth.users WHERE email = 'cliente@upse.edu.ec') AS cliente_count;
