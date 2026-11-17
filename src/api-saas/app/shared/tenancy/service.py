"""Servicio de plataforma multi-tenant.

Unico consumidor permitido del modelo `Tenant` desde los modulos de feature:
estos dependen de `TenancyPort`, no de la tabla.
"""
from __future__ import annotations

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.shared.tenancy.models import Tenant


class TenancyService:
    """Operaciones de lectura sobre la plataforma de tenants."""

    async def tenant_exists(self, db: AsyncSession, tenant_id: int) -> bool:
        """`True` si el tenant existe."""
        result = await db.execute(select(Tenant.id).where(Tenant.id == tenant_id))
        return result.scalar_one_or_none() is not None


#: Instancia sin estado: se inyecta via `TenancyServiceDep`.
tenancy_service = TenancyService()
