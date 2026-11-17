"""Modelo `Pedido`: orden de compra dentro de un tenant.

Reglas de negocio
-----------------
Un pedido siempre pertenece a UN tenant y nace en estado `pendiente`.

Las reglas de `cantidad` y `total` se validan en tres capas, a proposito:

1. **Pydantic** (`app/modules/pedidos/schemas.py`): rechaza el request antes
   de tocar la BD. Es la capa que ve el cliente de la API (422).
2. **CHECK constraints** (migracion `0005_pedidos`): la BD no acepta una fila
   invalida aunque alguien escriba por fuera de la API.
3. **Este modulo** (`validar()`): misma regla disponible para cualquier
   consumidor interno (tests, scripts, futuros casos de uso).

El ciclo de vida se modela con `EstadoPedido` + `TRANSICIONES_PERMITIDAS`
(`enums.py`): las transiciones viven en el dominio, no en el router.
"""
from __future__ import annotations

from decimal import Decimal

from sqlalchemy import CheckConstraint, ForeignKey, Index, Integer, Numeric, String
from sqlalchemy.orm import Mapped, mapped_column

from app.modules.pedidos.enums import EstadoPedido
from app.shared.db.base import Base, TimestampMixin


class Pedido(Base, TimestampMixin):
    """Orden de compra de un tenant."""

    __tablename__ = "pedidos"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)

    # Discriminador de aislamiento: toda query del repositorio filtra por
    # tenant_id y la politica RLS de `0005_pedidos` lo vuelve a verificar en
    # Postgres.
    tenant_id: Mapped[int] = mapped_column(
        ForeignKey("tenants.id", ondelete="CASCADE"),
        nullable=False,
    )

    cliente: Mapped[str] = mapped_column(String(255), nullable=False)
    cantidad: Mapped[int] = mapped_column(Integer, nullable=False)
    total: Mapped[Decimal] = mapped_column(Numeric(12, 2), nullable=False)
    estado: Mapped[str] = mapped_column(
        String(20), nullable=False, default=EstadoPedido.PENDIENTE.value, server_default="pendiente"
    )

    __table_args__ = (
        CheckConstraint("cantidad > 0", name="ck_pedidos_cantidad_positiva"),
        CheckConstraint("total > 0", name="ck_pedidos_total_positivo"),
        CheckConstraint(
            "estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado')",
            name="ck_pedidos_estado_valido",
        ),
        # El prefijo `tenant_id` sirve para las consultas del repositorio;
        # `estado` acelera el filtro opcional por estado de la bandeja.
        Index("ix_pedidos_tenant_estado", "tenant_id", "estado"),
    )

    def validar(self) -> None:
        """Valida las invariantes del agregado. Pensado para uso interno."""
        if self.cantidad <= 0:
            raise ValueError("La cantidad debe ser mayor que cero")
        if self.total <= 0:
            raise ValueError("El total debe ser mayor que cero")

    def __repr__(self) -> str:
        return (
            f"<Pedido id={self.id} tenant_id={self.tenant_id} "
            f"cantidad={self.cantidad} total={self.total} estado={self.estado!r}>"
        )
