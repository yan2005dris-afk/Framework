#!/bin/sh
#
# Entrypoint del contenedor Backend.
#
# Aplica las migraciones SQL de database/migrations/ (idempotentes, en orden
# alfabetico) antes de arrancar el binario, y recien despues ejecuta el
# servidor. Si una migracion falla, el script sale con codigo != 0 y Docker
# reinicia el contenedor: nunca queda un Backend corriendo contra un esquema
# a medio migrar.
#
# Mismo patron que auth-service/scripts/entrypoint.sh.
#
# Variables (vienen del .env del compose):
#   BACKEND_DB_HOST, BACKEND_DB_PORT, BACKEND_DB_USER,
#   BACKEND_DB_PASSWORD, BACKEND_DB_NAME
set -eu

: "${BACKEND_DB_HOST:=db}"
: "${BACKEND_DB_PORT:=5432}"
: "${BACKEND_DB_USER:=postgres}"
: "${BACKEND_DB_PASSWORD:=postgres}"
: "${BACKEND_DB_NAME:=multicatalogo}"
export PGPASSWORD="${BACKEND_DB_PASSWORD}"

MIGRATIONS_DIR="${BACKEND_MIGRATIONS_DIR:-/home/backend/database/migrations}"
BINARY="${BACKEND_BINARY:-/usr/local/bin/multicatalogo-backend}"

if [ ! -d "${MIGRATIONS_DIR}" ]; then
    echo "[migrations] ERROR: no existe ${MIGRATIONS_DIR}" >&2
    echo "[migrations] El Dockerfile deberia copiar database/migrations/ al contenedor." >&2
    exit 1
fi

# Espera activa a que la BD responda (hasta 30s). Necesaria porque
# depends_on solo espera al healthcheck del contenedor de Postgres, y porque
# el Backend depende de que el auth-service ya haya creado auth.users.
echo "[migrations] esperando a ${BACKEND_DB_HOST}:${BACKEND_DB_PORT}..."
db_ready=0
i=0
while [ "$i" -lt 30 ]; do
    if psql -h "${BACKEND_DB_HOST}" -p "${BACKEND_DB_PORT}" -U "${BACKEND_DB_USER}" \
        -d "${BACKEND_DB_NAME}" -c "SELECT 1" >/dev/null 2>&1; then
        db_ready=1
        break
    fi
    i=$((i + 1))
    sleep 1
done

if [ "${db_ready}" -ne 1 ]; then
    echo "[migrations] ERROR: la BD ${BACKEND_DB_NAME}@${BACKEND_DB_HOST}:${BACKEND_DB_PORT} no responde" >&2
    exit 1
fi

echo "[migrations] aplicando migraciones de ${MIGRATIONS_DIR}/"
for migration in "${MIGRATIONS_DIR}"/*.sql; do
    [ -e "${migration}" ] || continue
    echo "[migrations]   -> $(basename "${migration}")"
    # ON_ERROR_STOP=1: sin esto psql sigue despues de un error y la migracion
    # "parece" aplicada aunque no lo este.
    psql -h "${BACKEND_DB_HOST}" -p "${BACKEND_DB_PORT}" -U "${BACKEND_DB_USER}" \
        -d "${BACKEND_DB_NAME}" -v ON_ERROR_STOP=1 -f "${migration}"
done

echo "[migrations] migraciones aplicadas, arrancando el Backend..."
exec "${BINARY}"