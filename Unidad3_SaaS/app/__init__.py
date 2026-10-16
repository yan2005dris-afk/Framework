"""MultiCatalogo SaaS package.

Capa SaaS / multitenant sobre FastAPI. Vive junto al backend Go/Fiber
(`Backend/`) que es el nucleo transaccional. Esta capa agrega capacidades
SaaS: multi-tenancy, billing, admin y APIs que el nucleo no necesita
manejar en linea caliente.

Convenciones:
- Cada tabla SaaS tiene `tenant_id` FK -> `tenants.id`.
- Async-first: SQLAlchemy 2.x con asyncpg.
- Config por env via pydantic-settings (`app/core/config.py`).
- Settings publica en `app.core.config.settings`.
"""

__version__ = "0.1.0"
