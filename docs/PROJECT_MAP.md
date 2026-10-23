# PROJECT MAP — Framework Programación Web 2026

> Mapa arquitectónico del workspace. Leer ESTE archivo antes de explorar código.
> Proyecto didáctico "MultiCatálogo": tienda de cosméticos con red de ventas
> multinivel (MLM) + capa SaaS multi-tenant. Asignatura: Framework Programación Web (UPSE, 2026-02).

## Estado del stack (post-U3-T2 octubre 2026)

```
┌────────────────────────────────────────────────────────────────────────────┐
│  SPA React (Frontend/) — build estatico compilado dentro de la imagen      │
│  del gateway. cookies: auth_token + saas_token                             │
└────────────────────────────┬───────────────────────────────────────────────┘
                             │ HTTP, mismo origen :8080
                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│  gateway/ (Caddy 2)              :8080                                     │
│  file_server / (SPA) + reverse_proxy + forward_auth → auth-service:3001    │
└────┬─────────────────┬─────────────────┬─────────────────┬────────────────┘
     │                 │                 │                 │
     ▼                 ▼                 ▼                 ▼
┌──────────┐    ┌──────────────┐  ┌──────────────┐    ┌──────────┐
│ Backend/ │    │ auth-service/ │  │ fastAPI_SaaS/ │    │ redis    │
│ Go/Fiber │    │ Go/Fiber      │  │ FastAPI      │    │ :6379    │
│ :3000    │    │ :3001         │  │ :8000        │    └──────────┘
└────┬─────┘    └──────┬───────┘  └──────┬───────┘
     │                │                │
     └────────────────┴────────────────┘
                      │
                      ▼
              Postgres 16 :5432
              schemas: auth, backend (public), saas
```

### Frontend: produccion vs. desarrollo

- **Produccion (defecto):** el `dist/` del SPA se compila en el stage `frontend` de `gateway/Dockerfile` y el gateway lo sirve en `/` con fallback a `index.html`. SPA y API comparten origen en **:8080**, que es lo que hace funcionar la cookie httpOnly `auth_token` sin CORS. `docker compose up -d` ya deja la app completa; no hay frontend que levantar.
- **Desarrollo (opcional):** el servicio `frontend` del compose, bajo el profile `dev`, corre el dev server de Vite con HMR sobre el codigo del host en **:5173** y proxea `/api` al gateway (`VITE_DEV_API_TARGET=http://gateway:8080`). Es una comodidad, no parte del stack: sin `--profile dev` no arranca.

```bash
docker compose --profile dev up -d frontend    # http://localhost:5173
```

Ver `Frontend/README.md` y `gateway/README.md`.

## Servicios y sus READMEs

| Servicio | Stack | Puerto | README |
|---|---|---|---|
| `gateway/` | Caddy 2 (SPA + API) | 8080 | `gateway/README.md` |
| `Backend/` | Go/Fiber + pgx + redis | 3000 | `Backend/README.md` |
| `auth-service/` | Go/Fiber + JWT + bcrypt | 3001 | `auth-service/README.md` |
| `fastAPI_SaaS/` | FastAPI + SQLAlchemy 2 + Alembic | 8000 | `fastAPI_SaaS/README.md` |
| `Frontend/` | React 19 + Vite 8 + Tailwind 4 | 8080 (prod, via gateway) / 5173 (dev) | `Frontend/README.md` |

## Arquitectura detallada

Ver [`docs/architecture.md`](architecture.md) para:
- Diagrama completo del request flow.
- Convenciones (BD única con schemas, JWT en cookies, auth centralizada).
- Cómo se levanta todo con `docker compose up`.

## Documentos raíz (meta, no código)

| Archivo | Contenido |
|---|---|
| `README.md` | Quickstart del stack completo (gateway + servicios). |
| `docs/architecture.md` | Living doc de la arquitectura y el request flow. |
| `docs/PROJECT_MAP.md` | Este archivo. Índice de navegación. |
| `docs/REQUIREMENTS_CHECKLIST.md` | Estado del proyecto por fases (U1 → U2 → U3 → U4) + roadmap. |
| `odd/tasks/u*.md` | Feature docs ODD con el plan y log de implementación por sub-feature. |
| `odd/tasks/u1-t*.md` ... | Históricos. |

## Issues de GitHub abiertas (no olvidar)

| # | Qué |
|---|---|
| #15 | U3-T3: CRUD de `users_saas` (invitar usuarios a tenants). |
| #16 | U3-T3: planes + billing (Stripe webhook stub). |
| #17 | U3-T3: middleware aislamiento multi-tenant. |
| #18 | U3-T4+: tests E2E Playwright. |
| #19 | U4: SPA build + Caddy sirviendo `/`. |
| #20 | U4: TLS con Let's Encrypt (`auto_https`). |
| #21 | U4: rate_limit + observabilidad + deploy pipeline. |

## Comandos frecuentes

```bash
# Levantar todo (SPA incluido: lo sirve el gateway en :8080)
docker compose up -d --build

# Frontend con HMR en :5173, ademas del stack
docker compose --profile dev up -d frontend

# Ver logs
docker compose logs -f gateway auth-service

# Tests E2E bash/curl
bash tests/e2e/gateway_e2e.sh

# Backend — correr tests
cd Backend && go test ./...

# SaaS — correr tests
cd fastAPI_SaaS && uv run pytest

# auth-service — correr tests
cd auth-service && go test ./...
```
