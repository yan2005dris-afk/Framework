"""API endpoints para CRUD de users_saas.

Gestion de usuarios dentro de un tenant, con aislamiento multi-tenant.

Identidad: la inyecta el gateway (`forward_auth` -> auth-service) como
`X-User-Id` / `X-Tenant-Id` / `X-Rol-Saas`; estos handlers la reciben via
`PrincipalDep` y la validan contra el `{tenant_id}` del path. Un request sin
identidad recibe 401 y un `member` no administra usuarios (403).

Los errores de dominio (`UsuarioNoEncontradoError`, `EmailDuplicadoError`) los
traduce `register()` en `app/modules/usuarios/__init__.py`.
"""
from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Query, status
from fastapi.responses import Response

from app.modules.usuarios.dependencies import DbSession, UserServiceDep
from app.modules.usuarios.schemas import (
    UserSaaSCreate,
    UserSaaSListResponse,
    UserSaaSResponse,
    UserSaaSUpdate,
)
from app.shared.tenancy.dependencies import TenancyServiceDep, tenant_or_404
from app.shared.tenancy.principal import PrincipalDep, require_tenant_admin

router = APIRouter(prefix="/api/saas/tenants/{tenant_id}/users", tags=["users-saas"])


@router.post("", response_model=UserSaaSResponse, status_code=status.HTTP_201_CREATED)
async def create_user(
    tenant_id: int,
    user_data: UserSaaSCreate,
    db: DbSession,
    service: UserServiceDep,
    tenancy: TenancyServiceDep,
    principal: PrincipalDep,
) -> object:
    """Invita un usuario nuevo al tenant (admin/superadmin del tenant)."""
    require_tenant_admin(principal)
    await tenant_or_404(db, tenant_id, tenancy)

    return await service.create(
        db,
        tenant_id=tenant_id,
        email=user_data.email,
        password=user_data.password,
        rol=user_data.rol,
        check_unique=True,
    )


@router.get("", response_model=UserSaaSListResponse)
async def list_users(
    tenant_id: int,
    db: DbSession,
    service: UserServiceDep,
    principal: PrincipalDep,
    skip: Annotated[int, Query(ge=0)] = 0,
    limit: Annotated[int, Query(ge=1, le=100)] = 100,
) -> object:
    """Lista los usuarios del tenant (paginado)."""
    users, total = await service.list_by_tenant(db, tenant_id, skip, limit)
    return UserSaaSListResponse(users=users, total=total)


@router.get("/{user_id}", response_model=UserSaaSResponse)
async def get_user(
    tenant_id: int,
    user_id: int,
    db: DbSession,
    service: UserServiceDep,
    principal: PrincipalDep,
) -> object:
    """Devuelve un usuario del tenant."""
    return await service.obtener(db, tenant_id, user_id)


@router.patch("/{user_id}", response_model=UserSaaSResponse)
async def update_user(
    tenant_id: int,
    user_id: int,
    user_data: UserSaaSUpdate,
    db: DbSession,
    service: UserServiceDep,
    principal: PrincipalDep,
) -> object:
    """Actualiza email, rol o password de un usuario del tenant.

    Cambiar el rol o el password exige admin/superadmin del tenant.
    """
    if user_data.rol is not None or user_data.password is not None:
        require_tenant_admin(principal)

    return await service.update(
        db,
        user_id=user_id,
        tenant_id=tenant_id,
        email=user_data.email,
        rol=user_data.rol,
        password=user_data.password,
        check_unique=True,
    )


@router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_user(
    tenant_id: int,
    user_id: int,
    db: DbSession,
    service: UserServiceDep,
    principal: PrincipalDep,
) -> Response:
    """Elimina un usuario del tenant (admin/superadmin del tenant)."""
    require_tenant_admin(principal)
    await service.delete(db, user_id, tenant_id)
    return Response(status_code=status.HTTP_204_NO_CONTENT)
