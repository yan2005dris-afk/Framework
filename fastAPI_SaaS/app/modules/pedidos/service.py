"""Casos de uso del modulo `pedidos`.

El servicio concentra las reglas que no son simples queries: resolver la
entidad o fallar con un error de dominio, y validar la maquina de estados
antes de persistir. No conoce FastAPI ni HTTP.

`flush()` vs `commit()`: el servicio nunca commitea. La transaccion del
request la cierra `get_tenant_db` (ver `app/shared/db/session.py`).
"""
from __future__ import annotations

from decimal import Decimal
from typing import Optional

from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.pedidos.enums import EstadoPedido, puede_transicionar
from app.modules.pedidos.exceptions import PedidoNoEncontradoError, TransicionInvalidaError
from app.modules.pedidos.models import Pedido
from app.modules.pedidos.repository import PedidoRepositoryPort, pedido_repository


class PedidoService:
    """Operaciones de negocio sobre pedidos de un tenant."""

    def __init__(self, repository: PedidoRepositoryPort = pedido_repository) -> None:
        self._repository = repository

    async def crear(
        self,
        db: AsyncSession,
        tenant_id: int,
        *,
        cliente: str,
        cantidad: int,
        total: Decimal,
    ) -> Pedido:
        """Crea un pedido en estado `pendiente`."""
        return await self._repository.create(
            db,
            tenant_id=tenant_id,
            cliente=cliente,
            cantidad=cantidad,
            total=total,
        )

    async def listar(
        self,
        db: AsyncSession,
        tenant_id: int,
        *,
        skip: int = 0,
        limit: int = 100,
        estado: Optional[EstadoPedido] = None,
    ) -> tuple[list[Pedido], int]:
        """Lista pedidos del tenant y el total para paginar."""
        return await self._repository.list_by_tenant(
            db, tenant_id, skip=skip, limit=limit, estado=estado
        )

    async def obtener(self, db: AsyncSession, tenant_id: int, pedido_id: int) -> Pedido:
        """Devuelve un pedido del tenant o lanza `PedidoNoEncontradoError`."""
        pedido = await self._repository.get_by_id(db, pedido_id, tenant_id)
        if pedido is None:
            raise PedidoNoEncontradoError(pedido_id)
        return pedido

    async def cambiar_estado(
        self,
        db: AsyncSession,
        tenant_id: int,
        pedido_id: int,
        nuevo_estado: EstadoPedido,
    ) -> Pedido:
        """Aplica una transicion de estado valida.

        Lanza `PedidoNoEncontradoError` (404) si el pedido no pertenece al
        tenant y `TransicionInvalidaError` (400) si la transicion no esta
        permitida. `entregado` y `cancelado` son terminales.
        """
        pedido = await self.obtener(db, tenant_id, pedido_id)

        actual = EstadoPedido(pedido.estado)
        if not puede_transicionar(actual, nuevo_estado):
            raise TransicionInvalidaError(actual, nuevo_estado)

        return await self._repository.cambiar_estado(db, pedido, nuevo_estado)

    async def eliminar(self, db: AsyncSession, tenant_id: int, pedido_id: int) -> None:
        """Elimina un pedido del tenant o lanza `PedidoNoEncontradoError`."""
        deleted = await self._repository.delete(db, pedido_id, tenant_id)
        if not deleted:
            raise PedidoNoEncontradoError(pedido_id)


#: Instancia sin estado: se inyecta via `PedidoServiceDep`.
pedido_service = PedidoService()
