"""Estados y transiciones del ciclo de vida de un pedido.

Vive separado de `models.py` y `schemas.py` para que ambos puedan importarlo
sin crear un ciclo (`models` <-> `schemas`). No importa nada del proyecto.
"""
from __future__ import annotations

from enum import StrEnum


class EstadoPedido(StrEnum):
    """Estados posibles de un pedido.

    `StrEnum` (Python 3.11+) permite comparar contra strings y, a la vez,
    que FastAPI/Pydantic lo documente como enum en el OpenAPI.
    """

    PENDIENTE = "pendiente"
    PREPARANDO = "preparando"
    ENVIADO = "enviado"
    ENTREGADO = "entregado"
    CANCELADO = "cancelado"


#: Maquina de estados: desde -> hacia donde se puede ir.
#: `entregado` y `cancelado` son terminales.
TRANSICIONES_PERMITIDAS: dict[EstadoPedido, frozenset[EstadoPedido]] = {
    EstadoPedido.PENDIENTE: frozenset({EstadoPedido.PREPARANDO, EstadoPedido.CANCELADO}),
    EstadoPedido.PREPARANDO: frozenset({EstadoPedido.ENVIADO, EstadoPedido.CANCELADO}),
    EstadoPedido.ENVIADO: frozenset({EstadoPedido.ENTREGADO, EstadoPedido.CANCELADO}),
    EstadoPedido.ENTREGADO: frozenset(),
    EstadoPedido.CANCELADO: frozenset(),
}


def puede_transicionar(actual: EstadoPedido, nuevo: EstadoPedido) -> bool:
    """`True` si el pedido puede pasar de `actual` a `nuevo`."""
    return nuevo in TRANSICIONES_PERMITIDAS.get(actual, frozenset())
