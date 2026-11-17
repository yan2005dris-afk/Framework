"""Excepciones de dominio del modulo `pedidos`.

La capa HTTP las traduce en `register()` (400 / 404). El servicio y el
repositorio no conocen FastAPI.
"""
from __future__ import annotations

from app.modules.pedidos.enums import EstadoPedido, TRANSICIONES_PERMITIDAS


class TransicionInvalidaError(ValueError):
    """El pedido no puede pasar de su estado actual al estado pedido."""

    def __init__(self, actual: EstadoPedido, nuevo: EstadoPedido) -> None:
        self.actual = actual
        self.nuevo = nuevo
        permitidos = TRANSICIONES_PERMITIDAS.get(actual, frozenset())
        detalle = ", ".join(sorted(e.value for e in permitidos)) or "ninguno (estado terminal)"
        super().__init__(
            f"Transicion invalida: {actual.value} -> {nuevo.value}. "
            f"Desde {actual.value} solo se puede pasar a: {detalle}."
        )


class PedidoNoEncontradoError(LookupError):
    """No existe un pedido con ese id dentro del tenant."""

    def __init__(self, pedido_id: int) -> None:
        self.pedido_id = pedido_id
        super().__init__(f"Pedido {pedido_id} no encontrado")
