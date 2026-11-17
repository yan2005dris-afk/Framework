"""Repositorio de billing: planes y suscripciones.

- El catalogo de planes es global (sin `tenant_id`).
- **Toda** query de suscripciones filtra por `tenant_id`.
- `flush()` y no `commit()`: la transaccion del request la cierra
  `get_tenant_db`.
"""
from __future__ import annotations

from typing import Optional, Protocol

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.billing.models import Plan, Subscription


class BillingRepositoryPort(Protocol):
    """Contrato de persistencia que consume `BillingService`."""

    async def list_plans(self, db: AsyncSession) -> list[Plan]: ...

    async def get_plan(self, db: AsyncSession, plan_id: int) -> Optional[Plan]: ...

    async def get_subscription(
        self, db: AsyncSession, tenant_id: int
    ) -> Optional[Subscription]: ...

    async def create_subscription(
        self,
        db: AsyncSession,
        tenant_id: int,
        plan_id: int,
        status: str,
        cancel_at_period_end: bool,
    ) -> Subscription: ...

    async def cancel_subscription(
        self, db: AsyncSession, subscription: Subscription
    ) -> Subscription: ...


class BillingRepository:
    """Operaciones de persistencia para planes y suscripciones."""

    async def list_plans(self, db: AsyncSession) -> list[Plan]:
        """Catalogo completo, de menor a mayor precio mensual."""
        result = await db.execute(select(Plan).order_by(Plan.price_monthly))
        return list(result.scalars().all())

    async def get_plan(self, db: AsyncSession, plan_id: int) -> Optional[Plan]:
        """Plan por id o `None`."""
        result = await db.execute(select(Plan).where(Plan.id == plan_id))
        return result.scalar_one_or_none()

    async def get_subscription(
        self, db: AsyncSession, tenant_id: int
    ) -> Optional[Subscription]:
        """Suscripcion del tenant o `None`."""
        result = await db.execute(
            select(Subscription).where(Subscription.tenant_id == tenant_id)
        )
        return result.scalar_one_or_none()

    async def create_subscription(
        self,
        db: AsyncSession,
        tenant_id: int,
        plan_id: int,
        status: str,
        cancel_at_period_end: bool,
    ) -> Subscription:
        """Crea la suscripcion del tenant al plan indicado.

        STUB de Stripe: en produccion se crearia la suscripcion en Stripe y se
        guardaria su id en `stripe_subscription_id` via webhook.
        """
        subscription = Subscription(
            tenant_id=tenant_id,
            plan_id=plan_id,
            status=status,
            cancel_at_period_end=cancel_at_period_end,
        )
        db.add(subscription)
        await db.flush()
        await db.refresh(subscription)
        return subscription

    async def cancel_subscription(
        self, db: AsyncSession, subscription: Subscription
    ) -> Subscription:
        """Marca la suscripcion para cancelarse al final del periodo."""
        subscription.cancel_at_period_end = True
        subscription.status = "cancelled"
        await db.flush()
        return subscription


#: Instancia sin estado: se inyecta via `BillingServiceDep`.
billing_repository = BillingRepository()
