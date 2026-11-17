"""Modulo de pedidos: registro en la app y traduccion de errores a HTTP.

`main.py` solo hace `pedidos.register(app)`. El modulo es dueno de su router y
de sus exception handlers, asi que las URLs, los codigos y los mensajes viven
juntos y no se dispersan por el composition root.
"""
from __future__ import annotations

from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse

from app.modules.pedidos.exceptions import PedidoNoEncontradoError, TransicionInvalidaError
from app.modules.pedidos.router import router

__all__ = ["register", "router"]


async def _transicion_invalida(_request: Request, exc: Exception) -> JSONResponse:
    """400 con el mensaje del dominio (incluye los estados permitidos)."""
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={"detail": str(exc)},
    )


async def _pedido_no_encontrado(_request: Request, _exc: Exception) -> JSONResponse:
    """404 uniforme: no se filtra si el pedido existe en otro tenant."""
    return JSONResponse(
        status_code=status.HTTP_404_NOT_FOUND,
        content={"detail": "Pedido no encontrado"},
    )


def register(app: FastAPI) -> None:
    """Monta el router y registra los handlers de errores de dominio."""
    app.include_router(router)
    app.add_exception_handler(TransicionInvalidaError, _transicion_invalida)
    app.add_exception_handler(PedidoNoEncontradoError, _pedido_no_encontrado)
