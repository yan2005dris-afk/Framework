"""Fixtures pytest + cliente httpx async.

Convenciones:
- `client`: AsyncClient con ASGITransport que monta la app FastAPI in-process
  (sin levantar uvicorn). Cada test obtiene su propio cliente pero
  comparte el event loop (scope='session' evita 'RuntimeError: Event loop is closed').
- El engine async de SQLAlchemy NO se inicializa al importar `app.main`
  (es lazy en `get_db()`), asi que los tests de smoke de /health no
  requieren Postgres. Los tests que toquen DB usaran una fixture separada
  con sqlite + aiosqlite (T1.6+).
"""
from __future__ import annotations

import asyncio
from typing import AsyncIterator

import pytest
from httpx import ASGITransport, AsyncClient

from app.main import app


@pytest.fixture(scope="session")
def event_loop():
    """Event loop unico para toda la sesion de tests. Necesario para AsyncClient."""
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
async def client() -> AsyncIterator[AsyncClient]:
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test",
    ) as c:
        yield c
