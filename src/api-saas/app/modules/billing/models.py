"""Modelos del modulo billing: `Plan` y `Subscription`.

`Plan` es el catalogo global del SaaS (no tenant-scoped).
`Subscription` vincula un tenant con su plan (1:1 por el UNIQUE de `tenant_id`).

Las relaciones son unidireccionales hacia `Tenant` (que vive en
`app.shared.tenancy.models`) para no invertir la dependencia
shared -> modules.
"""
from __future__ import annotations

from datetime import datetime
from decimal import Decimal
from typing import Optional

from sqlalchemy import JSON, DateTime, ForeignKey, Numeric, String
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.shared.db.base import Base, TimestampMixin
from app.shared.tenancy.models import Tenant


class Plan(Base, TimestampMixin):
    """Plan de suscripcion disponible en el SaaS."""

    __tablename__ = "plans"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    slug: Mapped[str] = mapped_column(String(64), unique=True, nullable=False, index=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[Optional[str]] = mapped_column(String(1000), nullable=True)
    # Numeric -> Decimal en Python (no float): los precios son dinero.
    price_monthly: Mapped[Decimal] = mapped_column(Numeric(10, 2), nullable=False)
    price_yearly: Mapped[Optional[Decimal]] = mapped_column(Numeric(10, 2), nullable=True)
    # JSON: lista de features del plan (array de strings).
    features: Mapped[Optional[list[str]]] = mapped_column(JSON, nullable=True)

    # Un Plan agrupa muchas Subscriptions.
    subscriptions: Mapped[list["Subscription"]] = relationship(back_populates="plan")

    def __repr__(self) -> str:
        return f"<Plan id={self.id} slug={self.slug!r} name={self.name!r}>"


class Subscription(Base, TimestampMixin):
    """Suscripcion de un tenant a un plan."""

    __tablename__ = "subscriptions"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    tenant_id: Mapped[int] = mapped_column(
        ForeignKey("tenants.id", ondelete="CASCADE"),
        nullable=False,
        unique=True,  # Un tenant solo puede tener una suscripcion activa
    )
    plan_id: Mapped[int] = mapped_column(
        ForeignKey("plans.id", ondelete="CASCADE"),
        nullable=False,
    )
    status: Mapped[str] = mapped_column(
        String(50), nullable=False, default="active"
    )  # active, cancelled, past_due
    stripe_subscription_id: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    current_period_start: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    current_period_end: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    cancel_at_period_end: Mapped[bool] = mapped_column(default=False, nullable=False)

    tenant: Mapped[Tenant] = relationship()
    plan: Mapped[Plan] = relationship(back_populates="subscriptions")

    def __repr__(self) -> str:
        return f"<Subscription id={self.id} tenant_id={self.tenant_id} status={self.status!r}>"
