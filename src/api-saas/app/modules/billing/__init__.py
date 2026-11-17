"""Modulo de billing: registro en la app y traduccion de errores a HTTP."""
from __future__ import annotations

from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse

from app.modules.billing.exceptions import (
    PlanNoEncontradoError,
    SuscripcionExistenteError,
    SuscripcionNoEncontradaError,
)
from app.modules.billing.router import router

__all__ = ["register", "router"]


async def _plan_no_encontrado(_request: Request, _exc: Exception) -> JSONResponse:
    return JSONResponse(
        status_code=status.HTTP_404_NOT_FOUND,
        content={"detail": "Plan no encontrado"},
    )


async def _suscripcion_no_encontrada(_request: Request, _exc: Exception) -> JSONResponse:
    return JSONResponse(
        status_code=status.HTTP_404_NOT_FOUND,
        content={"detail": "Suscripción no encontrada"},
    )


async def _suscripcion_existente(_request: Request, exc: Exception) -> JSONResponse:
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={"detail": str(exc)},
    )


def register(app: FastAPI) -> None:
    """Monta el router y registra los handlers de errores de dominio."""
    app.include_router(router)
    app.add_exception_handler(PlanNoEncontradoError, _plan_no_encontrado)
    app.add_exception_handler(SuscripcionNoEncontradaError, _suscripcion_no_encontrada)
    app.add_exception_handler(SuscripcionExistenteError, _suscripcion_existente)
