"""MultiCatalogo SaaS package.

Capa SaaS / multitenant sobre FastAPI. Vive junto al backend Go/Fiber
(`Backend/`) que es el nucleo transaccional. Esta capa agrega capacidades
SaaS: multi-tenancy, billing, admin y APIs que el nucleo no necesita
manejar en linea caliente.

Convenciones:
- Cada tabla SaaS tiene `tenant_id` FK -> `tenants.id`.
- Async-first: SQLAlchemy 2.x con asyncpg.
- Organizacion por bounded contexts: `app/shared/` (plataforma) y
  `app/modules/{pedidos,usuarios,billing}/` (features).
- Config por env via pydantic-settings (`app/shared/config.py`).
- Settings publica en `app.shared.config.settings`.
"""

__version__ = "0.1.0"
