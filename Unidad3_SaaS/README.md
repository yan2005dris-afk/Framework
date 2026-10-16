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

# Necesitas Postgres corriendo en :5435 (o ajusta .env)
uv run alembic upgrade head       # crea el esquema (tenants, users_saas)
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

`database_url` y `database_url_sync` se construyen desde los 4 `SAAS_DB_*` para tener una sola fuente de verdad. Cambia el password de la BD en un solo lugar.

## Estructura

```
Unidad3_SaaS/
├── pyproject.toml          # uv-managed
├── README.md
├── Dockerfile              # uv-based multi-stage
├── docker-compose.yml      # saas-db + saas-api
├── .env.example            # contrato
├── alembic.ini
├── app/
│   ├── main.py             # FastAPI app, CORS, router mount
│   ├── core/
│   │   ├── config.py       # Pydantic Settings
│   │   └── db.py           # async engine + get_db
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
│       └── 0001_initial.py
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
