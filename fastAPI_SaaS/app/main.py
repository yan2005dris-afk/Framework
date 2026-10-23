"""FastAPI app entrypoint.

Bootstrap de la aplicacion:
- Crea la instancia `app`.
- Configura CORS con `AllowCredentials=True` y origenes por env
  (mismo patron que `Backend/main.go`).
- Monta routers: `app/api/health.py`, `app/api/auth.py` (proxy al
  auth-service), `app/api/users.py` (CRUD multi-tenant) y
  `app/api/billing.py` (planes y suscripciones).
- `lifespan` queda como hook para DB engine startup/shutdown (T1.2).

Para correr en dev:
    uv run uvicorn app.main:app --reload --port 8000

Para produccion:
    uv run uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4
"""
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api.auth import _close_client, router as auth_router
from app.api.billing import router as billing_router
from app.api.health import router as health_router
from app.api.users import router as users_router
from app.core.config import settings
from app.core.logging import configure_logging
from app.middleware.tenant import set_tenant_context

# Se llama a nivel de modulo: uvicorn ya configuro sus loggers cuando importa
# la app, asi que esta es la ultima palabra sobre el formato de salida.
configure_logging()


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Hook de startup/shutdown. Cierra el cliente httpx singleton del proxy."""
    yield
    await _close_client()


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

# Middleware de aislamiento multi-tenant
app.middleware("http")(set_tenant_context)

# Routers
app.include_router(health_router)
app.include_router(auth_router)
app.include_router(users_router)
app.include_router(billing_router)


@app.get("/")
async def root() -> dict[str, str]:
    """Documentacion interactiva en `/docs` (Swagger UI)."""
    return {
        "service": "multicatalogo-saas",
        "docs": "/docs",
        "openapi": "/openapi.json",
    }
