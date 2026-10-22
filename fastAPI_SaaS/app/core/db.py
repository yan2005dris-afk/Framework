"""Engine async de SQLAlchemy 2.x + factory de sesiones.

Cualquier consumidor (routers, tests, migraciones en runtime) obtiene
sesiones via `get_db()` (FastAPI dependency) o creando una `SessionLocal()`
directamente. El engine es lazy: se crea al primer uso.

Dos dependencias, dos alcances:

- `get_db`: sesion simple, sin contexto de tenant. Para endpoints que no
  tocan datos de un tenant (health, catalogo de planes).
- `get_tenant_db`: sesion de un request tenant-scoped. Hace tres cosas que
  `get_db` no hace:
    1. valida el scope (`require_tenant_scope`) ANTES de tocar la BD, asi
       que un 403 no deja ni una query ejecutada;
    2. setea el GUC `app.current_tenant` en la conexion (Postgres) para que
       las politicas RLS de `0003_rls_policies` filtren por fila;
    3. maneja la transaccion del request: un unico commit al final, rollback
       ante cualquier error. Por eso el CRUD hace `flush()` y no `commit()`:
       `SET LOCAL` vive dentro de la transaccion y un commit intermedio lo
       borraria, dejando el resto del request sin filtro de tenant.

Convenciones:
- `echo` solo en dev; en prod se silencia.
- `pool_pre_ping=True` verifica la conexion antes de cada checkout para
  evitar errores por conexiones muertas (importante en deploys largos).
- `expire_on_commit=False` para mantener los atributos del ORM accesibles
  despues del commit sin tener que hacer un refresh explicito.
"""
from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Annotated

from fastapi import Depends
from sqlalchemy import text
from sqlalchemy.ext.asyncio import (
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)

from app.core.config import settings
from app.deps.principal import PrincipalDep, require_tenant_scope

engine = create_async_engine(
    settings.database_url,
    echo=settings.app_env == "dev",
    pool_pre_ping=True,
)

SessionLocal: async_sessionmaker[AsyncSession] = async_sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

#: GUC que leen las politicas RLS.
TENANT_GUC = "app.current_tenant"


def _dialect_name(db: AsyncSession) -> str:
    """Nombre del dialecto de la sesion (`postgresql`, `sqlite`, ...)."""
    bind = db.bind
    if bind is not None and hasattr(bind, "dialect"):
        return bind.dialect.name
    return db.get_bind().dialect.name


async def get_db() -> AsyncIterator[AsyncSession]:
    """FastAPI dependency. Usar como `Depends(get_db)` en los routers."""
    async with SessionLocal() as session:
        yield session


async def get_tenant_db(
    tenant_id: int,
    principal: PrincipalDep,
    db: Annotated[AsyncSession, Depends(get_db)],
) -> AsyncIterator[AsyncSession]:
    """Sesion de un request tenant-scoped: scope + GUC + transaccion.

    `tenant_id` lo resuelve FastAPI desde el path (`/tenants/{tenant_id}/...`),
    asi que la dependencia no puede ejecutarse sin un tenant en la ruta.
    """
    effective_tenant = require_tenant_scope(principal, tenant_id)

    if _dialect_name(db) == "postgresql":
        # `set_config(..., true)` == SET LOCAL: vive en la transaccion actual y
        # desaparece en el commit. Con el GUC sin valor, las politicas RLS
        # comparan contra NULL y no dejan pasar ninguna fila (fail closed).
        await db.execute(
            text(f"SELECT set_config('{TENANT_GUC}', :tenant_id, true)"),
            {"tenant_id": str(effective_tenant)},
        )

    try:
        yield db
        await db.commit()
    except Exception:
        await db.rollback()
        raise