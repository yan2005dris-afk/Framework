# MultiCatálogo SaaS

Capa SaaS / multitenant del proyecto MultiCatálogo. FastAPI sobre SQLAlchemy 2 (async) + Alembic.

Vecina al backend Go/Fiber en `Backend/` (núcleo transaccional). Esta capa agrega capacidades SaaS: multi-tenant, billing, admin.

## Qué va aquí vs qué va en Backend/

La regla de partición entre los dos servicios:

| Concern | Backend (Go/Fiber) | SaaS (FastAPI) |
|---|---|---|
| Checkout transaccional ACID | ✅ (PG + Redis) | — |
| Catálogo de productos (catalogo público) | ✅ (`GET /api/productos`) | — |
| Red multinivel MLM | ✅ (`GET /api/red`, ya con auth) | — |
| Auth de usuarios finales (bcrypt + JWT HS256 + cookie httpOnly) | ✅ | — |
| Multi-tenancy (aislamiento por `tenant_id` para B2B) | — | ✅ |
| Catálogo SaaS (planes, features por tenant) | — | ✅ |
| Billing / suscripciones | — | ✅ |
| Admin de tenants (crear, invitar usuarios, suspender) | — | ✅ |
| APIs internas administrativas (no expone la SPA de cara) | — | ✅ |

Ambas capas pueden evolucionar independientemente. La SPA React consume hoy el Backend; cuando la U3 tenga endpoints `/api/saas/*` consumibles, el proxy de U4 (Caddy/Nginx) enrutará `/api/saas/*` hacia FastAPI y `/api/tx/*` hacia Go/Fiber.

## Requisitos

- Python 3.11+
- [uv](https://docs.astral.sh/uv/) (gestor de paquetes Astral, single binary, ~30MB)
- PostgreSQL 16+ (local o vía Docker)

## Setup local sin Docker

```bash
uv sync                          # crea .venv/ + instala deps
cp .env.example .env             # editar SAAS_DB_PASSWORD etc. segun entorno

# Necesitas Postgres corriendo en :5435 (o ajusta .env).
# `alembic upgrade head` usa las credenciales de admin (SAAS_DB_ADMIN_*), que
# son las unicas con DDL; ademas crea y sincroniza el rol `saas_app` con el que
# corre el servicio.
uv run alembic upgrade head
uv run uvicorn app.main:app --reload --port 8000
```

Abre `http://localhost:8000/docs` para Swagger UI.

## Setup local con Docker Compose

```bash
docker compose up --build        # arranca saas-db y saas-api
docker compose logs -f saas-api
docker compose down -v           # baja todo y borra el volumen (reset)
```

Cuando arranca, `saas-api` corre `alembic upgrade head` antes de uvicorn, así que el esquema queda migrado.

## Tests

```bash
uv run pytest                # tests sync + async (3 smoke tests de smoke)
uv run pytest -v             # verboso
uv run pytest --cov=app     # coverage
```

## Variables de entorno

Ver `.env.example`. Convenciones:

- **Prefijo `SAAS_`** para todo dato sensible: hosts, puertos, credenciales de BD.
- **Sin prefijo** para metadatos públicos: `APP_ENV`, `CORS_ORIGINS`, `API_TITLE`.

La BD se accede con **dos identidades**, y mezclarlas rompe el aislamiento:

| Identidad | Settings | Variables de entorno | Para qué |
|---|---|---|---|
| **App** (runtime) | `saas_db_user`, `saas_db_password` | `SAAS_DB_USER`, `SAAS_DB_PASSWORD` | Corre uvicorn. Rol **sin superuser** (default `saas_app`), que es lo que hace que las políticas RLS de la migración `0003` filtren. |
| **App** (nombre del rol) | `saas_db_app_role` | `SAAS_DB_APP_ROLE` | Nombre del rol que crea y permiso la migración `0004_saas_app_role`. |
| **Admin** | `saas_db_admin_user`, `saas_db_admin_password` | `SAAS_DB_ADMIN_USER`, `SAAS_DB_ADMIN_PASSWORD` | Solo Alembic: las migraciones hacen DDL (tablas, rol, políticas) y el rol de la app no puede ni debe poder. |

`database_url` y `database_url_sync` se construyen desde los `SAAS_DB_*` de runtime;
`migration_database_url_sync` es el mismo esquema con las credenciales de admin, y
es la que usa `alembic/env.py`.

**Por que importa**: un superuser de Postgres ignora las políticas RLS por
completo (`FORCE ROW LEVEL SECURITY` ata al owner de la tabla, no al superuser).
Si uvicorn conectara como `postgres`, cada request vería las filas de todos los
tenants. Los tests `tests/integration/test_saas_rls_integration.py` lo cubren en
los dos sentidos: con el rol de la app las políticas filtran, y con el superuser
no filtran.

## Estructura

```
fastAPI_SaaS/
├── pyproject.toml          # uv-managed
├── README.md
├── Dockerfile              # uv-based multi-stage
├── docker-compose.yml      # saas-db + saas-api
├── .env.example            # contrato
├── alembic.ini
├── app/
│   ├── main.py             # FastAPI app, CORS, router mount
│   ├── core/
│   │   ├── config.py       # Pydantic Settings (runtime + admin)
│   │   └── db.py           # async engine + get_db + get_tenant_db (GUC)
│   ├── api/
│   │   └── health.py
│   ├── models/
│   │   ├── base.py         # DeclarativeBase + TimestampMixin
│   │   ├── tenant.py
│   │   └── user.py
│   └── schemas/            # Pydantic v2 DTOs (placeholder)
├── alembic/
│   ├── env.py
│   ├── script.py.mako
│   └── versions/
│       ├── 0001_initial.py
│       ├── 0002_billing.py
│       ├── 0003_rls_policies.py
│       └── 0004_saas_app_role.py
└── tests/
    ├── conftest.py
    └── test_health.py
```

## Convenciones

- **Multi-tenant**: cualquier tabla SaaS (excepto `tenants` misma) tiene `tenant_id` FK → `tenants.id`. U3-T2 agrega la capa de auth + tenant claim en JWT y filtros `WHERE tenant_id = :current_tenant` en cada query.
- **Async-first**: SQLAlchemy 2.x con `asyncpg`. Los endpoints usan `Depends(get_db)` para obtener una `AsyncSession` por request.
- **No JWT/login todavía**: U3-T2 lo agrega. El middleware de auth del Backend Go/Fiber se inspira en esto pero es independiente; podemos elegir: SaaS usa el mismo JWT firmado por Go, o SaaS firma su propio token con la misma `JWT_SECRET`. Decisión pendiente.

## Próximos pasos

- **U3-T2**: auth + tenant claim + CRUD de tenants (admin-only).
- **U3-T3**: API de billing (planes, suscripción, webhooks de pago).
- **U3-T4+**: features específicas que definan los temas del sílabo siguientes.
- **U4**: Caddy/Nginx como proxy reverso unificando Backend :3000 y SaaS :8000 bajo `/api/tx/*` y `/api/saas/*`.
