"""Dependencies de autorizacion del servicio SaaS."""
from app.deps.principal import (  # noqa: F401
    Principal,
    PrincipalDep,
    get_principal,
    principal_from_headers,
    require_tenant_admin,
    require_tenant_scope,
)

__all__ = [
    "Principal",
    "PrincipalDep",
    "get_principal",
    "principal_from_headers",
    "require_tenant_admin",
    "require_tenant_scope",
]