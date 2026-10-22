"""Modelos SQLAlchemy 2.x.

Convencion multi-tenant: cualquier tabla SaaS (excepto `tenants` misma)
hereda `TimestampMixin` y carga con un `tenant_id` FK -> `tenants.id`.

El campo `tenant_id` es el discriminador de aislamiento de datos entre
inquilinos; las queries de lectura SIEMPRE filtran por `tenant_id`
proveniente del claim JWT (U3-T2 agrega la capa de auth + tenant claim).

Importar este paquete registra TODOS los modelos en `Base.metadata`
(debido a los side-effects del import). Sin este import, alembic
no podra autogenerar la primera migracion.
"""
# Importar aca cada modelo nuevo para registrarlo en Base.metadata.
# El orden importa: Tenant primero porque UserSaaS.tenants FK -> tenants.id.
from app.models.base import Base, TimestampMixin  # noqa: F401
from app.models.tenant import Tenant  # noqa: F401
from app.models.user import UserSaaS  # noqa: F401

__all__ = ["Base", "TimestampMixin", "Tenant", "UserSaaS"]
