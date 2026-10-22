"""Tests de `get_tenant_db`: scope + GUC + transaccion del request.

El aislamiento por fila se verifica contra Postgres en
`tests/integration/test_saas_rls_integration.py` (ahi RLS existe de verdad).
Aqui se cubre lo que si es testeable en sqlite: que la transaccion del request
es de un solo commit, que un error la revierte, y que el scope se valida antes
de tocar la BD.
"""
from __future__ import annotations

from typing import Any

import pytest
from fastapi import HTTPException
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from app.core import db as db_module
from app.core.db import get_tenant_db
from app.deps.principal import Principal

pytestmark = pytest.mark.asyncio


def _principal(**kwargs: Any) -> Principal:
    base = {
        "user_id": "1",
        "email": "admin@tenant.example",
        "tenant_id": 1,
        "rol_saas": "admin",
    }
    base.update(kwargs)
    return Principal(**base)  # type: ignore[arg-type]


@pytest.fixture
def db_url(tmp_path) -> str:
    """sqlite en ARCHIVO (no :memory:) para que otra sesion vea el commit."""
    return f"sqlite+aiosqlite:///{tmp_path / 'tenant_session.db'}"


@pytest.fixture
async def maker(db_url: str) -> async_sessionmaker[AsyncSession]:
    """Sessionmaker sobre sqlite en archivo, con el schema ya creado."""
    from app.models import Base

    engine = create_async_engine(db_url)
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    return async_sessionmaker(bind=engine, class_=AsyncSession, expire_on_commit=False)


class _RecordingSession:
    """Sesion falsa: solo registra lo que le pide la dependencia."""

    def __init__(self) -> None:
        self.executed: list[tuple[str, dict]] = []
        self.committed = False
        self.rolled_back = False

    async def execute(self, statement: Any, params: dict | None = None) -> None:
        self.executed.append((str(statement), params or {}))

    async def commit(self) -> None:
        self.committed = True

    async def rollback(self) -> None:
        self.rolled_back = True


async def test_dialect_of_sqlite_session_is_sqlite(maker) -> None:
    async with maker() as session:
        assert db_module._dialect_name(session) == "sqlite"


async def test_scope_is_rejected_before_any_query(maker) -> None:
    """Un 403 no debe ejecutar una sola sentencia."""
    async with maker() as session:
        gen = get_tenant_db(
            tenant_id=7, principal=_principal(tenant_id=1), db=session
        )
        with pytest.raises(HTTPException) as exc:
            await anext(gen)
        assert exc.value.status_code == 403
        assert exc.value.detail == "Solo puedes operar sobre tu propio tenant"


async def test_superadmin_reaches_any_tenant(maker) -> None:
    async with maker() as session:
        gen = get_tenant_db(
            tenant_id=7, principal=_principal(tenant_id=1, rol_saas="superadmin"), db=session
        )
        yielded = await anext(gen)
        assert yielded is session
        await gen.aclose()


async def test_request_commits_once_at_the_end(maker, db_url: str) -> None:
    from app.models.tenant import Tenant

    async with maker() as session:
        gen = get_tenant_db(tenant_id=1, principal=_principal(), db=session)
        db = await anext(gen)
        db.add(Tenant(slug="committed", name="committed"))
        await db.flush()
        with pytest.raises(StopAsyncIteration):
            await anext(gen)

    # Nueva sesion: la fila quedo committeada por la dependencia.
    async with maker() as reader:
        from sqlalchemy import select

        rows = (await reader.execute(select(Tenant))).scalars().all()
        assert [t.slug for t in rows] == ["committed"]


async def test_error_inside_request_rolls_back(maker) -> None:
    """Un 404 a mitad de request no deja datos persistidos."""
    from app.models.tenant import Tenant
    from sqlalchemy import select

    async with maker() as session:
        gen = get_tenant_db(tenant_id=1, principal=_principal(), db=session)
        db = await anext(gen)
        db.add(Tenant(slug="rolled-back", name="rolled-back"))
        await db.flush()
        # El error del handler se propaga tras el rollback.
        with pytest.raises(RuntimeError, match="boom"):
            await gen.athrow(RuntimeError("boom"))

    async with maker() as reader:
        assert (await reader.execute(select(Tenant))).scalars().all() == []


async def test_postgres_sets_tenant_guc(monkeypatch) -> None:
    """En Postgres la dependencia setea `app.current_tenant` con el tenant efectivo."""
    monkeypatch.setattr(db_module, "_dialect_name", lambda _db: "postgresql")
    fake = _RecordingSession()

    gen = get_tenant_db(
        tenant_id=9, principal=_principal(tenant_id=1, rol_saas="superadmin"), db=fake
    )
    await anext(gen)
    with pytest.raises(StopAsyncIteration):
        await anext(gen)

    assert len(fake.executed) == 1
    statement, params = fake.executed[0]
    assert "set_config('app.current_tenant'" in statement
    # El tenant efectivo es el de la ruta (superadmin), no el del principal.
    assert params == {"tenant_id": "9"}
    assert fake.committed is True
    assert fake.rolled_back is False


async def test_postgres_guc_is_transaction_local(monkeypatch) -> None:
    """`set_config(..., true)` == SET LOCAL: se pierde en el commit."""
    monkeypatch.setattr(db_module, "_dialect_name", lambda _db: "postgresql")
    fake = _RecordingSession()

    gen = get_tenant_db(
        tenant_id=9, principal=_principal(tenant_id=1, rol_saas="superadmin"), db=fake
    )
    await anext(gen)
    with pytest.raises(StopAsyncIteration):
        await anext(gen)

    statement, _ = fake.executed[0]
    assert statement.rstrip().endswith("true)")