# U3-T1 — Scaffold FastAPI `Unidad3_SaaS/`

Levantar la capa SaaS/multitenant del proyecto con FastAPI moderno y bases sólidas para crecer en U3-T2/T3 sin refactor. Esta rama es el commit fundacional del servicio, no incluye endpoints de gestión todavía — solo el setup que cualquier feature nueva va a poder usar desde el día siguiente.

L1 spec (usuario): _"ya mergee el pr sigue con el u3-t1, una nueva rama desde develop"_. L1.1 (operador): _"si crear feature desde main"_ — acordado que el flujo real del proyecto mergea a main directamente, así que la rama se crea desde main, no develop (develop quedó 22 commits atrás y se sincronizará aparte). L2 (decisión de scope): _"asumir FastAPI avanzado"_ — uv + pyproject.toml, SQLAlchemy 2.x + Alembic desde día 1, multi-tenant con `tenant_id` en cada tabla, Docker Compose con Postgres separado.

## Specs

- **S1** — Carpeta `Unidad3_SaaS/` en la raíz del repo con su propio `pyproject.toml`, `.python-version`, `uv.lock`, `.env.example`, `.gitignore`, `Dockerfile`, `docker-compose.yml`, `README.md`. Independiente del `Backend/` Go/Fiber; comparte solo el `docs/`.
- **S2** — Stack: Python 3.11+, FastAPI 0.115+, uvicorn 0.32+, SQLAlchemy 2.x (async: `asyncpg`/`psycopg`), Alembic, pydantic 2.x, pydantic-settings, python-dotenv, pytest, httpx (para tests). Gestor de paquetes: `uv` (Astral).
- **S3** — Estructura de paquetes:
  ```
  Unidad3_SaaS/
  ├── pyproject.toml
  ├── README.md
  ├── Dockerfile
  ├── docker-compose.yml
  ├── .env.example
  ├── .env
  ├── alembic.ini
  ├── app/
  │   ├── main.py             # FastAPI app instance, CORS, routers, startup/shutdown
  │   ├── core/
  │   │   ├── config.py      # Pydantic Settings (env vars)
  │   │   └── db.py          # Async engine + session dep
  │   ├── api/
  │   │   └── health.py      # GET /health, GET /api/saas/health
  │   ├── models/
  │   │   ├── base.py        # SQLAlchemy DeclarativeBase + TimestampMixin
  │   │   ├── tenant.py      # Tenant(id, slug, name, created_at) — TODAS las tablas SaaS tendran tenant_id
  │   │   └── user.py        # UserSaaS(id, tenant_id, email, hashed_password, rol, created_at)
  │   └── schemas/           # Pydantic v2 para request/response (placeholder)
  ├── alembic/
  │   ├── env.py
  │   ├── script.py.mako
  │   └── versions/
  │       └── 0001_initial.py  # crea tenants + users_saas
  └── tests/
      ├── conftest.py        # cliente async + DB en sqlite in-memory para tests
      └── test_health.py     # GET /health + GET /api/saas/health
  ```
- **S4** — `Tenant.slug` único + índice. `UserSaaS.email` único por tenant (uq(tenant_id, email)). Multi-tenant preparado: cualquier tabla SaaS futura hereda `tenant_id` FK.
- **S5** — Endpoints iniciales: `GET /health` (200 `{status:"ok"}`) y `GET /api/saas/health` (200 con ping a Postgres). Sin auth en estos dos (igual que `/api/health` de Go).
- **S6** — CORS con `AllowCredentials: true` y `CORS_ORIGINS` por env (default `http://localhost:5173`). Aplica la misma convención que `Backend/main.go`.
- **S7** — Dockerfile multi-stage con uv (copia `pyproject.toml` + `uv.lock`, hace `uv sync --frozen`, construye imagen final). `docker-compose.yml` con servicio `saas-db` (postgres:16-alpine) en `:5435` para no chocar con la BD de Go en `:5434`/`:5432` del repo, y servicio `saas-api` que depende del anterior con healthcheck.
- **S8** — Tests con pytest-asyncio + httpx.AsyncClient. Al menos 3 tests: `/health` 200, `/api/saas/health` 200 con DB up, configuración carga env vars.
- **S9** — README documenta: requisitos (uv), `uv sync`, `uv run pytest`, `docker compose up`, variables de entorno en `.env.example`, decisión de "qué va en SaaS vs lo que ya está en Go/Fiber" (Go = núcleo transaccional, FastAPI = capa SaaS/billing/admin/multi-tenant).

## Tasks (work-unit commits)

- **U3-T1.1 · `chore(u3-saas): pyproject.toml + uv lockfile + estructura inicial`** — autocontenido. Crea `Unidad3_SaaS/` con `pyproject.toml`, `.python-version`, estructura `app/`, `app/main.py` con FastAPI app + CORS + health endpoint. **Verificable**: `uv sync` && `uv run uvicorn app.main:app` arranca en :8000 y devuelve 200 en `/health`. Cubre S1, S2 parcial, S3 parcial, S5 parcial, S6.

- **U3-T1.2 · `feat(u3-saas): SQLAlchemy 2.x async + multi-tenant base models`** — `app/core/db.py` (async engine + session_dep), `app/models/base.py` (DeclarativeBase + TimestampMixin), `app/models/tenant.py`, `app/models/user.py`. **Verificable**: importan sin error, `Tenant` con slug único, `UserSaaS.tenant_id` FK + uq(tenant_id, email). Cubre S3 completo, S4.

- **U3-T1.3 · `feat(u3-saas): alembic + initial migration (tenants + users_saas)`** — `alembic.ini`, `alembic/env.py` async, primera migración crea las dos tablas. **Verificable**: `uv run alembic upgrade head` aplica contra Postgres local o el del docker-compose. Cubre S3.

- **U3-T1.4 · `chore(u3-saas): Dockerfile + docker-compose + .env.example + .gitignore`** — contenedor reproducible, aislado en :5435, healthcheck. **Verificable**: `docker compose up -d saas-db saas-api` y `curl localhost:8000/api/saas/health` devuelve 200 con postgres up. Cubre S7.

- **U3-T1.5 · `test(u3-saas): pytest + /health + config tests`** — pytest-asyncio, conftest con AsyncClient + DB sqlite in-memory para tests, 3+ tests. **Verificable**: `uv run pytest` 3/3 verde. Cubre S8.

- **U3-T1.6 · `docs(u3-saas): README explaining architecture decision`** — README explica requirements, cómo correr, qué va en SaaS vs el Go/Fiber existente, próximos pasos U3-T2. Cubre S9.

- **U3-T1.7 · `docs(odd): feature doc for u3-t1-fastapi-scaffold task unit`** — registro del alcance real.

## Log

- **L1** (usuario): "ya mergee el pr sigue con el u3-t1, una nueva rama desde develop".
- **L1.1** (operador): "si crear feature desde main". Razon: develop quedo 22 commits atras de main y el flujo del proyecto mergea directo a main. Sincronizar develop a main queda fuera del scope de esta tarea (se hara por separado si el operador lo pide).
- **L2** (decision): "asumir FastAPI avanzado" — uv + pyproject.toml, SQLAlchemy 2.x + Alembic multi-tenant desde dia 1, Docker Compose con Postgres separado en :5435.
- **L3** (implementacion, pendiente): commits U3-T1.1 a U3-T1.7 en `feature/u3-t1-fastapi-scaffold` desde main.
