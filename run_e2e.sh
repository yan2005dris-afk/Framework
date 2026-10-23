#!/usr/bin/env bash
set -e

export BASE_URL="${BASE_URL:-https://framework.yandro.tech}"
export CHROME_BIN="${CHROME_BIN:-/usr/bin/chromium}"
export NODE_OPTIONS="--experimental-strip-types"

if [ ! -d "node_modules" ]; then
    ln -s Frontend/node_modules node_modules
fi

echo "=========================================================="
echo " Ejecutando Suite E2E Completa (Cucumber + Playwright)"
echo " Target: $BASE_URL"
echo " Navegador: $CHROME_BIN"
echo " Features: $*"
echo "=========================================================="

if [ "$#" -gt 0 ]; then
    npx cucumber-js "$@"
else
    npx cucumber-js tests/e2e/features/admin_operaciones.feature --import tests/e2e/steps/admin_operaciones_steps.ts
fi
