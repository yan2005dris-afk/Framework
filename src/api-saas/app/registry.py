"""Registro de todos los modelos en `Base.metadata`.

Unico consumidor: `alembic/env.py` (autogenerate) y cualquier script que
necesite el schema completo. Importar `app.registry` ejecuta los imports de
todos los modelos, incluyendo los de cada bounded context.

`app.registry` es el UNICO lugar del arbol que conoce los tres modulos a la
vez; ni `app.shared` ni un modulo de feature importan a otro.
"""
from app.modules.billing.models import Plan, Subscription
from app.modules.pedidos.models import Pedido
from app.modules.usuarios.models import UserSaaS
from app.shared.db.base import Base, TimestampMixin
from app.shared.tenancy.models import Tenant

__all__ = [
    "Base",
    "TimestampMixin",
    "Pedido",
    "Plan",
    "Subscription",
    "Tenant",
    "UserSaaS",
]
