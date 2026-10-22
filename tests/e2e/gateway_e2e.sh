#!/usr/bin/env bash
#
# Tests E2E del stack MultiCatalogo contra el gateway Caddy unificado.
#
# Cubre GA7 del feature doc odd/tasks/u3-t2-api-gateway.md (E2E1). Cada caso
# ejecuta un curl real contra localhost:8080 (gateway) y verifica codigo +
# shape esperado.
#
# Pre-requisitos:
#   1. docker compose up -d --build desde la raiz del repo.
#   2. Esperar ~10 segundos para que healthchecks pasen (gateway espera
#      auth-service/backend/saas-api healthy).
#
# Uso:
#   bash tests/e2e/gateway_e2e.sh                    # corre toda la suite
#   bash tests/e2e/gateway_e2e.sh --case health      # corre solo ese caso
#   bash tests/e2e/gateway_e2e.sh --skip-compose     # asume stack ya arriba
#   bash tests/e2e/gateway_e2e.sh --verbose          # imprime bodies
#
# Exit code: 0 si todos los casos pasan, 1 si alguno falla.

set -uo pipefail

# --- Config ---------------------------------------------------------------
GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
BACKEND_URL="${BACKEND_URL:-http://localhost:3000}"
AUTHSVC_URL="${AUTHSVC_URL:-http://localhost:3001}"
SAAS_URL="${SAAS_URL:-http://localhost:8000}"

VERBOSE=0
SKIP_COMPOSE=0
CASES=()

# --- Args ------------------------------------------------------------------
while [ $# -gt 0 ]; do
    case "$1" in
        --case)        CASES+=("$2"); shift 2 ;;
        --verbose)     VERBOSE=1; shift ;;
        --skip-compose) SKIP_COMPOSE=1; shift ;;
        --help|-h)      sed -n '2,12p' "$0"; exit 0 ;;
        *)              echo "flag desconocido: $1" >&2; exit 2 ;;
    esac
done

# --- Helpers ---------------------------------------------------------------
red()    { printf '\033[31m%s\033[0m\n' "$*" >&2; }
green()  { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }

PASS_COUNT=0
FAIL_COUNT=0
CURRENT_CASE=""

# Muestra detalle solo en --verbose.
_run_curl() {
    if [ "$VERBOSE" = 1 ]; then
        echo "+ curl $*"
    fi
    curl -sS -o /tmp/last_body -w '%{http_code}|%{header_json}' "$@"
}

# expect_status <expected_code> <description>
expect_status() {
    local expected="$1"
    local desc="$2"
    local actual="$3"
    if [ "$actual" = "$expected" ]; then
        green "  PASS [$CURRENT_CASE] $desc  status=$actual"
        PASS_COUNT=$((PASS_COUNT+1))
    else
        red   "  FAIL [$CURRENT_CASE] $desc  expected=$expected got=$actual"
        FAIL_COUNT=$((FAIL_COUNT+1))
    fi
}

expect_body_contains() {
    local needle="$1"
    local desc="$2"
    local haystack
    haystack=$(cat /tmp/last_body 2>/dev/null || echo "")
    if printf '%s' "$haystack" | grep -q -F -- "$needle"; then
        green "  PASS [$CURRENT_CASE] $desc  contiene '$needle'"
        PASS_COUNT=$((PASS_COUNT+1))
    else
        red   "  FAIL [$CURRENT_CASE] $desc  NO contiene '$needle' en body:"
        printf '         %s\n' "$haystack" >&2
        FAIL_COUNT=$((FAIL_COUNT+1))
    fi
}

get_status() {
    # Extrae el codigo HTTP de la salida %{http_code}|{header_json} del -w.
    echo "$1" | cut -d'|' -f1
}

save_cookies() {
    local outfile="$1"
    : > "$outfile"
    while IFS= read -r line; do
        case "$line" in
            *Set-Cookie:*)
                # Parsea el Set-Cookie y guarda en formato curl.
                cookie=$(printf '%s' "${line#*:}" | sed 's/^ *//')
                name=$(printf '%s' "$cookie" | cut -d= -f1)
                val=$(printf '%s' "$cookie" | cut -d= -f2- | cut -d';' -f1)
                printf '%s\t%s\n' "$name" "$val" >> "$outfile"
                ;;
        esac
    done < <(echo "$2" | tr ',' '\n')
}

# --- Compose opcional ------------------------------------------------------
if [ "$SKIP_COMPOSE" = 0 ]; then
    if command -v docker-compose >/dev/null 2>&1 || docker compose version >/dev/null 2>&1; then
        yellow "[setup] subiendo stack via docker compose..."
        ( cd "$(git -C "$(dirname "$0")/../.." rev-parse --show-toplevel 2>/dev/null || echo .)" && docker compose up -d --build )
        # Esperar healthchecks (hasta 60s).
        for i in $(seq 1 30); do
            if curl -sS -o /dev/null -m 2 "$GATEWAY_URL/api/tx/health" 2>/dev/null; then
                green "[setup] gateway respondio en $GATEWAY_URL"
                break
            fi
            sleep 2
        done
    else
        yellow "[setup] docker compose no disponible; asume stack ya levantado."
    fi
else
    yellow "[setup] --skip-compose: asumo servicios ya levantados."
fi

# --- Casos -----------------------------------------------------------------

run_health_gateway() {
    CURRENT_CASE=health-gateway
    out=$(_run_curl -m 5 "$GATEWAY_URL/api/tx/health")
    expect_status 200 "gateway -> backend -> /api/tx/health" "$(get_status "$out")"
    expect_body_contains '"status":"ok"' "health body contiene status:ok"
}

