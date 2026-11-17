# gateway

Punto de entrada unico del MultiCatalogo. Caddy 2.x con:

- **File server** del SPA compilado (`Frontend/dist` → `/srv/frontend`), con fallback a `index.html` para las rutas de cliente.
- **Forward_auth** al `auth-service:3001` para validacion central de JWT.
- **Reverse proxy** a `backend:3000` (B2C) y `saas-api:8000` (SaaS).

El frontend va dentro de esta misma imagen, asi que el unico puerto publico del stack es `:8080` y el SPA y la API comparten origen. Eso es lo que hace funcionar la cookie httpOnly `auth_token` sin CORS.

## Endpoints publicos via :8080

| Path | Upstream |
|---|---|
| `/health` | el propio gateway (sin auth) |
| `/api/*` | backend Go/Fiber (B2C: catalogo, checkout, MLM) |
| `/api/saas/*` | saas-api FastAPI (multi-tenant) |
| `/` y rutas de cliente (`/catalogo`, `/mi-red`, `/producto/:id`) | el `index.html` del SPA |

## Como entra el frontend en la imagen

`gateway/Dockerfile` es multistage:

1. stage `frontend`: `node:22` + pnpm corre `pnpm build` sobre `Frontend/` y produce `dist/`.
2. stage final `caddy:2`: copia ese `dist/` a `/srv/frontend` y sirve.

Dos detalles que conviene no romper al tocar esto:

- **El contexto de build es la raiz del repo**, no `./gateway`: el stage de build necesita `Frontend/` y `pnpm-lock.yaml`. Por eso `docker-compose.yml` usa `context: .` + `dockerfile: gateway/Dockerfile`, y por eso existe el `.dockerignore` de la raiz (sin el, el contexto son ~440MB en vez de ~1MB).
- **`src/routeTree.gen.ts` no viaja al build**: lo genera `tsr generate` dentro de `pnpm build` y esta en `.gitignore` + `.dockerignore`. Mandar el del host lo deja desactualizado y rompe el `tsc -b`.

## SPA: por que `try_files`

El SPA tiene rutas de cliente. Un refresh en `/catalogo` o un deep link a `/producto/123` no matchean ningun archivo en disco, asi que sin fallback darian 404 en vez de arrancar la app:

```
root * /srv/frontend
try_files {path} /index.html
file_server
```

Los assets con hash (`/assets/*.js`) si existen en disco, asi que `try_files {path}` los sirve directos y no pasan por el `index.html`.

El bloque va **dentro del `route`** a proposito: Caddy ordena las directivas por tipo fuera de `route`, y un `file_server` suelto correria antes del `forward_auth`.

## Auth central

Todo request a `/api/*` y `/api/saas/*` pasa primero por `forward_auth auth-service:3001` con `uri /auth/validate`, salvo las rutas publicas listadas en el Caddyfile (`/api/login`, `/api/logout`, `/api/health`, `/api/productos*`, `/api/saas/health`, `/api/saas/webhooks/stripe`, `/api/saas/auth/*`). El auth-service devuelve 200 + headers `X-User-*` si el token (cookie `auth_token` o `Authorization: Bearer`) es valido; 401 si no.

Caddy copia esos headers al upstream con `copy_headers`. Los backends leen `X-User-Id`, `X-User-Email`, `X-User-Rol`, `X-Tenant-Id`, `X-Rol-Saas` para tomar identidad y tenant scope.

## Desarrollo

El stack completo, con el SPA ya servido:

```bash
docker compose up -d --build    # SPA en http://localhost:8080
```

### Con HMR (profile `dev`)

Para iterar sobre el frontend sin rebuild de imagen, el compose tiene un servicio `frontend` que corre el dev server de Vite con el codigo del host montado:

```bash
docker compose --profile dev up -d frontend    # http://localhost:5173
```

Es una comodidad de desarrollo, no parte del stack: `docker compose up -d` a secas **no** lo levanta y :5173 queda libre.

El proxy de `/api` de Vite apunta a `VITE_DEV_API_TARGET`, que el servicio inyecta como `http://gateway:8080` (dentro de la red de compose el gateway no se llama `localhost`). Ver `Frontend/vite.config.ts` y `Frontend/README.md`.

Cambiar dependencias del frontend exige rebuildear esa imagen: `docker compose --profile dev build frontend`.

## Verificacion

```bash
curl -i localhost:8080/health                    # sin auth
curl -i localhost:8080/                          # index.html del SPA
curl -i localhost:8080/catalogo                  # deep link -> index.html
curl -i localhost:8080/api/productos             # via gateway -> backend (publico)
curl -i -X POST localhost:8080/api/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@upse.edu.ec","password":"123456"}'
```

TLS llega en U4-T2 (`auto_https` con cert de Let's Encrypt). Por ahora HTTP puro para dev.
