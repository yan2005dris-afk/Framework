#!/usr/bin/env bash
#
# Script para ejecutar tests de integración de autenticación.
#
# Pre-requisitos:
#   - docker compose up -d --build
#   - Esperar ~15s para que healthchecks pasen
#
# Uso:
#   bash tests/integration/run_integration_tests.sh         # ejecuta todos
#   bash tests/integration/run_integration_tests.sh go     # solo Go
#   bash tests/integration/run_integration_tests.sh py     # solo Python
#   bash tests/integration/run_integration_tests.sh --skip-stack  # asume stack ya arriba
#
# La sección Python corre dos archivos:
#   - test_auth_integration.py: consistencia de JWT entre servicios.
#   - test_saas_rls_integration.py: aislamiento multi-tenant por RLS. Necesita
#     la BD (`docker compose up -d db`) y que el SaaS haya corrido
#     `alembic upgrade head`, que es lo que crea el rol de la app.
#
# Los dos usan el venv de fastAPI_SaaS (es el unico del repo con psycopg2 para
# el test de RLS y httpx para el de auth). Ojo: test_auth_integration.py ademas
# importa `jwt` (pyjwt), que ningun proyecto del repo declara hoy.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
BACKEND_URL="${BACKEND_URL:-http://localhost:3000}"
AUTHSVC_URL="${AUTHSVC_URL:-http://localhost:3001}"
SAAS_URL="${SAAS_URL:-http://localhost:8000}"

# Tests de RLS del SaaS: necesitan la BD del stack y el rol de la app.
# El rol admin solo se usa para el test que documenta que un superuser bypasea
# las politicas.
SAAS_DB_HOST="${SAAS_DB_HOST:-localhost}"
SAAS_DB_PORT="${SAAS_DB_PORT:-5432}"
SAAS_DB_NAME="${SAAS_DB_NAME:-multicatalogo}"
SAAS_DB_APP_ROLE="${SAAS_DB_APP_ROLE:-saas_app}"
SAAS_DB_USER="${SAAS_DB_USER:-$SAAS_DB_APP_ROLE}"
SAAS_DB_PASSWORD="${SAAS_DB_PASSWORD:-saas_app}"
SAAS_DB_ADMIN_USER="${SAAS_DB_ADMIN_USER:-postgres}"
SAAS_DB_ADMIN_PASSWORD="${SAAS_DB_ADMIN_PASSWORD:-postgres}"

SKIP_COMPOSE=0
RUN_GO=1
RUN_PY=1

# Parse args
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-compose)
            SKIP_COMPOSE=1
            shift
            ;;
        go)
            RUN_PY=0
            shift
            ;;
        py)
            RUN_GO=0
            shift
            ;;
        --help|-h)
            echo "Uso: $0 [--skip-compose] [go|py]"
            echo "  --skip-compose: asume stack ya levantado"
            echo "  go: ejecuta solo tests Go"
            echo "  py: ejecuta solo tests Python"
            exit 0
            ;;
        *)
            echo "Opción desconocida: $1"
            exit 1
            ;;
    esac
done

# Colores
red()    { printf '\033[31m%s\033[0m\n' "$*" >&2; }
green()  { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }

# Levantar stack si es necesario
if [ "$SKIP_COMPOSE" = 0 ]; then
    yellow "[setup] Levantando stack con docker compose..."
    cd "$PROJECT_ROOT"
    docker compose up -d --build

    yellow "[setup] Esperando healthchecks (hasta 60s)..."
    for i in {1..30}; do
        if curl -sS -o /dev/null -m 2 "$GATEWAY_URL/api/tx/health" 2>/dev/null; then
            green "[setup] Gateway respondió en $GATEWAY_URL"
            break
        fi
        sleep 2
    done
else
    yellow "[setup] --skip-compose: asumiendo servicios ya levantados"
fi

# Verificar que los servicios respondan
yellow "[check] Verificando servicios..."
check_service() {
    local name=$1
    local url=$2
    if curl -sS -o /dev/null -m 2 "$url" 2>/dev/null; then
        green "[check] ✓ $name responde"
    else
        red "[check] ✗ $name NO responde en $url"
        return 1
    fi
}

check_service "Gateway" "$GATEWAY_URL/api/tx/health" || exit 1
check_service "Backend" "$BACKEND_URL/api/health" || exit 1
check_service "Auth-Service" "$AUTHSVC_URL/auth/health" || exit 1
check_service "SaaS" "$SAAS_URL/api/saas/health" || exit 1

# Exportar variables de entorno para los tests
export GATEWAY_URL BACKEND_URL AUTHSVC_URL SAAS_URL
export SAAS_DB_HOST SAAS_DB_PORT SAAS_DB_NAME
export SAAS_DB_APP_ROLE SAAS_DB_USER SAAS_DB_PASSWORD
export SAAS_DB_ADMIN_USER SAAS_DB_ADMIN_PASSWORD

# Ejecutar tests Go
if [ "$RUN_GO" = 1 ]; then
    echo
    yellow "[tests] Ejecutando tests de integración (Go)..."
    cd "$SCRIPT_DIR"

    if command -v go >/dev/null 2>&1; then
        if go mod download 2>/dev/null; then
            if go test -v -tags=integration; then
                green "[tests] ✓ Tests Go pasaron"
            else
                red "[tests] ✗ Tests Go fallaron"
                exit 1
            fi
        else
            yellow "[tests] Go tests: go mod download falló, intentando ejecutar directamente..."
            if go test -v -tags=integration; then
                green "[tests] ✓ Tests Go pasaron"
            else
                red "[tests] ✗ Tests Go fallaron"
                exit 1
            fi
        fi
    else
        red "[tests] Go no está instalado, saltando tests Go"
    fi
fi

# Ejecutar tests Python
if [ "$RUN_PY" = 1 ]; then
    echo
    yellow "[tests] Ejecutando tests de integración (Python)..."
    cd "$SCRIPT_DIR"

    # Ambos archivos usan el venv de fastAPI_SaaS: es el único del repo con
    # psycopg2 (test de RLS) y httpx (test de auth). Se lanzan por separado para
    # que un archivo que no se pueda importar no corte la collection del otro.
    if command -v uv >/dev/null 2>&1; then
        PY_TEST=(uv run --project "$PROJECT_ROOT/fastAPI_SaaS" pytest)
    elif command -v pytest >/dev/null 2>&1; then
        # Fallback: `pytest` tiene que estar instalado y con psycopg2.
        PY_TEST=(pytest)
    else
        red "[tests] ni uv ni pytest están instalados, saltando tests Python"
        PY_TEST=()
    fi

    if [ ${#PY_TEST[@]} -gt 0 ]; then
        py_failed=0

        if "${PY_TEST[@]}" test_auth_integration.py -v; then
            green "[tests] ✓ Tests de auth pasaron"
        else
            red "[tests] ✗ Tests de auth fallaron (¿falta pyjwt? no lo declara ningún proyecto)"
            py_failed=1
        fi

        if "${PY_TEST[@]}" test_saas_rls_integration.py -v; then
            green "[tests] ✓ Tests de RLS pasaron"
        else
            red "[tests] ✗ Tests de RLS fallaron"
            py_failed=1
        fi

        if [ "$py_failed" = 1 ]; then
            red "[tests] Tests Python: hubo fallas"
            exit 1
        fi
        green "[tests] ✓ Tests Python pasaron"
    fi
fi

echo
green "=== Todos los tests de integración pasaron ==="
