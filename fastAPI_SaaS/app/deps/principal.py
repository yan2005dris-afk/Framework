"""Identidad del request y guards de autorizacion multi-tenant.

De donde viene la identidad
---------------------------
El unico emisor/validador de JWTs del sistema es `auth-service` (GA6). El
gateway (Caddy) corre `forward_auth` contra `POST /auth/validate`, y cuando
el token es valido copia al upstream los headers que ese handler devuelve:

    X-User-Id, X-User-Email, X-User-Rol, X-Tenant-Id, X-Rol-Saas

(fuentes: `gateway/Caddyfile` bloque `copy_headers` y
`auth-service/internal/handlers/auth.go`, handler `Validate`).

Este modulo traduce esos headers a un `Principal`. NO revalida el JWT: esa
responsabilidad vive en el auth-service, y duplicarla aqui seria una segunda
fuente de verdad.

Como se aplica
--------------
- `app/middleware/tenant.py` parsea los headers UNA vez por request y deja el
  `Principal` en `request.state.principal` (o `None` si no hay identidad).
- Los endpoints se inyectan con `PrincipalDep` y se protegen con
  `require_tenant_scope` / `require_tenant_admin`.

Un request sin `X-User-Id` no tiene identidad: `get_principal` responde 401.
Nunca se cae a un tenant por defecto.
"""
from __future__ import annotations

from dataclasses import dataclass
from typing import Annotated

from fastapi import Depends, HTTPException, Request, status

HEADER_USER_ID = "X-User-Id"
HEADER_USER_EMAIL = "X-User-Email"
HEADER_USER_ROL = "X-User-Rol"
HEADER_TENANT_ID = "X-Tenant-Id"
HEADER_ROL_SAAS = "X-Rol-Saas"

#: Roles SaaS que pueden administrar los usuarios de su tenant.
TENANT_ADMIN_ROLES = ("admin", "superadmin")


@dataclass(frozen=True, slots=True)
class Principal:
    """Identidad del request tal como la dejo el gateway."""

    user_id: str
    email: str
    rol: str = ""
    tenant_id: int | None = None
    rol_saas: str = ""

    @property
    def is_superadmin(self) -> bool:
        return self.rol_saas == "superadmin"

    @property
    def is_tenant_admin(self) -> bool:
        return self.rol_saas in TENANT_ADMIN_ROLES


def _optional_int(raw: str | None) -> int | None:
    """Parsea un header numerico. `''` o no numerico -> None.

    El gateway siempre manda `X-Tenant-Id` (puede venir como `'0'` para
    tokens B2C, ver `itoa(claims.TenantID)`), asi que un header vacio o
    ausente significa "este token no pertenece a ningun tenant SaaS".
    """
    if raw is None or raw == "":
        return None
    try:
        return int(raw)
    except ValueError:
        return None


def principal_from_headers(request: Request) -> Principal | None:
    """Construye el `Principal` desde los headers del gateway.

    Devuelve `None` cuando el request no trae identidad (`X-User-Id` vacio):
    un request sin forward_auth NO es un usuario anonimo valido, es un
    request que llego sin pasar por el gateway.
    """
    user_id = request.headers.get(HEADER_USER_ID, "").strip()
    if not user_id:
        return None

    tenant_id = _optional_int(request.headers.get(HEADER_TENANT_ID))
    if tenant_id == 0:
        # El gateway serializa `claims.TenantID` con itoa: los tokens B2C
        # viajan como tenant 0, que significa "sin tenant SaaS".
        tenant_id = None

    return Principal(
        user_id=user_id,
        email=request.headers.get(HEADER_USER_EMAIL, "").strip(),
        rol=request.headers.get(HEADER_USER_ROL, "").strip(),
        tenant_id=tenant_id,
        rol_saas=request.headers.get(HEADER_ROL_SAAS, "").strip(),
    )


async def get_principal(request: Request) -> Principal:
    """Dependency: identidad autenticada del request. 401 si no hay."""
    principal = getattr(request.state, "principal", None)
    if principal is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="No autenticado: el gateway no inyecto identidad",
        )
    return principal


#: Dependency lista para usar en firmas: `principal: PrincipalDep`.
PrincipalDep = Annotated[Principal, Depends(get_principal)]


def require_tenant_scope(principal: Principal, tenant_id: int) -> int:
    """Verifica que el principal pueda operar sobre `tenant_id`.

    `superadmin` (rol global del SaaS) puede operar sobre cualquier tenant.
    Cualquier otro rol solo sobre su propio tenant. Devuelve el `tenant_id`
    efectivo para que el handler lo reuse sin volver a leerlo del path.
    """
    if principal.tenant_id != tenant_id and not principal.is_superadmin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Solo puedes operar sobre tu propio tenant",
        )
    return tenant_id


def require_tenant_admin(principal: Principal) -> None:
    """Verifica rol admin/superadmin para administrar usuarios del tenant."""
    if not principal.is_tenant_admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Solo admin o superadmin pueden administrar usuarios",
        )