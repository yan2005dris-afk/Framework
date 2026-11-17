"""Modulo de usuarios: registro en la app y traduccion de errores a HTTP."""
from __future__ import annotations

from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse

from app.modules.usuarios.exceptions import EmailDuplicadoError, UsuarioNoEncontradoError
from app.modules.usuarios.router import router

__all__ = ["register", "router"]


async def _usuario_no_encontrado(_request: Request, _exc: Exception) -> JSONResponse:
    """404 uniforme: no se filtra si el usuario existe en otro tenant."""
    return JSONResponse(
        status_code=status.HTTP_404_NOT_FOUND,
        content={"detail": "Usuario no encontrado"},
    )


async def _email_duplicado(_request: Request, exc: Exception) -> JSONResponse:
    """400 cuando el email ya existe dentro del tenant."""
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={"detail": str(exc)},
    )


def register(app: FastAPI) -> None:
    """Monta el router y registra los handlers de errores de dominio."""
    app.include_router(router)
    app.add_exception_handler(UsuarioNoEncontradoError, _usuario_no_encontrado)
    app.add_exception_handler(EmailDuplicadoError, _email_duplicado)
