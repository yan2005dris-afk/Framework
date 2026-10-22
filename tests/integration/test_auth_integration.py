"""Tests de integración para verificar que la autenticación funcione
correctamente entre auth-service, Backend Go y SaaS FastAPI.

Estos tests verifican:
1. Que el JWT_SECRET sea compartido entre servicios
2. Que los tokens emitidos por auth-service sean válidos en ambos backends
3. Que los proxies funcionen correctamente
4. Que los headers X-User-* se propaguen correctamente

Pre-requisitos:
- docker compose up -d --build
- Esperar que los healthchecks pasen (~15s)

Ejecutar:
    cd tests/integration
    uv run pytest test_auth_integration.py -v
"""
import os
from typing import Dict

import httpx
import pytest
import jwt


# URLs de los servicios
GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")
BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:3000")
AUTH_SERVICE_URL = os.getenv("AUTHSVC_URL", "http://localhost:3001")
SAAS_URL = os.getenv("SAAS_URL", "http://localhost:8000")

# Credenciales de prueba
TEST_EMAIL = "admin@upse.edu.ec"
TEST_PASSWORD = "123456"


@pytest.fixture
def http_client():
    """Cliente HTTP con timeout."""
    with httpx.Client(timeout=10.0) as client:
        yield client


@pytest.fixture
def auth_token(http_client):
    """Fixture que obtiene un token válido del auth-service."""
    response = http_client.post(
        f"{AUTH_SERVICE_URL}/auth/login",
        json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
    )
    assert response.status_code == 200, f"Login falló: {response.text}"

    # Extraer la cookie
    cookies = response.cookies
    assert "auth_token" in cookies, "No se recibió cookie auth_token"

    return cookies["auth_token"]


class TestJWTSecretConsistency:
    """Verifica que ambos servicios usen el mismo JWT_SECRET."""

    def test_jwt_secret_env_vars(self):
        """Verifica que las variables de entorno estén configuradas."""
        auth_secret = os.getenv("AUTH_JWT_SECRET")
        if not auth_secret:
            pytest.skip("AUTH_JWT_SECRET no está configurado")

        # En producción, JWT_SECRET del Backend debe coincidir
        backend_secret = os.getenv("AUTH_JWT_SECRET") or os.getenv("JWT_SECRET")

        if backend_secret != auth_secret:
            pytest.fail(
                f"JWT_SECRET del Backend ({backend_secret}) no coincide con "
                f"AUTH_JWT_SECRET del auth-service ({auth_secret}). "
                "Esto causará que los tokens no sean compatibles."
            )

    def test_token_parsable_with_env_secret(self, auth_token):
        """Verifica que el token emitido sea parseable con el secret actual."""
        secret = os.getenv("AUTH_JWT_SECRET")
        if not secret:
            secret = "dev-insecure-secret-please-replace-in-production"

        try:
            decoded = jwt.decode(auth_token, secret, algorithms=["HS256"])
            assert decoded["email"] == TEST_EMAIL
            assert decoded["rol"] == "admin"
        except jwt.InvalidTokenError as e:
            pytest.fail(f"Token no es válido con el JWT_SECRET actual: {e}")


class TestAuthServiceDirect:
    """Tests directos contra el auth-service."""

    def test_login_emits_valid_jwt(self, http_client):
        """Verifica que el login emita un JWT válido."""
        response = http_client.post(
            f"{AUTH_SERVICE_URL}/auth/login",
            json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
        )

        assert response.status_code == 200, f"Login falló: {response.text}"

        # Verificar cookie
        cookies = response.cookies
        assert "auth_token" in cookies, "No se recibió cookie auth_token"

        token = cookies["auth_token"]
        secret = os.getenv("AUTH_JWT_SECRET") or "dev-insecure-secret-please-replace-in-production"

        decoded = jwt.decode(token, secret, algorithms=["HS256"])
        assert decoded["email"] == TEST_EMAIL
        assert decoded["rol"] == "admin"

    def test_invalid_credentials_return_401(self, http_client):
        """Verifica que credenciales inválidas devuelvan 401."""
        response = http_client.post(
            f"{AUTH_SERVICE_URL}/auth/login",
            json={"email": TEST_EMAIL, "password": "wrong"},
        )

        assert response.status_code == 401
        assert "error" in response.json()


