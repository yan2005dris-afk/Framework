"""Fixtures pytest + cliente httpx async.

Convenciones:
- `client`: AsyncClient con ASGITransport que monta la app FastAPI in-process
  (sin levantar uvicorn). Cada test obtiene su propio cliente pero
  comparte el event loop (scope='session' evita 'RuntimeError: Event loop is closed').
- El engine async de SQLAlchemy NO se inicializa al importar `app.main`
  (es lazy en `get_db()`), asi que los tests de smoke de /health no
  requieren Postgres.
- `db`: sesion async sobre sqlite in-memory con el schema de `Base.metadata`.
  sqlite no tiene RLS (eso se verifica contra Postgres en
  `tests/integration/saas_rls_integration.py`), asi que el aislamiento se
  prueba aqui a nivel de codigo: cada query del CRUD filtra por tenant_id.
- `auth_headers`: headers que inyecta el gateway via forward_auth. Los tests
  de API los pasan explicitamente; sin ellos la respuesta es 401.
"""
from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.pool import StaticPool

from app.core.db import get_db
from app.main import app
from app.models import Base  # noqa: F401  (registra todos los modelos)


@pytest.fixture
async def session_factory() -> AsyncIterator[async_sessionmaker[AsyncSession]]:
    """Engine + sessionmaker sqlite in-memory, schema nuevo por test."""
    engine = create_async_engine(
        "sqlite+aiosqlite://",
        poolclass=StaticPool,
        connect_args={"check_same_thread": False},
    )
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    factory = async_sessionmaker(bind=engine, class_=AsyncSession, expire_on_commit=False)
    yield factory
    await engine.dispose()


@pytest.fixture
async def db(session_factory: async_sessionmaker[AsyncSession]) -> AsyncIterator[AsyncSession]:
    """Sesion de BD para tests de CRUD directo."""
    async with session_factory() as session:
        yield session


@pytest.fixture
async def client(
    session_factory: async_sessionmaker[AsyncSession],
) -> AsyncIterator[AsyncClient]:
    """Cliente in-process con `get_db` sobrerodeido al sqlite del test."""

    async def _override_get_db() -> AsyncIterator[AsyncSession]:
        async with session_factory() as session:
            yield session

    app.dependency_overrides[get_db] = _override_get_db
    try:
        async with AsyncClient(
            transport=ASGITransport(app=app),
            base_url="http://test",
        ) as c:
            yield c
    finally:
        app.dependency_overrides.pop(get_db, None)


def auth_headers(
    user_id: str = "1",
    email: str = "admin@tenant.example",
    tenant_id: int | str | None = 1,
    rol_saas: str = "admin",
) -> dict[str, str]:
    """Headers que el gateway copia desde la respuesta del auth-service."""
    headers: dict[str, str] = {
        "X-User-Id": user_id,
        "X-User-Email": email,
        "X-User-Rol": "",
    }
    if tenant_id is not None:
        headers["X-Tenant-Id"] = str(tenant_id)
    if rol_saas:
        headers["X-Rol-Saas"] = rol_saas
    return headers

