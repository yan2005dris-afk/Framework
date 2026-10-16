#!/usr/bin/env bash
#
# Aplica las migraciones SQL del auth-service al schema 'auth' de la BD
# compartida. Convencion: corre todos los .sql de migrations/ en orden
# alfabetico.
#
# Requiere: psql en PATH, AUTH_DB_* en .env o env shell.
#
# Uso:
#   ./scripts/apply.sh                    # usa .env del directorio padre
#   AUTH_DB_HOST=... ./scripts/apply.sh   # override por env

set -euo pipefail

# Carga .env si existe (en el directorio del servicio, no en scripts/).
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SERVICE_DIR="$( cd "${SCRIPT_DIR}/.." && pwd )"
if [ -f "${SERVICE_DIR}/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    . "${SERVICE_DIR}/.env"
    set +a
fi

: "${AUTH_DB_HOST:=localhost}"
: "${AUTH_DB_PORT:=5432}"
: "${AUTH_DB_USER:=postgres}"
: "${AUTH_DB_PASSWORD:=postgres}"
: "${AUTH_DB_NAME:=multicatalogo}"

PSQL="psql -h ${AUTH_DB_HOST} -p ${AUTH_DB_PORT} -U ${AUTH_DB_USER} -d ${AUTH_DB_NAME} -v ON_ERROR_STOP=1 --set AUTOCOMMIT=off"
export PGPASSWORD="${AUTH_DB_PASSWORD}"

DIR="${SERVICE_DIR}/migrations"
if [ ! -d "${DIR}" ]; then
    echo "no existe ${DIR}" >&2
    exit 1
fi

# Aplica en orden alfabetico. Convencion: NNNN_descripcion.sql.
for f in "${DIR}"/*.sql; do
    [ -e "${f}" ] || continue
    echo ">>> applying $(basename "$f")"
    ${PSQL} -f "${f}"
done

echo "OK migrations applied"
