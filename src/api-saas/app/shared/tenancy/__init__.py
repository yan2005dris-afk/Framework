"""Plataforma multi-tenant: identidad, scope y modelo `Tenant`.

Re-exports de conveniencia. Para evitar ciclos, `app.shared.db.session` importa
`app.shared.tenancy.principal` directamente.
"""
from app.shared.tenancy.dependencies import TenancyServiceDep, get_tenancy_service, tenant_or_404
from app.shared.tenancy.models import Tenant
from app.shared.tenancy.ports import TenancyPort
from app.shared.tenancy.principal import (
    Principal,
    PrincipalDep,
    get_principal,
    principal_from_headers,
    require_tenant_admin,
    require_tenant_scope,
)
from app.shared.tenancy.service import TenancyService, tenancy_service

__all__ = [
    "Principal",
    "PrincipalDep",
    "Tenant",
    "TenancyPort",
    "TenancyService",
    "TenancyServiceDep",
    "get_principal",
    "get_tenancy_service",
    "principal_from_headers",
    "require_tenant_admin",
    "require_tenant_scope",
    "tenancy_service",
    "tenant_or_404",
]