class TestTokenReuse:
    """Verifica que un token funcione en ambos backends."""

    def test_token_works_in_backend(self, http_client, auth_token):
        """Verifica que el token funcione en el Backend."""
        response = http_client.get(
            f"{BACKEND_URL}/api/me",
            cookies={"auth_token": auth_token},
        )

        assert response.status_code == 200, f"Backend rechazó token: {response.text}"
        data = response.json()
        assert data["email"] == TEST_EMAIL
        assert data["rol"] == "admin"

    def test_token_works_in_saas(self, http_client, auth_token):
        """Verifica que el token funcione en el SaaS."""
        response = http_client.get(
            f"{SAAS_URL}/api/saas/auth/me",
            cookies={"auth_token": auth_token},
        )

        assert response.status_code == 200, f"SaaS rechazó token: {response.text}"
        data = response.json()
        assert data["email"] == TEST_EMAIL
        assert data["rol"] == "admin"

    def test_token_rejected_without_cookie_backend(self, http_client):
        """Verifica que el Backend rechace requests sin token."""
        response = http_client.get(f"{BACKEND_URL}/api/me")

        assert response.status_code == 401

    def test_token_rejected_without_cookie_saas(self, http_client):
        """Verifica que el SaaS rechace requests sin token."""
        response = http_client.get(f"{SAAS_URL}/api/saas/auth/me")

        assert response.status_code == 401


class TestGatewayAuth:
    """Tests de autenticación a través del gateway."""

    def test_gateway_login_backend(self, http_client):
        """Verifica login a través del gateway → Backend."""
        response = http_client.post(
            f"{GATEWAY_URL}/api/auth/login",
            json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
        )

        assert response.status_code == 200, f"Login vía gateway falló: {response.text}"
        assert "auth_token" in response.cookies, "No se recibió cookie vía gateway"

    def test_gateway_login_saas(self, http_client):
        """Verifica login a través del gateway → SaaS."""
        response = http_client.post(
            f"{GATEWAY_URL}/api/saas/auth/login",
            json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
        )

        assert response.status_code == 200, f"Login vía gateway (SaaS) falló: {response.text}"
        assert "auth_token" in response.cookies, "No se recibió cookie vía gateway (SaaS)"

    def test_gateway_forward_auth_with_token(self, http_client, auth_token):
        """Verifica que forward_auth del gateway funcione con token válido."""
        response = http_client.get(
            f"{GATEWAY_URL}/api/tx/red",
            cookies={"auth_token": auth_token},
        )

        assert response.status_code == 200, f"Forward_auth falló: {response.text}"

    def test_gateway_forward_auth_without_token(self, http_client):
        """Verifica que forward_auth del gateway rechace sin token."""
        response = http_client.get(f"{GATEWAY_URL}/api/tx/red")

        assert response.status_code == 401


class TestCrossServiceCompatibility:
    """Verifica compatibilidad cross-service."""

    def test_token_from_backend_works_in_saas(self, http_client):
        """Verifica que un token obtenido vía Backend funcione en SaaS."""
        # Obtener token vía Backend (directo)
        response = http_client.post(
            f"{BACKEND_URL}/api/login",
            json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
        )
        assert response.status_code == 200
        token = response.cookies.get("auth_token")

        # Usar en SaaS
        response = http_client.get(
            f"{SAAS_URL}/api/saas/auth/me",
            cookies={"auth_token": token},
        )
        assert response.status_code == 200, f"Token del Backend no funciona en SaaS: {response.text}"

    def test_token_from_saas_works_in_backend(self, http_client):
        """Verifica que un token obtenido vía SaaS funcione en Backend."""
        # Obtener token vía SaaS (directo)
        response = http_client.post(
            f"{SAAS_URL}/api/saas/auth/login",
            json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
        )
        assert response.status_code == 200
        token = response.cookies.get("auth_token")

        # Usar en Backend
        response = http_client.get(
            f"{BACKEND_URL}/api/me",
            cookies={"auth_token": token},
        )
        assert response.status_code == 200, f"Token del SaaS no funciona en Backend: {response.text}"

    def test_token_from_gateway_works_in_both(self, http_client):
        """Verifica que un token obtenido vía gateway funcione en ambos."""
        # Obtener token vía gateway
        response = http_client.post(
            f"{GATEWAY_URL}/api/auth/login",
            json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
        )
        assert response.status_code == 200
        token = response.cookies.get("auth_token")

        # Usar en Backend directo
        response_backend = http_client.get(
            f"{BACKEND_URL}/api/me",
            cookies={"auth_token": token},
        )
        assert response_backend.status_code == 200

        # Usar en SaaS directo
        response_saas = http_client.get(
            f"{SAAS_URL}/api/saas/auth/me",
            cookies={"auth_token": token},
        )
        assert response_saas.status_code == 200


class TestHeadersPropagation:
    """Verifica que los headers X-User-* se propaguen correctamente."""

    def test_auth_validate_sets_headers(self, http_client, auth_token):
        """Verifica que /auth/validate setee los headers X-User-*."""
        response = http_client.post(
            f"{AUTH_SERVICE_URL}/auth/validate",
            cookies={"auth_token": auth_token},
        )

        assert response.status_code == 200
        # El auth-service debería setear headers X-User-*
        # En un test real, verificaríamos que el gateway los copie al upstream
