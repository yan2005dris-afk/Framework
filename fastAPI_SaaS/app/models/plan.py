"""Modelo Plan.

Representa un plan de suscripción disponible en el SaaS.
"""
from __future__ import annotations

from decimal import Decimal
from typing import TYPE_CHECKING, Optional

from sqlalchemy import JSON, Numeric, String
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.models.base import Base, TimestampMixin

if TYPE_CHECKING:
    from app.models.subscription import Subscription


class Plan(Base, TimestampMixin):
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

    # Relación inversa: un Plan tiene muchas Subscriptions
    subscriptions: Mapped[list["Subscription"]] = relationship(
        back_populates="plan",
        cascade="all, delete-orphan",
    )

    def __repr__(self) -> str:
        return f"<Plan id={self.id} slug={self.slug!r} name={self.name!r}>"
