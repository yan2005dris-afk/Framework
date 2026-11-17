"""Excepciones de dominio del modulo `billing`.

La capa HTTP las traduce en `register()` (400 / 404). El servicio y el
repositorio no conocen FastAPI.
"""
from __future__ import annotations


class PlanNoEncontradoError(LookupError):
    """No existe un plan con ese id."""

    def __init__(self, plan_id: int) -> None:
        self.plan_id = plan_id
        super().__init__(f"Plan {plan_id} no encontrado")


class SuscripcionNoEncontradaError(LookupError):
    """El tenant no tiene una suscripcion."""

    def __init__(self, tenant_id: int) -> None:
        self.tenant_id = tenant_id
        super().__init__(f"Suscripcion del tenant {tenant_id} no encontrada")


class SuscripcionExistenteError(ValueError):
    """El tenant ya tiene una suscripcion (UNIQUE en subscriptions.tenant_id)."""

    def __init__(self, tenant_id: int) -> None:
        self.tenant_id = tenant_id
        super().__init__("El tenant ya tiene una suscripción activa")
