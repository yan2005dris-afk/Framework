# gateway

Punto de entrada unico del MultiCatalogo. Caddy 2.x con:

- **Forward_auth** al `auth-service:3001` para validacion central de JWT.
- **Reverse proxy** a `backend:3000` (B2C) y `saas-api:8000` (SaaS).
- **File server** placeholder para el build del frontend (U4).

## Endpoints publicos via :8080

| Path | Upstream |
|---|---|
| `/api/tx/*` | backend Go/Fiber (B2C: catalogo, checkout, MLM) |
| `/api/saas/*` | saas-api FastAPI (multi-tenant) |
| `/` | placeholder (file_server; build real en U4) |

## Auth central

Todo request a `/api/tx/*` y `/api/saas/*` pasa primero por `forward_auth auth-service:3001` con `uri /auth/validate`. El auth-service devuelve 200 + headers `X-User-*` si el token (cookie `auth_token` o `Authorization: Bearer`) es valido; 401 si no.

Caddy copia esos headers al upstream con `copy_headers`. Los backends leen `X-User-Id`, `X-User-Email`, `X-User-Rol`, `X-Tenant-Id`, `X-Rol-Saas` para tomar identidad y tenant scope.

## Setup local con docker-compose

(GA4b — el compose raiz levanta todo junto.) Por ahora, para probar:

```bash
cd auth-service && go run main.go &   # auth-service en :3001
cd Backend && go run main.go &       # backend en :3000
docker compose up caddy              # gateway solo
```

Caddy necesita resolucion DNS de los nombres de servicio (`backend`, `auth-service`). Cuando corra via docker-compose, esto viene por la red Docker. Para test fuera de compose, hay que mapear los hosts en `/etc/hosts` o usar `--network host`.

## Verificacion

```bash
curl -i localhost:8080/api/tx/productos     # via gateway -> backend
curl -i -X POST localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@upse.edu.ec","password":"123456"}'
```

TLS llega en U4 (`auto_https` con cert de Let's Encrypt). Por ahora HTTP puro para dev.
