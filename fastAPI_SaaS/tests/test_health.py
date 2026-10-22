"""Smoke tests: los endpoints de salud responden 200 con el JSON esperado."""
import pytest


@pytest.mark.asyncio
async def test_health_returns_ok(client):
    resp = await client.get("/health")
    assert resp.status_code == 200
    body = resp.json()
    assert body["status"] == "ok"
    assert body["service"] == "saas"
    assert "version" in body


@pytest.mark.asyncio
async def test_saas_health_returns_ok(client):
    """El alias con path publico (lo que el proxy de U4 expone) responde igual."""
    resp = await client.get("/api/saas/health")
    assert resp.status_code == 200
    body = resp.json()
    assert body["status"] == "ok"
    assert body["service"] == "saas"


@pytest.mark.asyncio
async def test_root_points_to_docs(client):
    resp = await client.get("/")
    assert resp.status_code == 200
    body = resp.json()
    assert body["docs"] == "/docs"
    assert body["service"] == "multicatalogo-saas"
