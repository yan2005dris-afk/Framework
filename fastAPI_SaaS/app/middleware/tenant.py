"""Middleware de contexto multi-tenant.

Responsabilidad unica: resolver la identidad del request UNA vez y dejarla
en `request.state.principal`, de modo que los endpoints no relean headers.

Que NO hace (a proposito):
- No autoriza nada. Autorizar es de los endpoints, via
  `app.deps.principal.require_tenant_scope` / `require_tenant_admin`.
- No devuelve 401. `/api/saas/health` y los endpoints publicos (login)
  corren sin identidad; el 401 lo emite `get_principal` cuando un endpoint
  que si la exige no la recibe.
- No revalida el JWT: lo hace el gateway via `forward_auth` al auth-service.

Aislamiento a nivel de fila (RLS) vive en la dependencia de sesion
`app/core/db.py`, que necesita la conexion, no el request.
"""
from __future__ import annotations

from collections.abc import Awaitable, Callable

from fastapi import Request, Response

from app.deps.principal import principal_from_headers


async def set_tenant_context(
    request: Request,
    call_next: Callable[[Request], Awaitable[Response]],
) -> Response:
    """Deja `request.state.principal` con la identidad del gateway (o None)."""
    request.state.principal = principal_from_headers(request)
    return await call_next(request)