run_health_saas() {
    CURRENT_CASE=health-saas
    out=$(_run_curl -m 5 "$GATEWAY_URL/api/saas/health")
    expect_status 200 "gateway -> saas-api -> /api/saas/health" "$(get_status "$out")"
    expect_body_contains '"service":"saas"' "saas health body contiene service:saas"
}

run_products_public() {
    CURRENT_CASE=products-public
    out=$(_run_curl -m 5 "$GATEWAY_URL/api/tx/productos")
    expect_status 200 "/api/tx/productos sin cookie" "$(get_status "$out")"
    # Body debe ser array JSON con al menos 1 producto.
    if jq -e 'type == "array" and length > 0' /tmp/last_body >/dev/null 2>&1; then
        green "  PASS [$CURRENT_CASE] /api/tx/productos devuelve array con elementos"
        PASS_COUNT=$((PASS_COUNT+1))
    else
        red "  FAIL [$CURRENT_CASE] /api/tx/productos no devuelve array con elementos"
        FAIL_COUNT=$((FAIL_COUNT+1))
    fi
}

run_red_unauth() {
    CURRENT_CASE=red-unauth
    out=$(_run_curl -m 5 "$GATEWAY_URL/api/tx/red")
    expect_status 401 "/api/tx/red sin cookie -> 401" "$(get_status "$out")"
}

run_login_admin() {
    CURRENT_CASE=login-admin
    out=$(_run_curl -m 5 -X POST "$GATEWAY_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -c /tmp/cookies.txt \
        -d '{"email":"admin@upse.edu.ec","password":"123456"}')
    expect_status 200 "login admin (gateway -> auth-service)" "$(get_status "$out")"
    expect_body_contains '"rol":"admin"' "login body contiene rol:admin"
    # La cookie auth_token debe haberse guardado.
    if grep -q '^auth_token' /tmp/cookies.txt; then
        green "  PASS [$CURRENT_CASE] cookie auth_token guardada en /tmp/cookies.txt"
        PASS_COUNT=$((PASS_COUNT+1))
    else
        red "  FAIL [$CURRENT_CASE] cookie auth_token NO guardada"
        FAIL_COUNT=$((FAIL_COUNT+1))
    fi
}

run_login_bad_pwd() {
    CURRENT_CASE=login-bad-password
    out=$(_run_curl -m 5 -X POST "$GATEWAY_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"admin@upse.edu.ec","password":"wrong"}')
    expect_status 401 "login bad password -> 401" "$(get_status "$out")"
    expect_body_contains '"error":"Credenciales incorrectas"' "401 body generico (no leak)"
}

run_me_authenticated() {
    CURRENT_CASE=me-authenticated
    out=$(_run_curl -m 5 -b /tmp/cookies.txt "$GATEWAY_URL/api/auth/me")
    expect_status 200 "me con cookie devuelve identidad" "$(get_status "$out")"
    expect_body_contains '"email":"admin@upse.edu.ec"' "me body contiene email"
}

run_me_no_cookie() {
    CURRENT_CASE=me-no-cookie
    out=$(_run_curl -m 5 "$GATEWAY_URL/api/auth/me")
    expect_status 401 "me sin cookie -> 401" "$(get_status "$out")"
}

run_saas_login_admin() {
    CURRENT_CASE=saas-login-admin
    # /api/saas/auth/login proxy-ea al auth-service via SaaS FastAPI.
    out=$(_run_curl -m 5 -X POST "$GATEWAY_URL/api/saas/auth/login" \
        -H "Content-Type: application/json" \
        -c /tmp/saas_cookies.txt \
        -d '{"email":"admin@upse.edu.ec","password":"123456"}')
    expect_status 200 "saas login admin proxy-via-saaas" "$(get_status "$out")"
    expect_body_contains '"rol":"admin"' "saas login body contiene rol:admin"
    if grep -q '^auth_token' /tmp/saas_cookies.txt; then
        green "  PASS [$CURRENT_CASE] cookie auth_token guardada via SaaS proxy"
        PASS_COUNT=$((PASS_COUNT+1))
    else
        red "  FAIL [$CURRENT_CASE] cookie auth_token NO guardada via SaaS proxy"
        FAIL_COUNT=$((FAIL_COUNT+1))
    fi
}

run_404_no_such_path() {
    CURRENT_CASE=path-404
    out=$(_run_curl -m 5 "$GATEWAY_URL/api/tx/no-existe")
    expect_status 404 "path no existente -> 404" "$(get_status "$out")"
}

# --- Orquestacion ----------------------------------------------------------

ALL_CASES=(
    health_gateway
    health_saas
    products_public
    red_unauth
    login_admin
    login_bad_pwd
    me_authenticated
    me_no_cookie
    saas_login_admin
    path_404
)

if [ "${#CASES[@]}" = 0 ]; then
    for c in "${ALL_CASES[@]}"; do
        "run_${c}"
    done
else
    for c in "${CASES[@]}"; do
        if declare -f "run_${c}" >/dev/null; then
            "run_${c}"
        else
            red "caso desconocido: $c"
        fi
    done
fi

# --- Reporte ---------------------------------------------------------------

echo
echo "==="
echo "Tests: $((PASS_COUNT + FAIL_COUNT))   PASS=$PASS_COUNT   FAIL=$FAIL_COUNT"
[ "$FAIL_COUNT" = 0 ] && green "TODOS PASARON" || red "HAY FALLOS"

exit "$FAIL_COUNT"
