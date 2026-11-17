"""API endpoints para planes, suscripciones y webhooks de Stripe.

- `GET  /api/saas/plans`: catalogo de planes, sin scoping de tenant.
- `GET/POST/DELETE /api/saas/tenants/{tenant_id}/subscription`: suscripcion
  del tenant, siempre dentro del scope del principal autenticado.
- `POST /api/saas/webhooks/stripe`: STUB sin integracion real con Stripe.

Identidad: la inyecta el gateway (ver `app/shared/tenancy/principal.py`).
Ningun endpoint de suscripcion acepta un `tenant_id` que no sea el del
principal (un `superadmin` si puede operar sobre cualquier tenant).

Errores de dominio los traduce `register()` en
`app/modules/billing/__init__.py`.
"""
from __future__ import annotations

from fastapi import APIRouter, Request, status
from fastapi.responses import Response

from app.modules.billing.dependencies import BillingServiceDep, DbSession, TenantDbSession
from app.modules.billing.schemas import PlanResponse, SubscriptionCreate, SubscriptionResponse
from app.shared.tenancy.principal import PrincipalDep

router = APIRouter(prefix="/api/saas", tags=["billing"])


@router.get("/plans", response_model=list[PlanResponse])
async def list_plans(
    db: DbSession,
    service: BillingServiceDep,
    principal: PrincipalDep,
) -> object:
    """Lista los planes disponibles (mismo catalogo para todos los tenants)."""
    del principal  # autenticado, pero el catalogo no es tenant-scoped
    return await service.listar_planes(db)


@router.get("/tenants/{tenant_id}/subscription", response_model=SubscriptionResponse)
async def get_subscription(
    tenant_id: int,
    db: TenantDbSession,
    service: BillingServiceDep,
    principal: PrincipalDep,
) -> object:
    """Devuelve la suscripcion vigente del tenant."""
    return await service.obtener_suscripcion(db, tenant_id)


@router.post(
    "/tenants/{tenant_id}/subscription",
    response_model=SubscriptionResponse,
    status_code=status.HTTP_201_CREATED,
)
async def create_subscription(
    tenant_id: int,
    sub_data: SubscriptionCreate,
    db: TenantDbSession,
    service: BillingServiceDep,
    principal: PrincipalDep,
) -> object:
    """Crea la suscripcion del tenant a un plan.

    Un tenant tiene como maximo una suscripcion (columna con UNIQUE).
    """
    return await service.crear_suscripcion(
        db,
        tenant_id=tenant_id,
        plan_id=sub_data.plan_id,
        status=sub_data.status,
        cancel_at_period_end=sub_data.cancel_at_period_end,
    )


@router.delete("/tenants/{tenant_id}/subscription", status_code=status.HTTP_204_NO_CONTENT)
async def cancel_subscription(
    tenant_id: int,
    db: TenantDbSession,
    service: BillingServiceDep,
    principal: PrincipalDep,
) -> Response:
    """Cancela la suscripcion al final del periodo vigente (no la borra)."""
    await service.cancelar_suscripcion(db, tenant_id)
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
