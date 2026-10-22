"""Schemas Pydantic para Plan y Subscription."""
from __future__ import annotations

from datetime import datetime
from decimal import Decimal
from typing import Optional

from pydantic import BaseModel, ConfigDict, Field


class PlanBase(BaseModel):
    """Base schema con campos comunes."""
    slug: str = Field(..., description="Slug único del plan")
    name: str = Field(..., description="Nombre del plan")
    description: Optional[str] = Field(None, description="Descripción del plan")
    price_monthly: Decimal = Field(..., ge=0, description="Precio mensual")
    price_yearly: Optional[Decimal] = Field(None, ge=0, description="Precio anual")
    features: Optional[list[str]] = Field(None, description="Lista de features")


class PlanCreate(PlanBase):
    """Schema para crear un plan."""
    pass


class PlanResponse(PlanBase):
    """Schema para respuesta de plan."""

    model_config = ConfigDict(from_attributes=True)

    id: int
    created_at: datetime
    updated_at: datetime


class SubscriptionBase(BaseModel):
    """Base schema para subscription."""
    status: str = Field(default="active", description="Estado: active, cancelled, past_due")
    cancel_at_period_end: bool = Field(default=False)


class SubscriptionCreate(SubscriptionBase):
    """Schema para crear una suscripción."""
    plan_id: int = Field(..., description="ID del plan a suscribirse")


class SubscriptionResponse(SubscriptionBase):
    """Schema para respuesta de subscription."""

    model_config = ConfigDict(from_attributes=True)

    id: int
    tenant_id: int
    plan_id: int
    stripe_subscription_id: Optional[str]
    current_period_start: Optional[datetime]
    current_period_end: Optional[datetime]
    created_at: datetime
    updated_at: datetime
