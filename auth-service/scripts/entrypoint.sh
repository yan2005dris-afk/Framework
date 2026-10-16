#!/usr/bin/env bash
#
# Entrypoint del contenedor auth-service.
# Aplica las migraciones SQL (idempotente, crea schema 'auth' + tablas)
# antes de arrancar el binario. Si falla, sale con codigo 1 y docker
# reinicia el contenedor.
set -euo pipefail

# Carga variables del entorno del proceso (las del .env del compose).
: "${AUTH_DB_HOST:=db}"
: "${AUTH_DB_PORT:=5432}"
: "${AUTH_DB_USER:=postgres}"
: "${AUTH_DB_PASSWORD:=postgres}"
: "${AUTH_DB_NAME:=multicatalogo}"
: "${AUTH_DB_SCHEMA:=auth}"
export PGPASSWORD="${AUTH_DB_PASSWORD}"

# Espera a que la BD responda (espera activa hasta 30s).
echo "[entrypoint] esperando a ${AUTH_DB_HOST}:${AUTH_DB_PORT}..."
for i in $(seq 1 30); do
    if psql -h "${AUTH_DB_HOST}" -p "${AUTH_DB_PORT}" -U "${AUTH_DB_USER}" -d "${AUTH_DB_NAME}" -c "SELECT 1" >/dev/null 2>&1; then
        echo "[entrypoint] DB lista"
        break
    fi
    sleep 1
done

# Aplica las migraciones.
echo "[entrypoint] aplicando migraciones de /home/auth/migrations/"
for f in /home/auth/migrations/*.sql; do
    [ -e "${f}" ] || continue
    echo "[entrypoint]   -> $(basename "$f")"
    psql -h "${AUTH_DB_HOST}" -p "${AUTH_DB_PORT}" -U "${AUTH_DB_USER}" -d "${AUTH_DB_NAME}" -v ON_ERROR_STOP=1 -f "${f}"
done

# Arranca el servidor (el CMD del Dockerfile). exec para que senales
# (SIGTERM en shutdown de docker) lleguen al proceso Go.
echo "[entrypoint] arrancando auth-service..."
exec /usr/local/bin/auth-service
