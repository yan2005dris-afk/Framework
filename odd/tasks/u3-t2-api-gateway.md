# U3-T2 — API Gateway + auth-service en Go (extracción)

Implementar el API Gateway como punto de entrada único del SPA **y extraer el auth a un microservicio Go/Fiber independiente** (`auth-service/`). El Backend (catálogo/checkout/MLM) sigue monolito B2C; el SaaS (FastAPI) sigue monolito multi-tenant. Lo que se extrae es SOLO la auth (lo que estaba duplicado entre Backend y SaaS) para evitar duplicar emision/validacion de JWTs.

L1 (usuario): "podriamos implementar un api gateway".
L2 (usuario): "ahora el gateway, para no olvidarnos que toca hacer y yo tambien tener lo encuenta crea issue por cada fase que falta".
L3 (operador): "y si mejoramos la arquitecutra teniendo un microservicio de la auth, un microservicio de los productos y etc" -> opcion full microservicios juzgada demasiada superficie para el silabo academico.
L4 (operador): "hibrido" -> extraer solo auth-service del Backend; mantener Backend y SaaS monolito.
L5 (operador): "Go (consistente con Backend)" -> stack Go/Fiber para auth-service, reusa `Backend/auth/claims.go` y `Backend/middleware/auth.go`.

## Arquitectura objetivo tras U3-T2

```
React SPA (:5173)
       |
   gateway/Caddy (:8080)             <- Caddy 2.x + forward_auth + headers X-User-*
       |
       +---> auth-service/ (:3001)   <- nuevo: emision y validacion de JWTs
       |       |                      endpoints: /auth/login, /auth/logout, /auth/refresh, /auth/validate
       |       +-- PG propia :5436   <- solo `users` y `saas_users` (replica de la BD SaaS, o vista)
       |
       +---> Backend/Go/Fiber (:3000)  <- B2C: catalogo+checkout+MLM. CONSUME auth-service para Login (en lugar de firmarlo el)
       |                              Ya NO tiene su propia bcrypt/JWT; las passwords las pide a auth-service
       |
       +---> Unidad3_SaaS/FastAPI (:8000)  <- SaaS: tenants+users+billing. CONSUME auth-service para su auth
       |     |                            ya NO tiene su propio /api/saas/auth/*; consume el auth-service
       |     +-- PG :5435
       |
       +---> Redis (:6379) compartido
```

## Specs

### Auth-service

- **AS1** — Carpeta nueva `auth-service/` en la raíz con su propio `go.mod`, `main.go`, layout estándar (`controllers/`, `middleware/`, `models/`, `repository/`, `routes/`, `database/`, `migrations/`). Go 1.26, Fiber 2.x, pgx, golang-jwt/v5, bcrypt.
- **AS2** — Endpoints:
  - `POST /auth/login` body `{email, password}` -> 200 con `{token, email, rol, expires}` + cookie `auth_token` httpOnly + cookie `saas_token` httpOnly para tenants con login unificado.
  - `POST /auth/logout` limpia ambas cookies.
  - `POST /auth/refresh` -> emite un token nuevo a partir de uno proximo a expirar (opcional).
  - `POST /auth/validate` -> recibe `{Authorization: Bearer <jwt>` o cookie `auth_token` o `saas_token`} y devuelve 200 con headers-pretendientes: `X-User-Id`, `X-User-Email`, `X-User-Rol`, `X-Tenant-Id`, `X-Rol-Saas`. Devuelve 401 si token invalido/vencido. Usado por Caddy via `forward_auth`.
  - `GET /auth/me` -> devuelve la identidad tomada del token (espejo del `/api/me` legacy).
  - `GET /auth/health` -> healthcheck ping a su BD.
