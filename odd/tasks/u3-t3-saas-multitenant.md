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
- **WU2 · `feat(saas): plans + subscriptions tables (migración 0002) + tests de billing`** — migración Alembic `0002_billing` (tablas `plans`/`subscriptions` + catálogo semilla free/pro/enterprise) y `tests/test_billing_api.py` (401/403/404/400/201/204, catálogo no tenant-scoped, webhook stub). Los endpoints de billing ya entraron montados y con identidad real en WU1; aquí se crea lo que les faltaba: la tabla y la prueba. Cubre S10-S13.
- **WU3 · `feat(saas): RLS tenant-scoped (migración 0003) + transacción/GUC por request`** — Alembic `0003_rls_policies` sobre `users_saas` y `subscriptions` con `FORCE ROW LEVEL SECURITY`, dependencia `get_tenant_db` que setea `app.current_tenant` por conexión y hace commit/rollback de la transacción del request, CRUD a `flush` (un commit por request), test unitario del GUC y test de integración RLS contra Postgres real. Cubre S4, S5.
  - Orden obligatorio: la migración 0003 referencia `subscriptions`, que nace en 0002. Por eso WU2 va antes que WU3.
- **WU4 · `fix(backend): pedidos -> auth.users + migraciones en startup + compose path`** — `003_migrate_to_auth_schema.sql` (copia datos, reapunta FK, dropea `public.usuarios`), `init.sql` sin `usuarios`, `repository.go` lee `auth.users`, `Dockerfile` aplica `.sql` antes de arrancar, `docker-compose.yml` con `./fastAPI_SaaS`. Cubre S14.
- **WU5 · `docs: U3-T3 checklist + project map + auth schema migration notes`** — `docs/REQUIREMENTS_CHECKLIST.md` (U3-T3 marcado con lo verificado), `docs/PROJECT_MAP.md` y `README.md` con el nombre real del directorio, `docs/MIGRATION_AUTH_SCHEMA.md` como runbook. Cubre S14.

## Log

- **L1** (usuario, 2026-10-22): "orita se estaba resolviendo todas las issue menos la 8 puedes continuar ?" -> continuar con #15, #17, #16 (el codigo estaba escrito y sin ejecutar) y despues #18, #19, #20, #21.
- **L2** (operador): la sesion anterior escribio el codigo de U3-T3 pero no lo commiteo ni lo ejecuto; la suite falla en import (greenlet, email-validator, NameError en `middleware/tenant.py`), `billing_router` no esta montado, y los endpoints de users usan identidad fake. Este documento arranca por WU1 para volver el servicio ejecutable antes de agregar comportamiento.
- **L3** (operador): identidad del request = headers `X-User-*` que inyecta el `forward_auth` del gateway (coherente con GA6, donde el SaaS no valida JWT). Se descarta reintroducir validacion de JWT en el SaaS.
- **L4** (operador): el `docker-compose.yml` que apunta a `./Unidad3_SaaS` es una regresion del rename (PR #29) y se corrige en WU4, no se deja para despues.
- **L5** (operador): issue #8 (Fiber -> Chi) fuera de alcance por decision del usuario.
- **L6** (operador, corte de slice): WU1 quedó en 1.619 líneas authored, muy por encima del slice de ~400. No es codigo nuevo decision por decision: es el trabajo de la sesion anterior (users API, billing API, crud, modelos, schemas, tests) que estaba sin commitear, mas los fixes que lo volvieron ejecutable. Dividirlo en commits mas pequenos habria dejado commits intermedios que no arrancan. WU2..WU5 se mantienen deliberadamente chicos.
- **L7** (implementación): orden WU2/WU3 invertido respecto al plan inicial. La migración RLS necesita la tabla `subscriptions`, que nace en la migración de billing: sin WU2 primero, `alembic upgrade head` falla.
- **L8** (implementación): el `.venv` local apuntaba al path viejo `/Unidad3_SaaS` (antes del rename), asi que sus scripts (`alembic`, `pytest`) no ejecutaban: `uv run alembic` daba `Failed to spawn`. Se recreo el venv con `uv sync`. Sin esto, los comandos que documenta el README del servicio no funcionan.
- **L9** (implementación): verificación WU2 contra Postgres real (`docker compose up -d db` + `alembic upgrade head` + `downgrade base` + `upgrade head`): tablas creadas, catálogo de 3 planes sembrado, sin índices duplicados, y suite completa en verde (47 tests).- **L10** (implementación): WU3, primera parte (`0003_rls_policies`, `get_tenant_db` con `SET LOCAL app.current_tenant` y transaccion por request, CRUD a `flush`, `tests/test_tenant_session.py`). El cambio de `commit()` a `flush()` es la pieza que hace que el GUC sobreviva todo el request: un commit intermedio lo borraria.
- **L11** (implementación, hallazgo): **la RLS era inerte con la configuracion de ese momento.** `postgres` en `postgres:16-alpine` es SUPERUSER, y los superusers ignoran RLS aunque la tabla tenga `FORCE ROW LEVEL SECURITY` (eso ata al owner, no al superuser). Reproducido a mano: conectado como `postgres`, el test de integracion veia las filas de los dos tenants. Sin esto habriamos的分 creado tres capas de aislamiento y dos no habrian filtrado nada.
- **L12** (implementación, delegada): el wiring del rol sin superuser (`0004_saas_app_role`, dos identidades en `config.py`, `alembic/env.py` con DSN de admin, compose con admin para migrar y rol de app para uvicorn, tests que documentan el bypass) lo escribio un subagent y fue revisado y re-verificado por el padre antes de commitear: 54 unitarios/API, 10 de integracion, `alembic upgrade head` idempotente, `downgrade base` -> `upgrade head`, y el rol con `rolsuper=f`/`rolbypassrls=f`.
- **L13** (implementación): WU6 no estaba en el plan original. El entrypoint del Backend aplica `database/migrations/*.sql` y la migracion 004 necesita `auth.users`, que crea el auth-service: sin `depends_on: auth-service` el Backend puede fallar al arrancar en frio. Se agrega a `docker-compose.yml` con su comentario explicando el por que.
- **L14** (operador, 2026-10-22): "si" a la propuesta de commitear todo, abrir PR y mergear, cerrando #15, #16 y #17. Correccion al planteo: el PR cierra 3 de las 8 issues abiertas, no "todas menos la 8". #18 (Playwright), #19 (SPA build + Caddy), #20 (TLS) y #21 (rate limit + observabilidad + deploy) siguen abiertas y sin empezar; #8 queda fuera de alcance por L5.
