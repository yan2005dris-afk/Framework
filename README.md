# MultiCatálogo

Sistema políglota de tienda cosmética + red MLM, organizado en microservicios.

## Componentes

| Servicio | Puerto | Stack | Rol |
|---|---|---|---|
| `gateway` | 8080 | Caddy 2 | Punto de entrada unico. Routing `/api/tx/*` y `/api/saas/*`. Auth central via `forward_auth` al auth-service. |
| `auth-service` | 3001 | Go/Fiber | Emision y validacion de JWTs (HS256). Schema `auth.users` en la BD compartida. |
| `backend` | 3000 | Go/Fiber | B2C: catalogo de productos, checkout transaccional ACID, red multinivel. Schema `backend.*` (legacy). |
| `saas-api` | 8000 | FastAPI | SaaS multi-tenant (tenants, planes, billing). Schema `saas.*` (en construccion). |
| `db` | 5432 | Postgres 16 | BD unica compartida. Cada servicio escribe solo en su schema. |
| `redis` | 6379 | Redis 7 | Cache-Aside del Backend (catalogo + red MLM). |

## Quickstart

```bash
# 1. Variables de entorno
cp .env.example .env

# 2. Levantar todo el stack
docker compose up -d --build

# 3. Aplicar la migracion del schema 'auth' (si no se aplico en el entrypoint)
bash auth-service/scripts/apply.sh

# 4. Login de prueba contra el gateway
curl -i -c /tmp/cookies.txt -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@upse.edu.ec","password":"123456"}'

# 5. Validar el token (forward_auth)
curl -i -b /tmp/cookies.txt -X POST http://localhost:8080/auth/validate
```

## Documentacion por area

- **Arquitectura completa (diagramas + request flow)**: `docs/architecture.md`.
- **Estado del proyecto (features y pendientes)**: `docs/REQUIREMENTS_CHECKLIST.md`.
- **Mapa del workspace**: `docs/PROJECT_MAP.md`.
- **Backend B2C**: `Backend/README.md`.
- **SaaS multi-tenant**: `fastAPI_SaaS/README.md`.
- **auth-service**: `auth-service/README.md`.
- **gateway**: `gateway/README.md`.
- **Logs de implementacion / decisiones**: `odd/tasks/`.
- **Tests de integración de autenticación**: `tests/integration/README.md`.

## Estado del proyecto

V1: U1 (Frontend SPA) y U2 (Backend B2C) completas.
V2: U3 (SaaS multi-tenant) en construccion. U4 (deploy con TLS) pendiente.

Issues abiertas en GitHub para no olvidar fases siguientes:

- **#15** U3-T3: CRUD de `users_saas`.
- **#16** U3-T3: planes y billing.
- **#17** U3-T3: middleware de aislamiento multi-tenant.
- **#18** U3-T4+: tests E2E con Playwright.
- **#19** U4: build del frontend + Caddy sirviendo `/`.
- **#20** U4: TLS con Let's Encrypt.
- **#21** U4: rate_limit + observabilidad + deploy pipeline.
