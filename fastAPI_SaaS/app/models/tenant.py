"""Modelo Tenant.

Representa uninquilino del SaaS. Todos los modelos SaaS (excepto el mismo
Tenant) tienen `tenant_id` FK -> `tenants.id` para garantizar el aislamiento
de datos entreinquilinos a nivel BD.

U3-T2 agrega la capa de auth que vincula el JWT con un `tenant_id` y aplica
filtros `WHERE tenant_id = :current_tenant` en cada query.
"""
from __future__ import annotations

from typing import TYPE_CHECKING

from sqlalchemy import String
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.models.base import Base, TimestampMixin

if TYPE_CHECKING:
    from app.models.user import UserSaaS


class Tenant(Base, TimestampMixin):
    __tablename__ = "tenants"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    slug: Mapped[str] = mapped_column(String(64), unique=True, nullable=False, index=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False)

    # Relacion inversa: un Tenant tiene muchos Users SaaS.
    users_saas: Mapped[list["UserSaaS"]] = relationship(
        back_populates="tenant",
        cascade="all, delete-orphan",
        passive_deletes=True,
    )

    def __repr__(self) -> str:
        return f"<Tenant id={self.id} slug={self.slug!r} name={self.name!r}>"
