"""Inyeccion de dependencias del modulo `billing`.

Dos sesiones con alcances distintos, igual que antes del refactor:

- `DbSession` (`get_db`): catalogo de planes, no tenant-scoped.
- `TenantDbSession` (`get_tenant_db`): suscripciones, con GUC + transaccion.
"""
from __future__ import annotations

from typing import Annotated

from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.billing.service import BillingService, billing_service
from app.shared.db.session import get_db, get_tenant_db


def get_billing_service() -> BillingService:
    """Inyecta el servicio de billing (singleton sin estado)."""
    return billing_service


#: Lectura de planes: sesion simple.
DbSession = Annotated[AsyncSession, Depends(get_db)]

#: Suscripciones: sesion tenant-scoped.
TenantDbSession = Annotated[AsyncSession, Depends(get_tenant_db)]

#: Servicio de casos de uso.
BillingServiceDep = Annotated[BillingService, Depends(get_billing_service)]
