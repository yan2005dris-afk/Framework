"""Base declarativa + mixin de timestamps para todas las tablas SaaS."""
from __future__ import annotations

from datetime import datetime

from sqlalchemy import DateTime, func
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column


class Base(DeclarativeBase):
    """Base declarativa. Todos los modelos heredan de aca + TimestampMixin."""


class TimestampMixin:
    """created_at + updated_at con default de servidor y onupdate.

    Mantener en todas las tablas SaaS salvo casos justificados (no aplica
    a tablas de solo-lectura como auditoria historica).
    """

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
        nullable=False,
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )
