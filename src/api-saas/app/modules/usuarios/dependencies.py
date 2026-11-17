"""Inyeccion de dependencias del modulo `usuarios`."""
from __future__ import annotations

from typing import Annotated

from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.usuarios.service import UserService, user_service
from app.shared.db.session import get_tenant_db


def get_user_service() -> UserService:
    """Inyecta el servicio de usuarios (singleton sin estado)."""
    return user_service


#: Sesion tenant-scoped: scope + GUC de RLS + transaccion del request.
DbSession = Annotated[AsyncSession, Depends(get_tenant_db)]

#: Servicio de casos de uso.
UserServiceDep = Annotated[UserService, Depends(get_user_service)]
