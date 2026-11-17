"""Casos de uso del modulo billing.

Reglas que no son queries: un tenant tiene como maximo una suscripcion y solo
se puede suscribir a un plan existente. El webhook de Stripe queda como stub
en el router (no toca persistencia todavia).
"""
from __future__ import annotations

from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.billing.exceptions import (
    PlanNoEncontradoError,
    SuscripcionExistenteError,
    SuscripcionNoEncontradaError,
)
from app.modules.billing.models import Plan, Subscription
from app.modules.billing.repository import BillingRepositoryPort, billing_repository


class BillingService:
    """Operaciones de negocio sobre planes y suscripciones."""

    def __init__(self, repository: BillingRepositoryPort = billing_repository) -> None:
        self._repository = repository

    async def listar_planes(self, db: AsyncSession) -> list[Plan]:
        """Catalogo de planes (igual para todos los tenants)."""
        return await self._repository.list_plans(db)

    async def obtener_suscripcion(
        self, db: AsyncSession, tenant_id: int
    ) -> Subscription:
        """Suscripcion vigente del tenant o `SuscripcionNoEncontradaError`."""
        subscription = await self._repository.get_subscription(db, tenant_id)
        if subscription is None:
            raise SuscripcionNoEncontradaError(tenant_id)
        return subscription

    async def crear_suscripcion(
        self,
        db: AsyncSession,
        tenant_id: int,
        plan_id: int,
        status: str,
        cancel_at_period_end: bool,
    ) -> Subscription:
        """Suscribe el tenant a un plan.

        Lanza `PlanNoEncontradoError` (404) si el plan no existe y
        `SuscripcionExistenteError` (400) si el tenant ya esta suscrito.
        """
        if await self._repository.get_plan(db, plan_id) is None:
            raise PlanNoEncontradoError(plan_id)

        if await self._repository.get_subscription(db, tenant_id) is not None:
            raise SuscripcionExistenteError(tenant_id)

        return await self._repository.create_subscription(
            db,
            tenant_id=tenant_id,
            plan_id=plan_id,
            status=status,
            cancel_at_period_end=cancel_at_period_end,
        )

    async def cancelar_suscripcion(
        self, db: AsyncSession, tenant_id: int
    ) -> Subscription:
        """Cancela al final del periodo vigente (no borra la fila)."""
        subscription = await self.obtener_suscripcion(db, tenant_id)
        return await self._repository.cancel_subscription(db, subscription)


#: Instancia sin estado: se inyecta via `BillingServiceDep`.
billing_service = BillingService()
