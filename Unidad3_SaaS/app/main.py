"""FastAPI app entrypoint.

Bootstrap de la aplicacion:
- Crea la instancia `app`.
- Configura CORS con `AllowCredentials=True` y origenes por env
  (mismo patron que `Backend/main.go`).
- Monta los routers (`app/api/health.py` por ahora).
- `lifespan` queda como hook para DB engine startup/shutdown (T1.2).

Para correr en dev:
    uv run uvicorn app.main:app --reload --port 8000

Para produccion:
    uv run uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4
"""
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api.health import router as health_router
from app.core.config import settings


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Hook de startup/shutdown. El engine async se instancia lazy en T1.2."""
    yield


app = FastAPI(
    title=settings.api_title,
    version="0.1.0",
    description="Capa SaaS / multitenant del MultiCatalogo.",
    lifespan=lifespan,
)

# CORS: allow_credentials=True es obligatorio para que la cookie httpOnly
# del nucleo Go/Fiber (si en U4 la compartimos entre origenes) o cualquier
# cookie que use la propia UI SaaS viaje. Origenes por env.
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins_list,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Routers
app.include_router(health_router)


@app.get("/")
async def root() -> dict[str, str]:
    """Documentacion interactiva en `/docs` (Swagger UI)."""
    return {
        "service": "multicatalogo-saas",
        "docs": "/docs",
        "openapi": "/openapi.json",
    }
