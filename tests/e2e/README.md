# Tests E2E del stack MultiCatálogo

`gateway_e2e.sh` ejecuta una serie de casos vía `curl` contra el gateway Caddy en `localhost:8080`. Cubre GA7 (tests E2E) del feature doc `odd/tasks/u3-t2-api-gateway.md`.

## Pre-requisitos

1. **Stack levantado**: `docker compose up -d --build` desde la raíz del repo. Esperá a que los healthchecks pasen (~10-30 segundos según la primera build).
2. **Credenciales seed**: el `auth_db` debe tener `admin@upse.edu.ec` y `cliente@upse.edu.ec` con bcrypt. El entrypoint del auth-service corre las migraciones automáticamente al primer arranque.
3. **`curl` + `jq` instalados** en el host donde corras el script.
4. **Bash 4+** (el script usa arrays asociativos implícitos).

## Uso rápido

```bash
# Levantar stack
docker compose up -d --build

# Esperar ~15s para healthchecks
sleep 15

# Correr suite completa
bash tests/e2e/gateway_e2e.sh

# Correr un solo caso
bash tests/e2e/gateway_e2e.sh --case login-admin

# Ver bodies de las requests (verbose)
bash tests/e2e/gateway_e2e.sh --verbose

# Asumir stack ya levantado (no ejecuta `docker compose up`)
bash tests/e2e/gateway_e2e.sh --skip-compose
```

## Casos cubiertos

| Caso | Path | Verifica |
|---|---|---|
| `health-gateway` | `GET /api/tx/health` | Gateway → backend → `/api/tx/health`. Status 200 + body con `status:"ok"`. |
| `health-saas` | `GET /api/saas/health` | Gateway → saas-api. Status 200 + `service:"saas"`. |
| `products-public` | `GET /api/tx/productos` | Catálogo público (sin auth). Status 200 + array con elementos. |
| `red-unauth` | `GET /api/tx/red` | Sin cookie → 401 (forward_auth del gateway). |
| `login-admin` | `POST /api/auth/login` | Credenciales válidas → 200 + cookie `auth_token` + body con `rol:"admin"`. |
| `login-bad-password` | `POST /api/auth/login` | Credenciales inválidas → 401 + mensaje genérico. |
| `me-authenticated` | `GET /api/auth/me` con cookie | 200 + identidad. |
| `me-no-cookie` | `GET /api/auth/me` sin cookie | 401. |
| `saas-login-admin` | `POST /api/saas/auth/login` | El SaaS proxy-ea al auth-service, devuelve 200 + cookie `auth_token`. |
| `path-404` | `GET /api/tx/no-existe` | Path desconocido → 404. |

## Casos pendientes (U3-T3 los habilita)

- `/api/saas/tenants` con `saas_token` `rol_saas=superadmin` → 200 con lista. Requiere el CRUD de tenants (issue #15).
- Cross-tenant: dos tenants, user de A no ve datos de B (issue #17).
- Login SaaS con `tenant_id` claim (issue #16, billing).
- Tests E2E con Playwright (issue #18) — esta suite bash cubre la capa HTTP; Playwright cubriría la UI.

## Exit code

`0` si todos los casos pasan, `1` si alguno falla (compatible con CI).

## Variables de entorno

| Var | Default | Uso |
|---|---|---|
| `GATEWAY_URL` | `http://localhost:8080` | URL base del gateway Caddy. |
| `BACKEND_URL` | `http://localhost:3000` | Para futuros checks directos al Backend (actualmente no usado). |
| `AUTHSVC_URL` | `http://localhost:3001` | Para futuros checks directos (actualmente no usado). |
| `SAAS_URL` | `http://localhost:8000` | Para futuros checks directos (actualmente no usado). |

## Limitaciones

- Solo verifica la capa HTTP del gateway; no prueba la UI React (eso es issue #18 con Playwright).
- Asume que la BD `multicatalogo` tiene los seeds bcrypt (el auth-service los crea en el primer arranque vía entrypoint). Si los volumes vienen de una versión vieja, recrear con `docker compose down -v`.
- No captura automáticamente screenshots ni logs detallados — para debugging leer `/tmp/last_body` después de un fallo.

## Si los tests fallan

1. **401 inesperado en un endpoint que debería pasar**: revisar que el `forward_auth` del Caddyfile apunta a `/auth/validate` y que ese endpoint setea los headers correctos.
2. **Connection refused en gateway**: probablemente los healthchecks de compose aún no pasaron. Esperar más.
3. **502 en `/api/saas/auth/*`**: el gateway → saas-api funciona, pero saas-api no logra alcanzar `auth-service:3001` (DNS Docker). Verificar que `saas-api` está en la misma red Docker que `auth-service`.
4. **500 inesperado**: ver `docker compose logs gateway auth-service backend saas-api`.
