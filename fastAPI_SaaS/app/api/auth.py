"""Proxy HTTP al auth-service para los endpoints de auth.

Post-GA6 el SaaS ya no firma JWTs ni verifica passwords directamente:
todas esas responsabilidades viven en auth-service:3001. Estos handlers
son un proxy HTTP transparente que reenvian la request al upstream y
devuelven su respuesta (status + headers + body) al cliente.

Convenciones:
- Path en SaaS: /api/saas/auth/login, /api/saas/auth/logout,
  /api/saas/auth/refresh, /api/saas/auth/me.
- Path upstream (auth-service): /auth/login, /auth/logout, /auth/refresh,
  /auth/me. Sin prefijo /api/saas porque el gateway ya hace el enrutado.

Esta indireccion existe para mantener compatibilidad con cualquier cliente
que pueda llegar al SaaS saltandose el gateway (defense in depth) y para
no requerir URLs distintas en el Frontend del SaaS.
"""
from __future__ import annotations

import logging
from typing import Optional

import httpx
from fastapi import APIRouter, Request
from fastapi.responses import JSONResponse, Response

from app.core.config import settings

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/saas/auth", tags=["auth"])


# Cliente httpx singleton. Lazy: se crea en el primer request.
# Los tests pueden sobrescribir el simbolo `app.api.auth._client` con
# monkeypatch para inyectar un cliente con transporte mockeado.
_client: Optional[httpx.AsyncClient] = None


def _get_client() -> httpx.AsyncClient:
    global _client
    if _client is None:
        _client = httpx.AsyncClient(timeout=5.0)
    return _client


async def _close_client() -> None:
    """Cierra el cliente httpx. Llamado desde el lifespan shutdown."""
    global _client
    if _client is not None:
        await _client.aclose()
        _client = None


def _stripped_path(request: Request) -> str:
    """Devuelve el path despues del prefijo del router.

    Ej: si la request llega a /api/saas/auth/login y el router tiene
    prefix='/api/saas/auth', esto devuelve '/login'.
    """
    return request.url.path.removeprefix(str(router.prefix))


def _forward_headers(request: Request) -> dict[str, str]:
    """Headers a propagar del cliente al auth-service."""
    fwd: dict[str, str] = {}
    for header in ("authorization", "cookie", "content-type"):
        v = request.headers.get(header)
        if v:
            fwd[header] = v
    return fwd


async def _proxy(request: Request, method: str, body: bytes | None) -> Response:
    """Reenvia una request al auth-service y devuelve su respuesta.

    - Construye la URL upstream: {auth_service_url}/auth{stripped_path}.
      Esto transforma el path /api/saas/auth/login (SaaS) en /auth/login (upstream).
    - Header-by-header forward.
    - Body crudo (JSON o vacio).
    - Devuelve status + headers + body integros.
    """
    upstream = f"{settings.auth_service_url.rstrip('/')}/auth{_stripped_path(request)}"
    client = _get_client()

    # Limpiar el cookie jar antes de cada request. httpx.AsyncClient trae un
    # jar propio: al proxear un login guardaba el `auth_token` de ese usuario
    # y lo reenviaba en TODAS las requests siguientes, asi que
    # /api/saas/auth/me respondia 200 con la identidad de ese usuario aunque
    # el cliente no mandara token (sesion filtrada entre usuarios).
    # `cookies=None` al construir el cliente NO alcanza: httpx igual crea un
    # jar vacio que se vuelve a llenar. Limpiarlo aca deja la unica cookie
    # que viaja como la que trae el request, que `_forward_headers` ya
    # propaga explicitamente.
    client.cookies.clear()

    try:
        upstream_resp = await client.request(
            method=method,
            url=upstream,
            headers=_forward_headers(request),
            content=body,
        )
    except httpx.RequestError as exc:
        logger.error("auth-service unreachable: %s", exc)
        return JSONResponse(
            status_code=502,
            content={"error": "auth-service no responde", "detalle": str(exc)},
        )

    # Copia headers de respuesta, en particular Set-Cookie (cookie httpOnly).
    # Saltamos hop-by-hop headers que el cliente no deberia recibir proxied.
    passthrough: dict[str, str] = {}
    skip = {"content-length", "transfer-encoding", "connection"}
    for k, v in upstream_resp.headers.items():
        if k.lower() in skip:
            continue
        passthrough[k] = v

    return Response(
        status_code=upstream_resp.status_code,
        headers=passthrough,
        content=upstream_resp.content,
    )


async def _read_body(request: Request) -> bytes:
    """Lee el body crudo. Para GET/HEAD no hay body."""
    if request.method in {"GET", "HEAD"}:
        return b""
    return await request.body()


# ----------------------------------------------------------------------
# Handlers: uno por endpoint. Logica comun delegada a _proxy.
# ----------------------------------------------------------------------

@router.api_route(
    "/login",
    methods=["POST"],
    summary="Proxy de POST /auth/login al auth-service.",
)
async def login(request: Request) -> Response:
    body = await _read_body(request)
    return await _proxy(request, "POST", body)


@router.api_route(
    "/logout",
    methods=["POST"],
    summary="Proxy de POST /auth/logout al auth-service.",
)
async def logout(request: Request) -> Response:
    return await _proxy(request, "POST", None)


@router.api_route(
    "/refresh",
    methods=["POST"],
    summary="Proxy de POST /auth/refresh al auth-service.",
)
async def refresh(request: Request) -> Response:
    return await _proxy(request, "POST", None)


@router.api_route(
    "/me",
    methods=["GET"],
    summary="Proxy de GET /auth/me al auth-service (identidad del token).",
)
async def me(request: Request) -> Response:
    return await _proxy(request, "GET", None)
