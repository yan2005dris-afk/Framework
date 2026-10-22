"""Tests del proxy /api/saas/auth -> auth-service.

Tecnica: httpx.MockTransport reemplaza el transporte del cliente httpx
interno que usa app.api.auth. Asi no hay conexiones reales al upstream.

Convencion: cada test construye su propio MockTransport con respuestas
programables y se inyecta como el cliente httpx que usa el proxy.
"""
from __future__ import annotations

import json
from typing import Any

import httpx
import pytest
from httpx import Response

from app.api import auth as auth_module
from app.core.config import settings


def _build_mock_client(handler) -> httpx.AsyncClient:
    """Construye un AsyncClient con un MockTransport que delega cada
    request al handler dado. El handler recibe la httpx.Request y debe
    devolver una httpx.Response."""
    transport = httpx.MockTransport(handler)
    return httpx.AsyncClient(transport=transport, timeout=5.0)


@pytest.fixture
def patched_client(monkeypatch: pytest.MonkeyPatch):
    """Fixture: instala un cliente mockeado como el singleton del modulo.
    Devuelve el AsyncClient para que el test pueda inspeccionarlo si quiere.
    Cleanup automatico al final del test: resetea el singleton.
    """

    def install(handler):
        client = _build_mock_client(handler)
        monkeypatch.setattr(auth_module, "_client", client)
        return client

    yield install


# ----------------------------------------------------------------------
# /api/saas/auth/login
# ----------------------------------------------------------------------

@pytest.mark.asyncio
async def test_login_proxies_y_devuelve_200_con_cookie(patched_client, client):
    """Login con credenciales validas -> el upstream devuelve 200 + Set-Cookie;
    el SaaS los forwardea integros al cliente."""

    def handler(req: httpx.Request) -> Response:
        # Verifica que la request llega al path correcto y con el body.
        assert req.method == "POST"
        assert req.url.path == "/auth/login"
        body = json.loads(req.content)
        assert body == {"email": "admin@upse.edu.ec", "password": "123456"}
        return Response(
            status_code=200,
            json={"email": "admin@upse.edu.ec", "rol": "admin", "expires": 7200},
            headers={"set-cookie": "auth_token=fake-jwt; Path=/; HttpOnly; SameSite=Lax"},
        )

    patched_client(handler)

    resp = await client.post(
        "/api/saas/auth/login",
        json={"email": "admin@upse.edu.ec", "password": "123456"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["email"] == "admin@upse.edu.ec"
    assert body["rol"] == "admin"
    # La cookie llego al cliente.
    cookies = resp.headers.get_list("set-cookie")
    assert any("auth_token=fake-jwt" in c for c in cookies)


@pytest.mark.asyncio
async def test_login_401_upstream_no_propag_cookie(patched_client, client):
    """Si el upstream devuelve 401, el SaaS NO debe inventar una cookie de sesion."""

    def handler(req: httpx.Request) -> Response:
        return Response(
            status_code=401,
            json={"error": "Credenciales incorrectas"},
            # Sin Set-Cookie intencionalmente.
        )

    patched_client(handler)

    resp = await client.post(
        "/api/saas/auth/login",
        json={"email": "admin@upse.edu.ec", "password": "wrong"},
    )
    assert resp.status_code == 401
    # El body se forwardea (mensaje del upstream).
    assert resp.json()["error"] == "Credenciales incorrectas"
    # Ninguna cookie debe setearse.
    cookies = resp.headers.get_list("set-cookie")
    assert len(cookies) == 0


@pytest.mark.asyncio
async def test_login_502_si_upstream_no_responde(patched_client, client):
    """Si el cliente httpx no puede conectar (RequestError), responder 502."""

    def handler(req: httpx.Request) -> Response:
        # raise dentro de MockTransport dispara RequestError en el caller.
        raise httpx.ConnectError("connection refused: upstream not reachable")

    patched_client(handler)

    resp = await client.post(
        "/api/saas/auth/login",
        json={"email": "x@y.z", "password": "x"},
    )
    assert resp.status_code == 502
    assert "auth-service no responde" in resp.json()["error"]


# ----------------------------------------------------------------------
# /api/saas/auth/logout
# ----------------------------------------------------------------------

@pytest.mark.asyncio
async def test_logout_proxied_y_cookie_limpiada(patched_client, client):
    """Logout: el upstream setea la cookie con Max-Age=-1 (borrar); el SaaS
    la propaga al cliente."""

    def handler(req: httpx.Request) -> Response:
        return Response(
            status_code=200,
            json={"status": "ok"},
            headers={"set-cookie": "auth_token=; Path=/; Max-Age=0"},
        )

    patched_client(handler)

    resp = await client.post("/api/saas/auth/logout")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}
    cookies = resp.headers.get_list("set-cookie")
    assert any("Max-Age=0" in c for c in cookies)


# ----------------------------------------------------------------------
# /api/saas/auth/me
# ----------------------------------------------------------------------

@pytest.mark.asyncio
async def test_me_proxied_y_devuelve_identidad(patched_client, client):
    """GET /api/saas/auth/me -> upstream GET /auth/me -> 200 con identidad."""

    def handler(req: httpx.Request) -> Response:
        assert req.method == "GET"
        assert req.url.path == "/auth/me"
        return Response(
            status_code=200,
            json={"id": 1, "email": "admin@upse.edu.ec", "rol": "admin"},
        )

    patched_client(handler)

    resp = await client.get("/api/saas/auth/me")
    assert resp.status_code == 200
    body = resp.json()
    assert body["id"] == 1
    assert body["email"] == "admin@upse.edu.ec"
    assert body["rol"] == "admin"


# ----------------------------------------------------------------------
# Forwarding de headers sensibles
# ----------------------------------------------------------------------

@pytest.mark.asyncio
async def test_login_forwarda_authorization_y_cookie(patched_client, client):
    """El cliente puede pasar Authorization y Cookie al endpoint; el SaaS los
    propaga al auth-service."""

    captured: dict[str, Any] = {}

    def handler(req: httpx.Request) -> Response:
        captured["authorization"] = req.headers.get("authorization")
        captured["cookie"] = req.headers.get("cookie")
        return Response(
            status_code=200,
            json={"email": "admin@upse.edu.ec"},
        )

    patched_client(handler)

    resp = await client.post(
        "/api/saas/auth/login",
        json={"email": "admin@upse.edu.ec", "password": "123456"},
        headers={
            "Authorization": "Bearer legacy-jwt",
            "Cookie": "auth_token=legacy-jwt; other=val",
        },
    )
    assert resp.status_code == 200
    assert captured["authorization"] == "Bearer legacy-jwt"
    assert captured["cookie"] == "auth_token=legacy-jwt; other=val"


# ----------------------------------------------------------------------
# URL upstream correcta
# ----------------------------------------------------------------------

@pytest.mark.asyncio
async def test_url_upstream_apunta_a_auth_service(patched_client, client, monkeypatch: pytest.MonkeyPatch):
    """El path del upstream es auth_service_url + stripped_path()."""
    monkeypatch.setattr(settings, "auth_service_url", "http://auth.internal:9999")

    captured: dict[str, Any] = {}

    def handler(req: httpx.Request) -> Response:
        captured["url"] = str(req.url)
        return Response(status_code=204)

    patched_client(handler)

    resp = await client.post("/api/saas/auth/login", json={"email": "x", "password": "y"})
    assert resp.status_code == 204
    # Trailing slash no debe romper.
    assert captured["url"] == "http://auth.internal:9999/auth/login"
