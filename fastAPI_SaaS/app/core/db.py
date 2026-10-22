"""Engine async de SQLAlchemy 2.x + factory de sesiones.

Cualquier consumidor (ruters, tests, migraciones en runtime) obtiene
sesiones via `get_db()` (FastAPI dependency) o creando un `SessionLocal()`
directamente. El engine es lazy: se crea al primer uso.

Convenciones:
- `echo` solo en dev; en prod se silencia.
- `pool_pre_ping=True` verifica la conexión antes de cada checkout para
  evitar errores por conexiones muertas (importante en deploys largos).
- `expire_on_commit=False` para mantener los atributos del ORM accesibles
  despues del commit sin tener que hacer un refresh explícito.
"""
from __future__ import annotations

from typing import AsyncGenerator

from sqlalchemy.ext.asyncio import (
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)

from app.core.config import settings

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


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    """FastAPI dependency. Usar como `Depends(get_db)` en los routers."""
    async with SessionLocal() as session:
        yield session
