# Arquitectura del stack MultiCatálogo

> Living doc de cómo están armadas las piezas y cómo se habla el stack.
> Última revisión: post-U3-T2 (octubre 2026), tras merger de GA1-GA8.

## Visión de una sola imagen

```
                         ┌──────────────────────────────────────────────────────────┐
                         │  Browser (React SPA servida por el gateway)              │
                         │  cookies: auth_token (B2C) + saas_token (SaaS)          │
                         └───────────────────────┬──────────────────────────────────┘
                                                 │ HTTP
                                                 ▼
┌────────────────────────────────────────────────────────────────────────────────────┐
│  gateway (Caddy 2) — :8080                                                        │
│                                                                                    │
│  /                 ───► SPA estatico (dist compilado en la imagen)                 │
│  /api/saas/*       ───► saas-api:8000                                              │
│  /api/*            ───► backend:3000  (login, /me, productos, red,                 │
│                                      checkout; el Backend ya sirve /api)           │
│                                                                                    │
│  Sin forward_auth (publicas): /api/login, /api/logout, /api/productos*,            │
│  /api/saas/auth/*, /api/saas/health, /api/saas/webhooks/stripe                     │
│                                                                                    │
│  forward_auth auth-service:3001 { uri /auth/validate, copy_headers X-User-* }       │
│  reverse_proxy X-User-* upstream (viaje transparente al backend/SaaS)               │
└──────────────┬──────────────────┬──────────────────────┬─────────────────────────────┘
               │                  │                      │
   ┌───────────▼────────┐ ┌────────▼────────────┐ ┌───────▼──────────────────┐
   │  backend          │ │  saas-api           │ │  auth-service             │
   │  Go/Fiber :3000   │ │  FastAPI :8000      │ │  Go/Fiber :3001           │
   │  Login B2C        │ │  Multi-tenant       │ │  Emite/valida JWTs        │
   │  Catalogo publico │ │  Tenants, billing   │ │  schema 'auth'           │
   │  Checkout ACID    │ │                     │ │                           │
   │  Red MLM          │ │                     │ │                           │
   │  Http-proxy       │ │  Http-proxy         │ │  NO firma JWT para SaaS   │
   │  ─► auth-service  │ │  ─► auth-service    │ │  directo (futuro); hoy:    │
   └────────┬──────────┘ └─────────┬────────────┘ │  Frontend habla al         │
            │                     │              │  gateway que valida.       │
            └──────────┬──────────┘              └─────────────┬───────────────┘
                       │                                       │
                       └─────────┬──────────────────────┬──────┘
                                 │                      │
                  ┌──────────────▼──────────┐  ┌──────────▼──────────┐
                  │  Postgres 16            │  │  Redis 7            │
                  │  :5432 (interno)         │  │  :6379               │
                  │                          │  │  Cache-Aside B2C     │
                  │  Schemas:                │  │                     │
                  │   - auth (auth-service)  │  └─────────────────────┘
                  │   - backend (backend)    │
                  │   - saas (saas-api)      │
                  │                          │
                  │  Multitenant con         │
                  │  tenant_id (a partir     │
                  │  de U3-T3, schema saas)  │
                  └──────────────────────────┘
```

## Reglas clave

### 1. Auth centralizada vive en el gateway (no en cada servicio)

Caddy hace `forward_auth auth-service:3001` con `uri /auth/validate`. Para cada request a `/api/*` o `/api/saas/*` (salvo las rutas públicas de arriba):

1. Caddy extrae token (cookie `auth_token` o header `Authorization: Bearer`).
2. Hace sub-request POST a `auth-service:3001/auth/validate`.
3. Si 200, copia headers `X-User-Id`, `X-User-Email`, `X-User-Rol`, `X-Tenant-Id`, `X-Rol-Saas` al upstream.
4. Si 401, rechaza directo al cliente.

Defense in depth: cada servicio backend (Backend, SaaS) tiene su propio `RequireAuth` que **re-verifica el token localmente** si llega a sus endpoints. Útil si el gateway cae o si un request bypasea.

### 2. BD única con schemas separados

Una sola instancia Postgres. Cada servicio **posee** un schema y escribe solo ahí:

