"""Modelo `Tenant` (plataforma, no feature).

Representa un inquilino del SaaS. Es la raiz del aislamiento: todas las tablas
tenant-scoped (users_saas, subscriptions, pedidos) tienen `tenant_id` FK ->
`tenants.id`.

Vive en `shared` y no en un modulo de feature porque no es propiedad de
pedidos, usuarios ni billing: es la entidad de plataforma que los tres
referencian y sobre la que actua la politica RLS.

Las relaciones inversas (`users_saas`, `subscription`) viven en los modelos
hijos con `relationship()` unidireccional. Asi `shared` no necesita importar
`app.modules` (lo verifica el contrato `forbidden` de import-linter) y el
borrado en cascada sigue garantizado por los `ON DELETE CASCADE` del schema.
"""
from __future__ import annotations

from sqlalchemy import String
from sqlalchemy.orm import Mapped, mapped_column

from app.shared.db.base import Base, TimestampMixin


class Tenant(Base, TimestampMixin):
    __tablename__ = "tenants"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    slug: Mapped[str] = mapped_column(String(64), unique=True, nullable=False, index=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False)

    def __repr__(self) -> str:
        return f"<Tenant id={self.id} slug={self.slug!r} name={self.name!r}>"
