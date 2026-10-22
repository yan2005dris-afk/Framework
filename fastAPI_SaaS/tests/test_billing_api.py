"""Tests de API de billing: planes, suscripciones y webhook stub.

Verifica el contrato HTTP con la identidad real del gateway: sin identidad
401, fuera de scope 403, y que un tenant no pueda ver/cancelar la
suscripcion de otro.
"""
from __future__ import annotations

from decimal import Decimal

import pytest
from httpx import AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession

from tests.conftest import auth_headers

pytestmark = pytest.mark.asyncio

SUB_URL = "/api/saas/tenants/{tenant_id}/subscription"
PLANS_URL = "/api/saas/plans"
WEBHOOK_URL = "/api/saas/webhooks/stripe"


async def _tenant(db: AsyncSession, slug: str) -> int:
    from app.models.tenant import Tenant

    tenant = Tenant(slug=slug, name=slug)
    db.add(tenant)
    await db.commit()
    await db.refresh(tenant)
    return tenant.id


async def _plan(db: AsyncSession, slug: str = "pro") -> int:
    from app.models.plan import Plan

    plan = Plan(
        slug=slug,
        name=slug.title(),
        description="plan de prueba",
        price_monthly=Decimal("29.90"),
        price_yearly=Decimal("299.00"),
        features=["a", "b"],
    )
    db.add(plan)
    await db.commit()
    await db.refresh(plan)
    return plan.id


async def test_list_plans_requires_identity(client: AsyncClient) -> None:
    resp = await client.get(PLANS_URL)

    assert resp.status_code == 401, resp.text


async def test_list_plans_returns_catalog(client: AsyncClient, db: AsyncSession) -> None:
    await _plan(db, "free")
    await _plan(db, "pro")

    resp = await client.get(PLANS_URL, headers=auth_headers(tenant_id=1))

    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert {p["slug"] for p in body} == {"free", "pro"}
    # El precio llega como numero desde Numeric(10,2) via Decimal.
    assert {p["slug"]: p["price_monthly"] for p in body}["pro"] in (29.9, "29.90")


async def test_plans_accepts_any_tenant(client: AsyncClient, db: AsyncSession) -> None:
    """El catalogo de planes no es tenant-scoped (mismo para todos)."""
    await _plan(db)

    resp = await client.get(PLANS_URL, headers=auth_headers(tenant_id=42, rol_saas="member"))

    assert resp.status_code == 200, resp.text


async def test_create_subscription_returns_201(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _tenant(db, "acme")
    plan_id = await _plan(db)

    resp = await client.post(
        SUB_URL.format(tenant_id=tenant_id),
        json={"plan_id": plan_id, "status": "active", "cancel_at_period_end": False},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 201, resp.text
    body = resp.json()
    assert body["tenant_id"] == tenant_id
    assert body["plan_id"] == plan_id
    assert body["cancel_at_period_end"] is False


async def test_create_subscription_unknown_plan_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _tenant(db, "acme")

    resp = await client.post(
        SUB_URL.format(tenant_id=tenant_id),
        json={"plan_id": 9999},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 404, resp.text
    assert "Plan" in resp.json()["detail"]


async def test_create_subscription_twice_returns_400(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _tenant(db, "acme")
    plan_id = await _plan(db)
    payload = {"plan_id": plan_id}
    first = await client.post(
        SUB_URL.format(tenant_id=tenant_id), json=payload, headers=auth_headers(tenant_id=tenant_id)
    )
    assert first.status_code == 201, first.text

    second = await client.post(
        SUB_URL.format(tenant_id=tenant_id), json=payload, headers=auth_headers(tenant_id=tenant_id)
    )

    assert second.status_code == 400, second.text
    assert "ya tiene" in second.json()["detail"]


async def test_get_subscription_of_other_tenant_returns_403(
    client: AsyncClient, db: AsyncSession
) -> None:
    own = await _tenant(db, "own")
    other = await _tenant(db, "other")
    plan_id = await _plan(db)
    await client.post(
        SUB_URL.format(tenant_id=other),
        json={"plan_id": plan_id},
        headers=auth_headers(tenant_id=other),
    )

    resp = await client.get(
        SUB_URL.format(tenant_id=other), headers=auth_headers(tenant_id=own)
    )

    assert resp.status_code == 403, resp.text


async def test_get_subscription_without_identity_returns_401(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _tenant(db, "acme")

    resp = await client.get(SUB_URL.format(tenant_id=tenant_id))

    assert resp.status_code == 401, resp.text


async def test_get_subscription_missing_returns_404(client: AsyncClient, db: AsyncSession) -> None:
    tenant_id = await _tenant(db, "acme")

    resp = await client.get(
        SUB_URL.format(tenant_id=tenant_id), headers=auth_headers(tenant_id=tenant_id)
    )

    assert resp.status_code == 404, resp.text


async def test_cancel_marks_period_end_instead_of_deleting(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _tenant(db, "acme")
    plan_id = await _plan(db)
    await client.post(
        SUB_URL.format(tenant_id=tenant_id),
        json={"plan_id": plan_id},
        headers=auth_headers(tenant_id=tenant_id),
    )

    resp = await client.delete(
        SUB_URL.format(tenant_id=tenant_id), headers=auth_headers(tenant_id=tenant_id)
    )

    assert resp.status_code == 204, resp.text

    # La fila sigue existiendo: cancelada, no borrada.
    from sqlalchemy import select

    from app.models.subscription import Subscription

    sub = (
        await db.execute(select(Subscription).where(Subscription.tenant_id == tenant_id))
    ).scalar_one()
    assert sub.cancel_at_period_end is True
    assert sub.status == "cancelled"


async def test_cancel_without_subscription_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _tenant(db, "acme")

    resp = await client.delete(
        SUB_URL.format(tenant_id=tenant_id), headers=auth_headers(tenant_id=tenant_id)
    )

    assert resp.status_code == 404, resp.text


async def test_superadmin_reads_any_tenant_subscription(
    client: AsyncClient, db: AsyncSession
) -> None:
    other = await _tenant(db, "other")
    plan_id = await _plan(db)
    await client.post(
        SUB_URL.format(tenant_id=other),
        json={"plan_id": plan_id},
        headers=auth_headers(tenant_id=other),
    )

    resp = await client.get(
        SUB_URL.format(tenant_id=other),
        headers=auth_headers(tenant_id=1, rol_saas="superadmin"),
    )

    assert resp.status_code == 200, resp.text


async def test_stripe_webhook_stub_echoes_event_type(client: AsyncClient) -> None:
    resp = await client.post(WEBHOOK_URL, json={"type": "invoice.payment_succeeded"})

    assert resp.status_code == 200, resp.text
    assert resp.json() == {
        "status": "received",
        "event_type": "invoice.payment_succeeded",
    }


async def test_stripe_webhook_stub_defaults_unknown_event(client: AsyncClient) -> None:
    resp = await client.post(WEBHOOK_URL, json={})

    assert resp.status_code == 200, resp.text
    assert resp.json()["event_type"] == "unknown"