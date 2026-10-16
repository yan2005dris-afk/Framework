"""Modelos SQLAlchemy 2.x.

Convencion multi-tenant: cualquier tabla SaaS (excepto `tenants` misma)
hereda `TimestampMixin` y carga con un `tenant_id` FK -> `tenants.id`.

El campo `tenant_id` es el discriminador de aislamiento de datos entre
inquilinos; las queries de lectura SIEMPRE filtran por `tenant_id`
proveniente del claim JWT (U3-T2 agrega la capa de auth + tenant claim).
"""
