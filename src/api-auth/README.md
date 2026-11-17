# auth-service

Microservicio de autenticación extraído del Backend Go/Fiber (U3-T2). Centraliza emisión y validación de JWTs.

## Rol

```
SPA -> gateway (Caddy, :8080)
       |
       +-> /api/tx/*    -> backend Go/Fiber (:3000)   [B2C: catalogo + checkout + MLM]
       +-> /api/saas/*  -> Unidad3 SaaS/FastAPI (:8000) [multi-tenant]
       |
       +-> forward_auth -> auth-service (:3001)        [validacion central de JWTs]
                          |
                          +-> auth-db Postgres (:5436)
```

El auth-service emite **dos tipos de tokens** con la misma `AUTH_JWT_SECRET` (compatible con el `JWT_SECRET` del Backend):

- **B2C** (`auth_token` cookie): claims `{sub, email, rol, iat, exp}`.
- **SaaS multi-tenant** (`saas_token` cookie): claims `{sub, email, tenant_id, rol_saas, iat, exp}`.

## Setup local

```bash
go mod download
cp .env.example .env
# ajustar AUTH_DB_PASSWORD, AUTH_JWT_SECRET segun entorno

go run main.go
# levanta en :3001
```

Abre `http://localhost:3001/auth/health` para el healthcheck (200 con `{status:"ok"}` si la BD responde; 503 si no).

## Endpoints

- `GET  /auth/health` — healthcheck ping a la BD.
- `POST /auth/login` — (GA2) body `{email, password}` -> tokens B2C y SaaS si corresponden.
- `POST /auth/logout` — (GA2) limpia cookies.
- `POST /auth/refresh` — (GA2) emite token nuevo a partir de uno proximo a expirar.
- `POST /auth/validate` — (GA2) usado por el gateway via `forward_auth`. Devuelve 200 con headers `X-User-*` que Caddy pasa al backend upstream.
- `GET  /auth/me` — (GA2) identidad del token.

## Variables de entorno

Ver `.env.example`.

- `AUTH_JWT_SECRET`: secret HS256. Si coincide con el `JWT_SECRET` del Backend, los tokens son compatibles; sino, cada servicio valida solo los suyos.
- `AUTH_DB_*`: conexion a su propia BD (`auth_db`) en `:5436`.
