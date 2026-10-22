"""Tests de API del CRUD de users_saas.

Verifica el contrato HTTP y, sobre todo, el aislamiento multi-tenant:
identidad obligatoria (401), scope de tenant (403), roles (403) y que un
usuario de otro tenant nunca se ve ni se modifica (404).
"""
from __future__ import annotations

import pytest
from httpx import AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession

from tests.conftest import auth_headers

pytestmark = pytest.mark.asyncio

USERS_URL = "/api/saas/tenants/{tenant_id}/users"


async def _create_tenant(db: AsyncSession, slug: str) -> int:
    from app.models.tenant import Tenant

    tenant = Tenant(slug=slug, name=slug)
    db.add(tenant)
    await db.commit()
    await db.refresh(tenant)
    return tenant.id


async def _create_user(db: AsyncSession, tenant_id: int, email: str, rol: str = "member") -> int:
    from app.crud.user import user_crud

    user = await user_crud.create(
        db, tenant_id=tenant_id, email=email, password="secret123", rol=rol
    )
    return user.id


async def test_create_user_returns_201_without_hash(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.post(
        USERS_URL.format(tenant_id=tenant_id),
        json={"email": "nuevo@acme.example", "password": "secret123", "rol": "member"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 201, resp.text
    body = resp.json()
    assert body["tenant_id"] == tenant_id
    assert body["email"] == "nuevo@acme.example"
    assert "hashed_password" not in body
    assert "password" not in body


async def test_create_user_without_identity_returns_401(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.post(
        USERS_URL.format(tenant_id=tenant_id),
        json={"email": "anon@acme.example", "password": "secret123"},
    )

    assert resp.status_code == 401, resp.text
    assert await _user_count(db, tenant_id) == 0


async def test_member_cannot_create_user(client: AsyncClient, db: AsyncSession) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.post(
        USERS_URL.format(tenant_id=tenant_id),
        json={"email": "member@acme.example", "password": "secret123"},
        headers=auth_headers(tenant_id=tenant_id, rol_saas="member"),
    )

    assert resp.status_code == 403, resp.text
    assert await _user_count(db, tenant_id) == 0


async def test_admin_of_other_tenant_gets_403(client: AsyncClient, db: AsyncSession) -> None:
    own = await _create_tenant(db, "own")
    other = await _create_tenant(db, "other")

    resp = await client.get(
        USERS_URL.format(tenant_id=other), headers=auth_headers(tenant_id=own)
    )

    assert resp.status_code == 403, resp.text


async def test_superadmin_can_reach_any_tenant(client: AsyncClient, db: AsyncSession) -> None:
    other = await _create_tenant(db, "other")

    resp = await client.get(
        USERS_URL.format(tenant_id=other),
        headers=auth_headers(tenant_id=1, rol_saas="superadmin"),
    )

    assert resp.status_code == 200, resp.text


async def test_duplicate_email_in_same_tenant_returns_400(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    await _create_user(db, tenant_id, "dup@acme.example")

    resp = await client.post(
        USERS_URL.format(tenant_id=tenant_id),
        json={"email": "dup@acme.example", "password": "secret123"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 400, resp.text
    assert "ya existe" in resp.json()["detail"]


async def test_same_email_allowed_in_two_tenants(client: AsyncClient, db: AsyncSession) -> None:
    first = await _create_tenant(db, "acme")
    second = await _create_tenant(db, "globex")

    payload = {"email": "shared@example.com", "password": "secret123"}
    r1 = await client.post(
        USERS_URL.format(tenant_id=first), json=payload, headers=auth_headers(tenant_id=first)
    )
    r2 = await client.post(
        USERS_URL.format(tenant_id=second),
        json=payload,
        headers=auth_headers(tenant_id=second),
    )

    assert r1.status_code == 201, r1.text
    assert r2.status_code == 201, r2.text


async def test_list_isolates_tenants(client: AsyncClient, db: AsyncSession) -> None:
    first = await _create_tenant(db, "acme")
    second = await _create_tenant(db, "globex")
    await _create_user(db, first, "a@acme.example")
    await _create_user(db, second, "b@globex.example")

    resp = await client.get(
        USERS_URL.format(tenant_id=first), headers=auth_headers(tenant_id=first)
    )

    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["total"] == 1
    assert [u["email"] for u in body["users"]] == ["a@acme.example"]


async def test_get_user_from_other_tenant_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    first = await _create_tenant(db, "acme")
    second = await _create_tenant(db, "globex")
    user_id = await _create_user(db, second, "b@globex.example")

    resp = await client.get(
        f"{USERS_URL.format(tenant_id=first)}/{user_id}",
        headers=auth_headers(tenant_id=first),
    )

    assert resp.status_code == 404, resp.text


async def test_patch_changes_role_and_password(client: AsyncClient, db: AsyncSession) -> None:
    tenant_id = await _create_tenant(db, "acme")
    user_id = await _create_user(db, tenant_id, "user@acme.example")

    resp = await client.patch(
        f"{USERS_URL.format(tenant_id=tenant_id)}/{user_id}",
        json={"rol": "admin", "password": "nuevasecreta"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 200, resp.text
    assert resp.json()["rol"] == "admin"

    from app.crud.user import user_crud

    stored = await user_crud.get_by_id(db, user_id, tenant_id)
    assert user_crud.verify_password("nuevasecreta", stored.hashed_password)
    assert not user_crud.verify_password("secret123", stored.hashed_password)


async def test_patch_duplicate_email_returns_400(client: AsyncClient, db: AsyncSession) -> None:
    tenant_id = await _create_tenant(db, "acme")
    await _create_user(db, tenant_id, "a@acme.example")
    target = await _create_user(db, tenant_id, "b@acme.example")

    resp = await client.patch(
        f"{USERS_URL.format(tenant_id=tenant_id)}/{target}",
        json={"email": "a@acme.example"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 400, resp.text


async def test_member_cannot_delete(client: AsyncClient, db: AsyncSession) -> None:
    tenant_id = await _create_tenant(db, "acme")
    user_id = await _create_user(db, tenant_id, "user@acme.example")

    resp = await client.delete(
        f"{USERS_URL.format(tenant_id=tenant_id)}/{user_id}",
        headers=auth_headers(tenant_id=tenant_id, rol_saas="member"),
    )

    assert resp.status_code == 403, resp.text


async def test_delete_returns_204_and_removes_user(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    user_id = await _create_user(db, tenant_id, "user@acme.example")

    resp = await client.delete(
        f"{USERS_URL.format(tenant_id=tenant_id)}/{user_id}",
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 204, resp.text
    assert resp.content == b""

    from app.crud.user import user_crud

    assert await user_crud.get_by_id(db, user_id, tenant_id) is None


async def test_delete_unknown_user_returns_404(client: AsyncClient, db: AsyncSession) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.delete(
        f"{USERS_URL.format(tenant_id=tenant_id)}/424242",
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 404, resp.text


async def test_create_user_on_unknown_tenant_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    resp = await client.post(
        USERS_URL.format(tenant_id=999),
        json={"email": "nuevo@acme.example", "password": "secret123"},
        headers=auth_headers(tenant_id=999),
    )

    assert resp.status_code == 404, resp.text


async def _user_count(db: AsyncSession, tenant_id: int) -> int:
    """Cuenta usuarios del tenant leyendo directo de la BD."""
    from app.crud.user import user_crud

    _, total = await user_crud.list_by_tenant(db, tenant_id)
    return total