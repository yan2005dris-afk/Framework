"""Schemas Pydantic v2 para el recurso `Pedido`.

Aqui vive la primera barrera de validacion (la que ve el cliente de la API):

- `cantidad` y `total` usan `Field(gt=0)`: un pedido con cantidad o total
  menor o igual a cero se rechaza con **422** antes de ejecutar una sola
  query. Es exactamente la regla que pide la actividad, aplicada en el borde.
- `estado` usa el enum `EstadoPedido`: un estado inexistente tambien es 422,
  con la lista de valores validos en el detalle del error.

Los schemas de respuesta son independientes de los de entrada: el cliente no
puede mandar `id`, `tenant_id` ni `estado` al crear un pedido.
"""
from __future__ import annotations

from datetime import datetime
from decimal import Decimal

from pydantic import BaseModel, ConfigDict, Field

from app.modules.pedidos.enums import EstadoPedido


class PedidoCreate(BaseModel):
    """Payload de `POST /api/{tenant_id}/pedidos`."""

    model_config = ConfigDict(str_strip_whitespace=True)

    cliente: str = Field(
        ...,
        min_length=1,
        max_length=255,
        description="Nombre del cliente final",
        examples=["Ana Perez"],
    )
    cantidad: int = Field(
        ...,
        gt=0,
        description="Unidades pedidas. Debe ser mayor que cero.",
        examples=[3],
    )
    total: Decimal = Field(
        ...,
        gt=0,
        max_digits=12,
        decimal_places=2,
        description="Importe total del pedido. Debe ser mayor que cero.",
        examples=["149.90"],
    )


class PedidoEstadoUpdate(BaseModel):
    """Payload de `PUT /api/{tenant_id}/pedidos/{pedido_id}/estado`."""

    estado: EstadoPedido = Field(
        ...,
        description="Estado destino del pedido",
        examples=["preparando"],
    )


class PedidoResponse(BaseModel):
    """Representacion publica de un pedido."""

    model_config = ConfigDict(from_attributes=True)

    id: int
    tenant_id: int
    cliente: str
    cantidad: int
    total: float
    estado: EstadoPedido
    created_at: datetime
    updated_at: datetime


class PedidoListResponse(BaseModel):
    """Respuesta paginada de `GET /api/{tenant_id}/pedidos`."""

    pedidos: list[PedidoResponse]
    total: int
