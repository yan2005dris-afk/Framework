"""Endpoints de salud.

`GET /health` y `GET /api/saas/health` son publicos (sin auth) y siempre
responden rapido. La diferencia es que `/api/saas/health` es el path
que el proxy reverso de U4 (Caddy/Nginx) expone y por donde llega el
SPA al SaaS; `/health` es para checks internos / Docker healthcheck.
"""
from fastapi import APIRouter

router = APIRouter(tags=["health"])


@router.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok", "service": "saas", "version": "0.1.0"}


@router.get("/api/saas/health")
async def saas_health() -> dict[str, str]:
    """Alias con el path publico del proxy. Misma respuesta que `/health`."""
    return {"status": "ok", "service": "saas", "version": "0.1.0"}
