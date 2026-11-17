"""Dependencies FastAPI de la plataforma de tenancy."""
from __future__ import annotations

from typing import Annotated

from fastapi import Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.shared.tenancy.ports import TenancyPort
from app.shared.tenancy.service import tenancy_service


def get_tenancy_service() -> TenancyPort:
    """Inyecta el servicio de tenancy. Los tests pueden override esta dep."""
    return tenancy_service


#: `tenancy: TenancyServiceDep` en las firmas de los routers.
TenancyServiceDep = Annotated[TenancyPort, Depends(get_tenancy_service)]


async def tenant_or_404(
    db: AsyncSession,
    tenant_id: int,
    tenancy: TenancyPort,
) -> None:
    """404 explicito si el tenant del path no existe.

    Helper compartido por los modulos que escriben filas tenant-scoped: un FK
    inexistente se traduce a un 404 de dominio, no a un 500 de integridad.
    """
    if not await tenancy.tenant_exists(db, tenant_id):
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Tenant no encontrado",
        )
