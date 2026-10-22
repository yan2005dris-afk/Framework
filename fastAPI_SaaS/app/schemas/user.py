"""Schemas Pydantic para UserSaaS.

DTOs para creación, actualización y respuesta de usuarios SaaS.
"""
from __future__ import annotations

from datetime import datetime
from typing import Optional

from pydantic import BaseModel, ConfigDict, EmailStr, Field


class UserSaaSBase(BaseModel):
    """Base schema con campos comunes."""
    email: EmailStr
    rol: str = Field(default="member", description="Rol del usuario: admin, member, viewer")


class UserSaaSCreate(UserSaaSBase):
    """Schema para crear un usuario SaaS."""
    password: str = Field(..., min_length=6, description="Password del usuario (mínimo 6 caracteres)")


class UserSaaSUpdate(BaseModel):
    """Schema para actualizar un usuario SaaS."""
    email: Optional[EmailStr] = None
    rol: Optional[str] = Field(None, description="Rol del usuario: admin, member, viewer")
    password: Optional[str] = Field(None, min_length=6, description="Nuevo password (si se desea cambiar)")


class UserSaaSResponse(UserSaaSBase):
    """Schema para respuesta de usuario SaaS."""

    model_config = ConfigDict(from_attributes=True)

    id: int
    tenant_id: int
    created_at: datetime
    updated_at: datetime


class UserSaaSListResponse(BaseModel):
    """Schema para lista de usuarios SaaS."""
    users: list[UserSaaSResponse]
    total: int
