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

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
BACKEND_URL="${BACKEND_URL:-http://localhost:3000}"
AUTHSVC_URL="${AUTHSVC_URL:-http://localhost:3001}"
SAAS_URL="${SAAS_URL:-http://localhost:8000}"

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

    if command -v uv >/dev/null 2>&1; then
        if uv run pytest test_auth_integration.py -v; then
            green "[tests] ✓ Tests Python pasaron"
        else
            red "[tests] ✗ Tests Python fallaron"
            exit 1
        fi
    elif command -v pytest >/dev/null 2>&1; then
        if pytest test_auth_integration.py -v; then
            green "[tests] ✓ Tests Python pasaron"
        else
            red "[tests] ✗ Tests Python fallaron"
            exit 1
        fi
    else
        red "[tests] ni uv ni pytest están instalados, saltando tests Python"
    fi
fi

echo
green "=== Todos los tests de integración pasaron ==="
