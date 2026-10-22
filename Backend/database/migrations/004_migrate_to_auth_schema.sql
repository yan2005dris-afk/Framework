-- Migración 004: autenticacion contra auth.users (U3-T2 GA5)
--
-- Contexto: el auth-service es dueño del schema 'auth' y su tabla
-- `auth.users`. El Backend ya no crea ni escribe usuarios: autentica contra
-- `auth.users` (ver Backend/repository/repository.go, GA5).
--
-- Qué hace, en orden:
--   1. Copia a auth.users los usuarios que quedaron en public.usuarios y no
--      existan todavía en auth.users (volúmenes creados antes de GA5).
--   2. Reapunta la FK de public.pedidos.usuario_id a auth.users(id).
--   3. Elimina public.usuarios, ya redundante.
--
-- Es idempotente: se puede aplicar N veces.
--
-- Pre-requisito: `auth.users` debe existir. La crea el entrypoint del
-- auth-service (auth-service/migrations/0001_init.sql). Si el Backend arranca
-- antes, este script falla con un error claro y Docker reintenta el arranque.
--
-- Nota sobre volúmenes nuevos: `init.sql` crea `pedidos` sin FK porque corre
-- antes que el auth-service; este archivo agrega la FK.

-- 1. Migración de datos, solo si public.usuarios todavía existe.
DO $mig$
BEGIN
    IF to_regclass('public.usuarios') IS NULL THEN
        RAISE NOTICE 'public.usuarios no existe: no hay datos que migrar';
        RETURN;
    END IF;

    IF to_regclass('auth.users') IS NULL THEN
        RAISE EXCEPTION 'auth.users no existe: el auth-service todavía no inicializó el schema auth';
    END IF;

    INSERT INTO auth.users (id, email, password, rol, created_at, updated_at)
    SELECT u.id, u.email, u.password, u.rol, NOW(), NOW()
    FROM public.usuarios u
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'usuarios migrados a auth.users';
END
$mig$;

-- 2. Reapuntar la FK de pedidos a auth.users (idempotente: primero la saca).
ALTER TABLE public.pedidos DROP CONSTRAINT IF EXISTS pedidos_usuario_id_fkey;
ALTER TABLE public.pedidos
    ADD CONSTRAINT pedidos_usuario_id_fkey
    FOREIGN KEY (usuario_id) REFERENCES auth.users(id) ON DELETE CASCADE;

-- 3. Eliminar la tabla redundante. Sin CASCADE a propósito: si algo más
-- dependiera de ella, el error debe verse en el log del contenedor en vez de
-- borrar esa dependencia en silencio.
DROP TABLE IF EXISTS public.usuarios;

SELECT 'migracion 004 aplicada: usuarios unificados en auth.users' AS status;