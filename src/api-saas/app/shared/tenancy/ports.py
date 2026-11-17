"""Puerto de tenancy consumido por los modulos de feature.

Los modulos necesitan saber si un `{tenant_id}` del path existe antes de
escribir (404 explicito en vez de un error de FK). No deben consultar la
tabla `tenants` directamente: eso los acoplaria al modelo de plataforma.
"""
from __future__ import annotations

from typing import Protocol

from sqlalchemy.ext.asyncio import AsyncSession


class TenancyPort(Protocol):
    """Contrato minimo de la plataforma multi-tenant."""

    async def tenant_exists(self, db: AsyncSession, tenant_id: int) -> bool:
        """`True` si el tenant existe."""
        ...