| Schema | Owner | Tablas hoy |
|---|---|---|
| `auth` | auth-service | `users` (con seed bcrypt) |
| `backend` | Backend Go/Fiber | `usuarios`, `productos`, `referidos`, `pedidos`, `pedido_items` |
| `saas` | SaaS | (futuras: `tenants`, `users_saas`, `plans`, `subscriptions`) |

Los seeds actuales viven en `auth.users` (auth-service) y `public.*` (Backend init.sql). Las tablas `pedidos`/`pedido_items` quedan en `public` históricamente.

Las FK cross-schema funcionan (Postgres las soporta dentro de la misma BD). Las vistas son la manera limpia de exponer un schema a otro cuando hace falta lectura cruzada.

### 3. JWT firma/verifica vive solo en auth-service

- Backend emite HTTP a `auth-service:3001/auth/login` (proxy en GA5) — ya no firma JWTs localmente.
- SaaS hace lo mismo (proxy en GA6).
- Middleware `RequireAuth` del backend y del SaaS siguen corriendo pero **solo validan** JWTs (no emiten). Comparten el mismo `JWT_SECRET` env (o `AUTH_JWT_SECRET` + fallback).

### 4. Componentes por servicio

#### `gateway/` (Caddy 2)

- Caddyfile único.
- Dockerfile multi-stage `caddy:2`.
- 4 Servicios Docker en su `docker-compose.yml` adyacente (no confundir con el raíz): `gateway`, `auth-service`, `backend`, `saas-api`, `redis`, `db`.

#### `auth-service/` (Go/Fiber)

- Go 1.26, Fiber 2.x, golang-jwt/v5, bcrypt, pgx.
- `migrations/0001_init.sql` crea schema `auth` + tabla `users` con seed bcrypt.
- `scripts/entrypoint.sh` aplica migraciones antes del binario (espera activa a la BD).
- Endpoints: `/auth/login`, `/auth/logout`, `/auth/refresh`, `/auth/validate`, `/auth/me`, `/auth/health`.
- Cookies: `auth_token` (B2C) y `saas_token` (SaaS). Ambas httpOnly + SameSite=Lax + Secure=(env!=dev).

#### `Backend/` (Go/Fiber B2C)

- Go 1.26, Fiber 2.x, pgx, golang-jwt/v5.
- Endpoints: `/api/{login,logout,me,productos,productos/:id,red,checkout,health}`.
- Login/Logout/Me son **proxies HTTP** al auth-service (GA5).
- El resto sigue usando `database_repository` con SQL real y `redis_cache` para Cache-Aside.

#### `fastAPI_SaaS/` (FastAPI SaaS)

- Python 3.11, FastAPI 0.115+, uvicorn, SQLAlchemy 2.x async, Alembic, httpx, pydantic v2, bcrypt.
- Endpoints:
  - `/health`, `/api/saas/health`
  - `/api/saas/auth/{login,logout,refresh,me}` — proxies HTTP al auth-service
  - `/api/saas/tenants/{tenant_id}/users` (CRUD) y `/api/saas/tenants/{tenant_id}/subscription`
  - `/api/saas/plans`, `/api/saas/webhooks/stripe` (stub)
- Identidad: **no** revalida el JWT. Lee los headers que el gateway copia desde
  `forward_auth` (`X-User-Id`, `X-Email`, `X-Rol`, `X-Tenant-Id`, `X-Rol-Saas`)
  y aplica `require_tenant_scope` / `require_tenant_admin`
  (`app/deps/principal.py`).
- Aislamiento multi-tenant en tres capas:
  1. filtro `WHERE tenant_id = :tenant` en cada query del CRUD;
  2. guards de scope/rol por endpoint (401/403) antes de tocar la BD;
  3. RLS en Postgres (`0003_rls_policies`) con `FORCE ROW LEVEL SECURITY`, que
     compara contra el GUC `app.current_tenant` que cada request setea con
     `SET LOCAL`. Requiere que el servicio conecte con un rol **no superuser**
     (migración `0004_saas_app_role`): un superuser ignora RLS.
- Modelos en `app/models/`: `base.py` (con `TimestampMixin`), `tenant.py`,
  `user.py`, `plan.py`, `subscription.py`.
