"""Excepciones de dominio del modulo `usuarios`.

La capa HTTP las traduce en `register()` (400 / 404). El servicio y el
repositorio no conocen FastAPI.
"""
from __future__ import annotations


class UsuarioNoEncontradoError(LookupError):
    """No existe un usuario con ese id dentro del tenant."""

    def __init__(self, user_id: int) -> None:
        self.user_id = user_id
        super().__init__(f"Usuario {user_id} no encontrado")


class EmailDuplicadoError(ValueError):
    """El email ya esta registrado dentro del mismo tenant."""

    def __init__(self, email: str) -> None:
        self.email = email
        super().__init__("El email ya existe en este tenant")
