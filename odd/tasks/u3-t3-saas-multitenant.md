# U3-T3 — SaaS multi-tenant funcional: users, aislamiento y billing

Cerrar las tres issues de U3-T3 (#15 CRUD de users_saas, #17 aislamiento multi-tenant, #16 planes/suscripciones + billing) dejando el servicio SaaS realmente ejecutable, aislado y probado. La issue #8 (Fiber→Chi) queda fuera por decision del operador.

El trabajo de la sesion anterior quedo **escrito pero nunca ejecutado ni commiteado**: el arbol tiene `app/api/users.py`, `app/api/billing.py`, `app/crud/user.py`, `app/middleware/tenant.py`, modelos `plan`/`subscription`, schemas, migracion Alembic `0002_rls_policies`, tests `test_users_crud.py` y cambios de migracion en el Backend. Al ejecutar la suite esa codigo **no importa**: falta `greenlet` (SQLAlchemy asyncio) y `email-validator` (`EmailStr`), `app/middleware/tenant.py` usa `Depends` sin importarlo (NameError en import), `main.py` importa `billing_router` y no lo monta, y los endpoints de users usan parametros fake `current_tenant_id: int = 1, current_rol: str = "admin"` con `TODO: Obtener del JWT` (agujero de seguridad: cualquiera crea usuarios en cualquier tenant). Ademas `docker-compose.yml` todavia tiene `context: ./Unidad3_SaaS`, ruta que ya no existe tras el rename (PR #29).

## Specs

### Identidad y aislamiento (issue #17)

- **S1** — La identidad del request llega al SaaS por los headers que el gateway copia desde el `forward_auth` al auth-service: `X-User-Id`, `X-User-Email`, `X-User-Rol`, `X-Tenant-Id`, `X-Rol-Saas` (fuente: `gateway/Caddyfile` `copy_headers` + `auth-service/internal/handlers/auth.go:144-148`). El SaaS NO vuelve a validar el JWT (decision ya tomada en GA6: el unico emisor/validador es el auth-service).
- **S2** — Un endpoint tenant-scoped sin `X-User-Id` responde **401**, nunca 200 con datos de otro tenant. Los parametros fake `current_tenant_id: int = 1` / `current_rol: str = "admin"` desaparecen de toda firma de endpoint.
- **S3** — Si `X-Tenant-Id` no coincide con el `{tenant_id}` de la ruta, la respuesta es **403** salvo que el rol SaaS sea `superadmin`.
- **S4** — Aislamiento a nivel de fila en Postgres: politicas RLS sobre `users_saas` y `subscriptions` comparando `tenant_id` contra el GUC `app.current_tenant`, que cada request setea por conexion con `SET LOCAL`.
- **S5** — El aislamiento se sostiene a nivel de codigo (filtro `tenant_id` en cada query del CRUD) **y** a nivel de base (RLS). La suite corre sobre sqlite (sin RLS) y la verificacion RLS real corre contra Postgres del stack.

### CRUD de users_saas (issue #15)

- **S6** — Endpoints bajo `/api/saas/tenants/{tenant_id}/users`:
  - `POST` inviting usuario (201), `GET` listar (200 con `{users, total}`), `GET /{user_id}`, `PATCH /{user_id}`, `DELETE /{user_id}` (204).
- **S7** — Email unico **por tenant** (`uq_users_saas_tenant_email`): el mismo email puede vivir en dos tenants distintos.
- **S8** — Solo `admin` o `superadmin` puede invitar, cambiar rol o eliminar usuarios; un `member` no.
- **S9** — Passwords hasheados con bcrypt; la API nunca devuelve el hash.

### Planes y suscripciones (issue #16)

- **S10** — `GET /api/saas/plans` lista los planes sin scoping de tenant.
- **S11** — `GET/POST/DELETE /api/saas/tenants/{tenant_id}/subscription` con aislamiento de tenant (S3).
- **S12** — `POST /api/saas/webhooks/stripe` es un **stub**: recibe el evento y devuelve `{status, event_type}`; no hay integracion real con Stripe.
- **S13** — Un tenant tiene como maximo una suscripcion (409/400 al duplicar); cancelar marca `cancel_at_period_end` en vez de borrar.

### Ejecucion y contrato compartido

- **S14** — `docker compose up -d` construye y levanta el stack: el contexto de build del servicio SaaS es `./fastAPI_SaaS` (la ruta renombrada), no `./Unidad3_SaaS`.
- **S15** — `uv sync` en `fastAPI_SaaS/` instala todo lo que el codigo importa: el extra asyncio de SQLAlchemy (`greenlet`) y el validador de `EmailStr`.
- **S16** — Todo endpoint nuevo se monta en `app/main.py`; ningun router se importa sin `include_router`.

## Tasks (work-unit commits)

- **WU1 · `fix(saas): make service importable + real tenant identity dependency`** — pyproject (`sqlalchemy[asyncio]`, `pydantic[email]`), `app/deps/principal.py` (401/403 reales), `app/middleware/tenant.py` corregido, `users.py` sin parametros fake, `main.py` monta billing. Tests API de users (401/403/201/404/204/aislamiento). Cubre S1, S2, S3, S6-S9, S15, S16.
- **WU2 · `feat(saas): RLS policies + per-request app.current_tenant`** — Alembic `0002` idempotente sobre `users_saas` + `subscriptions`, dependencia que setea `SET LOCAL app.current_tenant` en Postgres, test de la dependencia y test de aislamiento cross-tenant, script de verificacion RLS contra el stack. Cubre S4, S5.
- **WU3 · `feat(saas): plans + subscriptions + Stripe webhook stub`** — seeds de planes, endpoints de billing con la misma identidad real, tests de billing (plans, 403 cross-tenant, 404 plan inexistente, 400 suscripcion duplicada, 204 cancelar, webhook stub). Cubre S10-S13.
- **WU4 · `fix(backend): pedidos -> auth.users + migraciones en startup + compose path`** — `003_migrate_to_auth_schema.sql` (copia datos, reapunta FK, dropea `public.usuarios`), `init.sql` sin `usuarios`, `repository.go` lee `auth.users`, `Dockerfile` aplica `.sql` antes de arrancar, `docker-compose.yml` con `./fastAPI_SaaS`. Cubre S14.
- **WU5 · `docs: U3-T3 checklist + project map + auth schema migration notes`** — `docs/REQUIREMENTS_CHECKLIST.md` (U3-T3 marcado con lo verificado), `docs/PROJECT_MAP.md` y `README.md` con el nombre real del directorio, `docs/MIGRATION_AUTH_SCHEMA.md` como runbook. Cubre S14.

## Log

- **L1** (usuario, 2026-10-22): "orita se estaba resolviendo todas las issue menos la 8 puedes continuar ?" -> continuar con #15, #17, #16 (el codigo estaba escrito y sin ejecutar) y despues #18, #19, #20, #21.
- **L2** (operador): la sesion anterior escribio el codigo de U3-T3 pero no lo commiteo ni lo ejecuto; la suite falla en import (greenlet, email-validator, NameError en `middleware/tenant.py`), `billing_router` no esta montado, y los endpoints de users usan identidad fake. Este documento arranca por WU1 para volver el servicio ejecutable antes de agregar comportamiento.
- **L3** (operador): identidad del request = headers `X-User-*` que inyecta el `forward_auth` del gateway (coherente con GA6, donde el SaaS no valida JWT). Se descarta reintroducir validacion de JWT en el SaaS.
- **L4** (operador): el `docker-compose.yml` que apunta a `./Unidad3_SaaS` es una regresion del rename (PR #29) y se corrige en WU4, no se deja para despues.
- **L5** (operador): issue #8 (Fiber -> Chi) fuera de alcance por decision del usuario.