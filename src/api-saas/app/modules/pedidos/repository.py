"""Repositorio de `Pedido` con aislamiento por tenant.

Convenciones (iguales a `app/modules/usuarios/repository.py`):

- **Toda** query filtra por `tenant_id`. El aislamiento no depende solo de
  RLS: un pedido de otro tenant nunca se lee, se modifica ni se borra.
- Se usa `flush()` y no `commit()`: la transaccion del request la cierra
  `get_tenant_db`, para que el GUC `app.current_tenant` (SET LOCAL) siga
  activo durante toda la operacion.
- La transicion de estado se valida en el dominio
  (`app/modules/pedidos/enums.py`), no en el router ni en el repositorio:
  aca solo se persiste el resultado ya validado por el servicio.
"""
from __future__ import annotations

from decimal import Decimal
from typing import Optional, Protocol

from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.pedidos.enums import EstadoPedido
from app.modules.pedidos.models import Pedido


class PedidoRepositoryPort(Protocol):
    """Contrato de persistencia que consume `PedidoService`."""

    async def create(
        self,
        db: AsyncSession,
        tenant_id: int,
        cliente: str,
        cantidad: int,
        total: Decimal,
    ) -> Pedido: ...

    async def get_by_id(
        self, db: AsyncSession, pedido_id: int, tenant_id: int
    ) -> Optional[Pedido]: ...

    async def list_by_tenant(
        self,
        db: AsyncSession,
        tenant_id: int,
        skip: int = 0,
        limit: int = 100,
        estado: Optional[EstadoPedido] = None,
    ) -> tuple[list[Pedido], int]: ...

    async def cambiar_estado(
        self,
        db: AsyncSession,
        pedido: Pedido,
        nuevo_estado: EstadoPedido,
    ) -> Pedido: ...

    async def delete(self, db: AsyncSession, pedido_id: int, tenant_id: int) -> bool: ...


class PedidoRepository:
    """Operaciones de persistencia para `Pedido`."""

    async def create(
        self,
        db: AsyncSession,
        tenant_id: int,
        cliente: str,
        cantidad: int,
        total: Decimal,
    ) -> Pedido:
        """Crea un pedido en estado `pendiente` dentro del tenant."""
        pedido = Pedido(
            tenant_id=tenant_id,
            cliente=cliente,
            cantidad=cantidad,
            total=total,
            estado=EstadoPedido.PENDIENTE.value,
        )
        pedido.validar()  # defensa en profundidad: la regla tambien vive aca
        db.add(pedido)
        await db.flush()
        await db.refresh(pedido)
        return pedido

    async def get_by_id(
        self, db: AsyncSession, pedido_id: int, tenant_id: int
    ) -> Optional[Pedido]:
        """Busca un pedido por ID dentro del tenant."""
        result = await db.execute(
            select(Pedido).where(
                Pedido.id == pedido_id,
                Pedido.tenant_id == tenant_id,
            )
        )
        return result.scalar_one_or_none()

    async def list_by_tenant(
        self,
        db: AsyncSession,
        tenant_id: int,
        skip: int = 0,
        limit: int = 100,
        estado: Optional[EstadoPedido] = None,
    ) -> tuple[list[Pedido], int]:
        """Lista los pedidos del tenant, con filtro opcional por estado."""
        query = select(Pedido).where(Pedido.tenant_id == tenant_id)
        count_query = select(func.count()).select_from(Pedido).where(
            Pedido.tenant_id == tenant_id
        )

        if estado is not None:
            query = query.where(Pedido.estado == estado.value)
            count_query = count_query.where(Pedido.estado == estado.value)

        result = await db.execute(
            query.order_by(Pedido.created_at.desc(), Pedido.id.desc())
            .offset(skip)
            .limit(limit)
        )
        pedidos = list(result.scalars().all())
        total = await db.scalar(count_query)
        return pedidos, int(total or 0)

    async def cambiar_estado(
        self,
        db: AsyncSession,
        pedido: Pedido,
        nuevo_estado: EstadoPedido,
    ) -> Pedido:
        """Persiste el nuevo estado de un pedido ya resuelto por el servicio.

        La validacion de la transicion la hace `PedidoService` ANTES de llamar
        aca, para que el repositorio no dependa de reglas de dominio.
        """
        pedido.estado = nuevo_estado.value
        await db.flush()
        await db.refresh(pedido)
        return pedido

    async def delete(self, db: AsyncSession, pedido_id: int, tenant_id: int) -> bool:
        """Elimina un pedido del tenant. `False` si no existe."""
        pedido = await self.get_by_id(db, pedido_id, tenant_id)
        if pedido is None:
            return False

        await db.delete(pedido)
        await db.flush()
        return True


#: Instancia sin estado: se inyecta via `PedidoServiceDep`.
pedido_repository = PedidoRepository()