- Migraciones Alembic: `0001_initial` (tenants + users_saas), `0002_billing`
  (plans + subscriptions + catálogo semilla), `0003_rls_policies`,
  `0004_saas_app_role`.

### 5. Flujo de un request típico

Ejemplo: `GET /api/red` con cookie `auth_token` válida:

1. SPA hace `fetch('/api/red', { credentials: 'include' })`. La ruta es **relativa**: el browser la resuelve contra el origen del SPA, que es el mismo `:8080` que sirve la API. Por eso no hay CORS y la cookie host-only `auth_token` se emite y vuelve en el mismo host.
2. Caddy recibe en `:8080`. Como el path matchea `/api/*` (y no `/api/saas/*`), sabe que va al backend.
3. Caddy extrae la cookie, hace sub-request POST a `auth-service:3001/auth/validate` con la cookie.
4. auth-service verifica JWT (HS256 con `AUTH_JWT_SECRET`), responde 200 y setea headers `X-User-Id:1`, `X-User-Email:admin@...`, `X-User-Rol:admin`.
5. Caddy copia esos headers y hace `reverse_proxy /api/red http://backend:3000`.
6. Backend recibe el request con los headers `X-User-*`. El middleware `RequireAuth` **defensa en profundidad** también verifica (lee el token de la cookie, lo verifica con el mismo secret, y poblaría `c.Locals`).
7. Backend responde con el árbol MLM.
8. Caddy devuelve la respuesta al cliente.

Si la cookie es inválida o ausente: el sub-request a `/auth/validate` da 401, Caddy responde 401 directo, el request nunca llega al backend.

### 6. Variables de entorno por servicio

| Servicio | Vars clave |
|---|---|
| gateway | (sin vars, lee config desde compose) |
| frontend (solo profile `dev`) | `VITE_DEV_API_TARGET`: target del proxy de dev (en compose es `http://gateway:8080`) |
| auth-service | `APP_ENV`, `AUTH_JWT_SECRET`, `AUTH_DB_*`, `CORS_ORIGINS` |
| backend | `APP_ENV`, `JWT_SECRET`/`AUTH_JWT_SECRET`, `BACKEND_DB_*`, `REDIS_*`, `AUTH_SERVICE_URL`, `CORS_ORIGINS` |
| saas-api | `APP_ENV`, `AUTH_SERVICE_URL`, `SAAS_DB_*`, `CORS_ORIGINS` |
| db | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` |

`AUTH_JWT_SECRET` debe ser **el mismo valor** en auth-service y en backend/saas para que sus JWTs sean compatibles. Si Backend usa `AUTH_JWT_SECRET=foo` y auth-service usa `AUTH_JWT_SECRET=foo`, todo funciona; si difieren, cada servicio valida solo los suyos.

### 7. ¿Cómo se levanta todo?

```bash
cp .env.example .env                  # AJUSTAR AUTH_JWT_SECRET a algo fuerte en prod
docker compose up -d --build
# Esperar ~30s para healthchecks
bash tests/e2e/gateway_e2e.sh

# El SPA ya lo sirve el gateway en http://localhost:8080 (el dist se compila
# dentro de la imagen). Para desarrollo con HMR:
docker compose --profile dev up -d frontend    # SPA en http://localhost:5173
```

Nota de entorno: si el puerto 5432 del host ya esta ocupado, `docker compose up` falla
con `Bind for 0.0.0.0:5432 failed`. Se levanta con `DB_PORT=5433 docker compose up -d`,
que solo cambia el puerto del host; dentro de la red de compose el Postgres sigue en 5432.

Para dev nativo (sin Docker): arrancar cada servicio individualmente con su `main.go` / `uvicorn`, ajustar `AUTH_SERVICE_URL` a `http://localhost:3001`.

### 8. Decisiones arquitectónicas clave (L4-L7)

- **L4** (operador): descartar full microservicios por superficie excesiva; extraer solo `auth-service`.
- **L5** (operador): stack Go/Fiber para el auth-service (consistente con Backend).
- **L6** (operador): BD única compartida con schemas separados — pragmatic, terminable en semestre.
- **L7** (operador): el mismo JWT_SECRET se usa para los dos tipos de tokens.

Issues abiertas para trabajo futuro en `docs/REQUIREMENTS_CHECKLIST.md` y los issues #15-#21.