- **AS3** — BD propia del auth-service (`auth_db` en :5436) con `users` (mirror de `usuarios` del Backend: `id`, `email`, `password hash bcrypt`, `rol`) y `saas_users` (mirror de `users_saas` + `tenant_id`). Mantiene sincronizadas con Backend/SaaS via migracion inicial (seed con los mismos usuarios demo). LA SINCRONIZACION OPERATIVA ES MANUAL en esta version (no hay CDC aun); una nota en el README lo dice.
- **AS4** — Reusa `Backend/auth/claims.go` movido a un paquete neutral `auth/` interno del auth-service (o compartido via Go modules replace, pero por simpleza se duplica en `auth-service/auth/claims.go`).
- **AS5** — Firma HS256 con `AUTH_JWT_SECRET` env (default igual a `JWT_SECRET` para compatibilidad). Claims B2C: `{sub, email, rol, iat, exp}`. Claims SaaS: `{sub, email, tenant_id, rol_saas, iat, exp}`.
- **AS6** — Mantiene los cookies httpOnly `auth_token` (B2C) y `saas_token` (SaaS) con SameSite=Lax, Secure=(env!=dev). Una misma sesion fisicamente puede tener ambas si el usuario tiene cuentas en los dos sistemas.

### Gateway (Caddy)

- **G1** — `gateway/Caddyfile` con `:8080` (host) manejando tres rutas:
  - `/api/tx/*` -> reverse proxy `backend:3000` con `header_up X-Forwarded-*` y `forward_auth auth-service:3001` (validates B2C tokens).
  - `/api/saas/*` -> reverse proxy `saas-api:8000` con `forward_auth auth-service:3001` (validates SaaS tokens con claim `tenant_id`).
  - `/` y `/assets/*` -> `file_server` desde un volumen con el `dist/` del frontend (placeholder en U3-T2; el build real es U4).
- **G2** — Cuando `forward_auth` falla (401), Caddy responde 401 directo sin pasar al backend. Cuando pasa (200), copia los headers `X-User-Id`, `X-User-Email`, `X-User-Rol`, `X-Tenant-Id`, `X-Rol-Saas` al request upstream para que el backend los lea.

### Backend (Go/Fiber, refactor mínimo)

- **B1** — `Backend/controllers/authController.go` ahora hace HTTP al auth-service (en lugar de firmar local). Las cookies se setean en el auth-service y se devuelven al cliente via proxy pass-through.
- **B2** — `Backend/middleware/auth.go` sigue existiendo como **defensa en profundidad**: si el gateway está caído o el path se accede internamente, RequireAuth corre su propia validacion HS256 contra `JWT_SECRET` (no depende del gateway). Es duplicación consciente: el gateway es la primera línea, el middleware es la segunda.
- **B3** — `routes.go` elimina los handlers legacy de authController (`authCtrl.Login`, `authCtrl.Logout`, `authCtrl.Me`) y enruta esos paths como proxy al auth-service o los deja responded por el gateway.
- **B4** — Los tests existentes (`controllers/authController_test.go`, `controllers/checkoutController_test.go`) siguen pasando porque el middleware RequireAuth no cambió.

### SaaS (FastAPI)

- **SA1** — `Unidad3_SaaS/app/api/auth.py` se convierte en cliente HTTP al auth-service (mismo patron que B1). Conserva la dependencia `passlib[bcrypt]` y `python-jose` solo si las signatures de JWT se localmente computan (no la decisión: emitir/validar via auth-service).
- **SA2** — Tests actualizados: los smoke tests de `/health` siguen pasando; los nuevos tests de auth son contra el auth-service con mocks.

### docker-compose raíz

- **DC1** — Servicios: `gateway` (Caddy :8080 -> :8080 host), `auth-service` (:3001), `backend` (:3000), `saas-api` (:8000), `redis` (:6379), `db` (PG :5432 para Backend), `saas-db` (PG :5435 para SaaS), `auth-db` (PG :5436 para auth-service). 8 contenedores. Red interna Docker. Único puerto expuesto al host: 8080.
- **DC2** — Variables de entorno centralizadas en `.env` raíz (o por servicio). Convencion: `AUTH_JWT_SECRET` (si es distinto de `JWT_SECRET`), `BACKEND_DB_*`, `SAAS_DB_*`, `AUTH_DB_*`, `REDIS_*`, `CORS_ORIGINS`, `APP_ENV`.

