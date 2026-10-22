"""API endpoints para planes, suscripciones y webhooks de Stripe.

- `GET  /api/saas/plans`: catalogo de planes, sin scoping de tenant.
- `GET/POST/DELETE /api/saas/tenants/{tenant_id}/subscription`: suscripcion
  del tenant, siempre dentro del scope del principal autenticado.
- `POST /api/saas/webhooks/stripe`: STUB sin integracion real con Stripe.

Identidad: la inyecta el gateway (ver `app/deps/principal.py`). Ningun
endpoint de suscripcion acepta un `tenant_id` que no sea el del principal
(un `superadmin` si puede operar sobre cualquier tenant).
"""
from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, Request, status
from fastapi.responses import Response
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.db import get_db, get_tenant_db
from app.deps.principal import PrincipalDep
from app.models.plan import Plan
from app.models.subscription import Subscription
from app.schemas.plan import PlanResponse, SubscriptionCreate, SubscriptionResponse

router = APIRouter(prefix="/api/saas", tags=["billing"])

# Lectura de planes: sesion simple (el catalogo no es tenant-scoped).
DbSession = Annotated[AsyncSession, Depends(get_db)]
# Suscripciones: sesion tenant-scoped (GUC + transaccion del request).
TenantDbSession = Annotated[AsyncSession, Depends(get_tenant_db)]


async def _get_subscription_or_404(db: AsyncSession, tenant_id: int) -> Subscription:
    result = await db.execute(select(Subscription).where(Subscription.tenant_id == tenant_id))
    subscription = result.scalar_one_or_none()
    if not subscription:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Suscripción no encontrada",
        )
    return subscription


@router.get("/plans", response_model=list[PlanResponse])
async def list_plans(db: DbSession, principal: PrincipalDep) -> list[Plan]:
    """Lista los planes disponibles (mismo catalogo para todos los tenants)."""
    del principal  # autenticado, pero el catalogo no es tenant-scoped
    result = await db.execute(select(Plan).order_by(Plan.price_monthly))
    return list(result.scalars().all())


@router.get("/tenants/{tenant_id}/subscription", response_model=SubscriptionResponse)
async def get_subscription(
    tenant_id: int,
    db: TenantDbSession,
    principal: PrincipalDep,
) -> Subscription:
    """Devuelve la suscripcion vigente del tenant."""
    return await _get_subscription_or_404(db, tenant_id)


@router.post("/tenants/{tenant_id}/subscription", response_model=SubscriptionResponse, status_code=status.HTTP_201_CREATED)
async def create_subscription(
    tenant_id: int,
    sub_data: SubscriptionCreate,
    db: TenantDbSession,
    principal: PrincipalDep,
) -> Subscription:
    """Crea la suscripcion del tenant a un plan.

    Un tenant tiene como maximo una suscripcion (columna con UNIQUE).
    """
    plan = (
        await db.execute(select(Plan).where(Plan.id == sub_data.plan_id))
    ).scalar_one_or_none()
    if not plan:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Plan no encontrado",
        )

    existing = (
        await db.execute(select(Subscription).where(Subscription.tenant_id == tenant_id))
    ).scalar_one_or_none()
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="El tenant ya tiene una suscripción activa",
        )

    # STUB de Stripe: en produccion se crearia la suscripcion en Stripe y se
    # guardaria su id en `stripe_subscription_id` via webhook.
    subscription = Subscription(
        tenant_id=tenant_id,
        plan_id=sub_data.plan_id,
        status=sub_data.status,
        cancel_at_period_end=sub_data.cancel_at_period_end,
    )
    db.add(subscription)
    # flush(), no commit(): la transaccion la cierra `get_tenant_db`.
    await db.flush()
    await db.refresh(subscription)
    return subscription


@router.delete("/tenants/{tenant_id}/subscription", status_code=status.HTTP_204_NO_CONTENT)
async def cancel_subscription(
    tenant_id: int,
    db: TenantDbSession,
    principal: PrincipalDep,
) -> Response:
    """Cancela la suscripcion al final del periodo vigente (no la borra)."""
    subscription = await _get_subscription_or_404(db, tenant_id)
    subscription.cancel_at_period_end = True
    subscription.status = "cancelled"
    await db.flush()
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@router.post("/webhooks/stripe")
async def stripe_webhook(request: Request) -> dict[str, str]:
    """Webhook stub de Stripe.

    NO autentica ni verifica nada: es un placeholder para el flujo real, que
    debe validar la firma HMAC de Stripe (`Stripe-Signature`) antes de
    tocar la suscripcion, y responder rapido para no agotar los reintentos.
    """
    payload = await request.json()
    return {"status": "received", "event_type": payload.get("type", "unknown")}