"""API endpoints para CRUD de users_saas.

Gestion de usuarios dentro de un tenant, con aislamiento multi-tenant.

Identidad: la inyecta el gateway (`forward_auth` -> auth-service) como
`X-User-Id` / `X-Tenant-Id` / `X-Rol-Saas`; estos handlers la reciben via
`PrincipalDep` y la validan contra el `{tenant_id}` del path. Un request sin
identidad recibe 401 y un `member` no administra usuarios (403).
"""
from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, Query, status
from fastapi.responses import Response
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.db import get_db
from app.crud.user import user_crud
from app.deps.principal import (
    PrincipalDep,
    require_tenant_admin,
    require_tenant_scope,
)
from app.models.tenant import Tenant
from app.schemas.user import (
    UserSaaSCreate,
    UserSaaSListResponse,
    UserSaaSResponse,
    UserSaaSUpdate,
)

router = APIRouter(prefix="/api/saas/tenants/{tenant_id}/users", tags=["users-saas"])

DbSession = Annotated[AsyncSession, Depends(get_db)]


async def _tenant_or_404(db: AsyncSession, tenant_id: int) -> None:
    """Verifica que el tenant exista. 404 explicito en vez de FK error."""
    exists = await db.execute(select(Tenant.id).where(Tenant.id == tenant_id))
    if exists.scalar_one_or_none() is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Tenant no encontrado",
        )


@router.post("", response_model=UserSaaSResponse, status_code=status.HTTP_201_CREATED)
async def create_user(
    tenant_id: int,
    user_data: UserSaaSCreate,
    db: DbSession,
    principal: PrincipalDep,
) -> object:
    """Invita un usuario nuevo al tenant (admin/superadmin del tenant)."""
    require_tenant_scope(principal, tenant_id)
    require_tenant_admin(principal)
    await _tenant_or_404(db, tenant_id)

    existing = await user_crud.get_by_email(db, user_data.email, tenant_id)
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="El email ya existe en este tenant",
        )

    return await user_crud.create(
        db,
        tenant_id=tenant_id,
        email=user_data.email,
        password=user_data.password,
        rol=user_data.rol,
    )


@router.get("", response_model=UserSaaSListResponse)
async def list_users(
    tenant_id: int,
    db: DbSession,
    principal: PrincipalDep,
    skip: Annotated[int, Query(ge=0)] = 0,
    limit: Annotated[int, Query(ge=1, le=100)] = 100,
) -> object:
    """Lista los usuarios del tenant (paginado)."""
    require_tenant_scope(principal, tenant_id)

    users, total = await user_crud.list_by_tenant(db, tenant_id, skip, limit)
    return UserSaaSListResponse(users=users, total=total)


@router.get("/{user_id}", response_model=UserSaaSResponse)
async def get_user(
    tenant_id: int,
    user_id: int,
    db: DbSession,
    principal: PrincipalDep,
) -> object:
    """Devuelve un usuario del tenant."""
    require_tenant_scope(principal, tenant_id)

    user = await user_crud.get_by_id(db, user_id, tenant_id)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Usuario no encontrado",
        )
    return user


@router.patch("/{user_id}", response_model=UserSaaSResponse)
async def update_user(
    tenant_id: int,
    user_id: int,
    user_data: UserSaaSUpdate,
    db: DbSession,
    principal: PrincipalDep,
) -> object:
    """Actualiza email, rol o password de un usuario del tenant.

    Cambiar el rol o el password exige admin/superadmin del tenant.
    """
    require_tenant_scope(principal, tenant_id)
    if user_data.rol is not None or user_data.password is not None:
        require_tenant_admin(principal)

    if user_data.email is not None:
        clash = await user_crud.get_by_email(db, user_data.email, tenant_id)
        if clash and clash.id != user_id:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="El email ya existe en este tenant",
            )

    user = await user_crud.update(
        db,
        user_id=user_id,
        tenant_id=tenant_id,
        email=user_data.email,
        rol=user_data.rol,
        password=user_data.password,
    )
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Usuario no encontrado",
        )
    return user


@router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_user(
    tenant_id: int,
    user_id: int,
    db: DbSession,
    principal: PrincipalDep,
) -> Response:
    """Elimina un usuario del tenant (admin/superadmin del tenant)."""
    require_tenant_scope(principal, tenant_id)
    require_tenant_admin(principal)

    deleted = await user_crud.delete(db, user_id, tenant_id)
    if not deleted:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Usuario no encontrado",
        )
    return Response(status_code=status.HTTP_204_NO_CONTENT)