"""Modelo UserSaaS: usuario dentro de un tenant.

Cada UserSaaS pertenece a UN tenant y tiene un rol local al tenant
(no global al SaaS). El email es unico POR TENANT (un usuario puede
tener cuentas en distintos tenants con el mismo email).
"""
from __future__ import annotations

from typing import TYPE_CHECKING

from sqlalchemy import ForeignKey, String, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.models.base import Base, TimestampMixin

if TYPE_CHECKING:
    from app.models.tenant import Tenant


class UserSaaS(Base, TimestampMixin):
    __tablename__ = "users_saas"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    tenant_id: Mapped[int] = mapped_column(
        ForeignKey("tenants.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    email: Mapped[str] = mapped_column(String(255), nullable=False)
    hashed_password: Mapped[str] = mapped_column(String(255), nullable=False)
    rol: Mapped[str] = mapped_column(String(50), nullable=False, default="member")

    __table_args__ = (
        UniqueConstraint("tenant_id", "email", name="uq_users_saas_tenant_email"),
    )

    tenant: Mapped["Tenant"] = relationship(back_populates="users_saas")

    def __repr__(self) -> str:
        return f"<UserSaaS id={self.id} tenant_id={self.tenant_id} email={self.email!r}>"