### Tests E2E

- **E2E1** — `tests/e2e/gateway_e2e.py`: con `docker compose up -d`, secuencias curl/bash que verifican:
  - `curl localhost:8080/api/tx/productos` -> 200 con el catálogo.
  - `curl localhost:8080/api/saas/health` -> 200 con JSON del SaaS.
  - Login en backend con curl a `localhost:8080/api/login` setea cookie `auth_token` y devuelve token.
  - Login en SaaS con curl a `localhost:8080/api/saas/auth/login` setea cookie `saas_token`.
  - Request a `/api/tx/red` sin cookie -> 401.
  - Request a `/api/saas/tenants` sin cookie -> 401.
  - Request a `/api/saas/tenants` con `auth_token` B2C -> 401 (wrong token type).
  - Request a `/api/saas/tenants` con `saas_token` rol_saas=superadmin -> 200 con lista.

## Tasks (work-unit commits)

Hay 8 unidades de commit. Cada una cierra un commit autocontenido.

- **GA1 · `feat(auth-service): scaffold Go/Fiber project with go.mod + main + health`** — `auth-service/go.mod`, `auth-service/main.go`, layout estándar, `/auth/health` pingea su BD. Cubre AS1 parcial + AS2 parcial.

- **GA2 · `feat(auth-service): core auth handlers (login, logout, validate, me, refresh)`** — endpoints completos con bcrypt + JWT signing para ambos tipos de tokens (B2C y SaaS). Reusa `Backend/auth/claims.go` (movido). Cubre AS2 + AS4 + AS5 + AS6.

- **GA3 · `feat(auth-service): users + saas_users schemas + Alembic-style SQL migration`** — schema init.sql con seed de los 2 usuarios demo (admin@upse.edu.ec, cliente@upse.edu.ec) y 1 tenant demo con superadmin. Cubre AS3.

- **GA4 · `feat(gateway): Caddyfile + Dockerfile + docker-compose integration`** — `gateway/Caddyfile` con forward_auth y reverse proxy. `gateway/Dockerfile`. Actualizar `docker-compose.yml` raíz con los 8 servicios. Cubre G1 + G2 + DC1.

- **GA5 · `refactor(backend): backend Login/Logout/Me consume auth-service via HTTP`** — `Backend/controllers/authController.go` ya no firma local; hace HTTP al auth-service. `routes.go` deja de registrar esos 3 endpoints locales. Cubre B1 + B3.

- **GA6 · `refactor(saas): SaaS auth endpoints consume auth-service via HTTP`** — `Unidad3_SaaS/app/api/auth.py` cliente HTTP al auth-service. Cubre SA1.

- **GA7 · `test(e2e): bash + curl scripts exercising gateway + auth-service end-to-end`** — `tests/e2e/gateway_e2e.sh` con el listado E2E1. Cubre DC2 + E2E1.

- **GA8 · `docs: architecture update + auth-service README + diagram`** — Actualizar `docs/PROJECT_MAP.md` y `README.md` raíz, `auth-service/README.md` explica su rol. Diagrama ASCII de la nueva arquitectura. Cubre todo.

## Log

- **L1** (usuario, 2026-10-16): "podriamos implementar un api gateway".
- **L2** (usuario, 2026-10-16): "ahora el gateway, para no olvidarnos que toca hacer y yo tambien tener lo encuenta crea issue por cada fase que falta".
- **L3** (usuario, 2026-10-16): "y si mejoramos la arquitecutra teniendo un microservicio de la auth, un microservicio de los productos y etc".
- **L4** (operador): full microservicios descartado por superficie; hibrido extractor solo auth-service.
- **L5** (operador, 2026-10-16): "Go (consistente con Backend)".
- **L6** (implementación, pendiente): GA1-GA8.
