# MultiCatálogo SaaS

Capa SaaS / multitenant del proyecto MultiCatálogo. FastAPI sobre SQLAlchemy 2 (async) + Alembic.

Vecina al backend Go/Fiber en `Backend/` que es el nucleo transaccional. Esta capa agrega capacidades SaaS: multi-tenant, billing, admin.

## Requisitos

- Python 3.11+
- [uv](https://docs.astral.sh/uv/) (gestor de paquetes)
- PostgreSQL 16+ (opcional para correr con Docker)

## Setup local

```bash
uv sync                          # crea .venv/ + instala deps
cp .env.example .env             # editar valores segun entorno

# Levantar Postgres (opcional) — se puede usar el docker-compose mas adelante
docker compose up -d saas-db

uv run uvicorn app.main:app --reload --port 8000
```

Abre `http://localhost:8000/health` o `http://localhost:8000/docs` para la UI Swagger.

## Tests

```bash
uv run pytest
```

## Variables de entorno

Ver `.env.example`. Las sensibles llevan prefijo `SAAS_`.
