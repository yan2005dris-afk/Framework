"""Tests del CRUD de users_saas.

Verifica la creación, listado, actualización y eliminación de usuarios
con aislamiento multi-tenant.
"""
from __future__ import annotations

import pytest
from httpx import AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession

from app.crud.user import user_crud
from app.models.user import UserSaaS
from app.schemas.user import UserSaaSCreate


@pytest.mark.asyncio
async def test_create_user(db: AsyncSession):
    """Verifica la creación de un usuario SaaS."""
    user = await user_crud.create(
        db,
        tenant_id=1,
        email="test@example.com",
        password="password123",
        rol="member",
    )

    assert user.id is not None
    assert user.email == "test@example.com"
    assert user.rol == "member"
    assert user.tenant_id == 1
    assert user_crud.verify_password("password123", user.hashed_password)


@pytest.mark.asyncio
async def test_get_user_by_id(db: AsyncSession):
    """Verifica la obtención de un usuario por ID."""
    user = await user_crud.create(
        db,
        tenant_id=1,
        email="test2@example.com",
        password="password123",
    )

    found = await user_crud.get_by_id(db, user.id, tenant_id=1)
    assert found is not None
    assert found.email == "test2@example.com"

    # Aislamiento: no encontrar en otro tenant
    not_found = await user_crud.get_by_id(db, user.id, tenant_id=2)
    assert not_found is None


@pytest.mark.asyncio
async def test_get_user_by_email(db: AsyncSession):
    """Verifica la obtención de un usuario por email."""
    await user_crud.create(
        db,
        tenant_id=1,
        email="test3@example.com",
        password="password123",
    )

    found = await user_crud.get_by_email(db, "test3@example.com", tenant_id=1)
    assert found is not None
    assert found.email == "test3@example.com"


@pytest.mark.asyncio
async def test_list_users_by_tenant(db: AsyncSession):
    """Verifica el listado de usuarios por tenant."""
    await user_crud.create(db, tenant_id=1, email="user1@example.com", password="pass")
    await user_crud.create(db, tenant_id=1, email="user2@example.com", password="pass")
    await user_crud.create(db, tenant_id=2, email="user3@example.com", password="pass")

    users, total = await user_crud.list_by_tenant(db, tenant_id=1)
    assert len(users) == 2
    assert total == 2

    # Aislamiento: tenant 2 solo tiene 1 usuario
    users2, total2 = await user_crud.list_by_tenant(db, tenant_id=2)
    assert len(users2) == 1
    assert total2 == 1


@pytest.mark.asyncio
async def test_update_user(db: AsyncSession):
    """Verifica la actualización de un usuario."""
    user = await user_crud.create(
        db,
        tenant_id=1,
        email="old@example.com",
        password="oldpass",
        rol="member",
    )

    updated = await user_crud.update(
        db,
        user_id=user.id,
        tenant_id=1,
        email="new@example.com",
        rol="admin",
    )

    assert updated is not None
    assert updated.email == "new@example.com"
    assert updated.rol == "admin"


@pytest.mark.asyncio
async def test_update_password(db: AsyncSession):
    """Verifica la actualización de password."""
    user = await user_crud.create(
        db,
        tenant_id=1,
        email="test@example.com",
        password="oldpass",
    )

    await user_crud.update(
        db,
        user_id=user.id,
        tenant_id=1,
        password="newpass",
    )

    updated = await user_crud.get_by_id(db, user.id, tenant_id=1)
    assert user_crud.verify_password("newpass", updated.hashed_password)
    assert not user_crud.verify_password("oldpass", updated.hashed_password)


@pytest.mark.asyncio
async def test_delete_user(db: AsyncSession):
    """Verifica la eliminación de un usuario."""
    user = await user_crud.create(
        db,
        tenant_id=1,
        email="test@example.com",
        password="pass",
    )

    deleted = await user_crud.delete(db, user.id, tenant_id=1)
    assert deleted is True

    found = await user_crud.get_by_id(db, user.id, tenant_id=1)
    assert found is None


@pytest.mark.asyncio
async def test_email_unique_per_tenant(db: AsyncSession):
    """Verifica que el email sea único por tenant."""
    from sqlalchemy.exc import IntegrityError

    first = await user_crud.create(
        db,
        tenant_id=1,
        email="test@example.com",
        password="pass",
    )

    # Mismo email en el mismo tenant viola uq_users_saas_tenant_email.
    with pytest.raises(IntegrityError):
        await user_crud.create(
            db,
            tenant_id=1,
            email="test@example.com",
            password="pass",
        )
    # Tras el IntegrityError la sesión queda abortada: hay que revertirla
    # antes de seguir usando la misma sesión.
    await db.rollback()

    # El intento fallido no persistió nada.
    assert (await user_crud.get_by_email(db, "test@example.com", tenant_id=1)).id == first.id

    # Mismo email en diferente tenant sí funciona.
    user2 = await user_crud.create(
        db,
        tenant_id=2,
        email="test@example.com",
        password="pass",
    )
    assert user2 is not None
