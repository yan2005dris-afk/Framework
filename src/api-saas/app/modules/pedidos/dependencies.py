"""Inyeccion de dependencias del modulo `pedidos`.

Se usa `Annotated[..., Depends(...)]` en las firmas (patron recomendado por
FastAPI). Los tests pueden sustituir el servicio/repositorio con
`app.dependency_overrides` o inyectando otra implementacion del puerto.
"""
from __future__ import annotations

from typing import Annotated

from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.pedidos.service import PedidoService, pedido_service
from app.shared.db.session import get_tenant_db


def get_pedido_service() -> PedidoService:
    """Inyecta el servicio de pedidos (singleton sin estado)."""
    return pedido_service


#: Sesion tenant-scoped: scope + GUC de RLS + transaccion del request.
DbSession = Annotated[AsyncSession, Depends(get_tenant_db)]

#: Servicio de casos de uso.
PedidoServiceDep = Annotated[PedidoService, Depends(get_pedido_service)]
