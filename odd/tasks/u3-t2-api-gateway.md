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
       +---> auth-service/ (:3001)   <- emision y validacion de JWTs (escriben schema 'auth')
       |
       +---> Backend/Go/Fiber (:3000)  <- B2C: catalogo+checkout+MLM (escriben schema 'backend')
       |
       +---> Unidad3_SaaS/FastAPI (:8000)  <- SaaS: tenants+users+billing (escriben schema 'saas')
       |
       +---> Redis (:6379) compartido
       |
       +---> POSTGRES (:5432)          <- UNA sola instancia, esquemas separados
              auth.users, auth.saas_users, auth.saas_tenants
              backend.usuarios, backend.productos, backend.referidos, backend.pedidos, ...
              saas.planes, saas.suscripciones (futuro)
```

### Decisión arquitectónica clave (L7 operador): BD ÚNICA con schemas

Decidido vs. una BD-por-servicio por simplicidad y terminabilidad en el semestre:

- **Una sola BD Postgres en :5432** (la que ya usa el Backend).
- Cada servicio **posee un schema** y escribe solo ahí:
  - `auth-service` -> schema `auth` (users, saas_users, saas_tenants).
  - `backend-service` -> schema `backend` (usuarios, productos, referidos, pedidos, pedido_items).
  - `saas-service` -> schema `saas` (planes, suscripciones, futuras tablas tenant-scoped).
- **FK cross-schema funcionan** en Postgres porque todos estan en la misma BD (`REFERENCES auth.users(id)`).
- **Lectura cruzada** permitida via vistas controladas: por ejemplo, backend puede consultar `auth.users` para resolver el email del usuario autenticado sin escribir.
- **Sync eliminado**: los usuarios viven una sola vez en `auth.users`. El `backend.usuarios` legacy queda como espejo (solo lectura) para no romper `Backend/repository/repository.go` actual que lo lee de `usuarios`. Migración futura: eliminar `backend.usuarios` y leer de `auth.users` via vista.
- **Sync eliminado NO significa zero cambios**: el auth-service crea `auth.users` con los mismos seeds que `backend.usuarios` tiene hoy (admin@upse.edu.ec, cliente@upse.edu.ec con bcrypt), y la migración inicial del Backend se mantiene. Las dos tablas viven hasta que decidamos unificarlas via vista.

## Specs (actualizado para BD única + schemas)

### Auth-service

- **AS1** — Carpeta nueva `auth-service/` en la raíz con su propio `go.mod`, `main.go`, layout estándar (`controllers/`, `middleware/`, `models/`, `repository/`, `routes/`, `internal/db/`). Go 1.26, Fiber 2.x, pgx, golang-jwt/v5, bcrypt.
- **AS2** — Endpoints:
  - `POST /auth/login` body `{email, password}` -> 200 con `{token, email, rol, expires}` + cookie `auth_token` httpOnly + cookie `saas_token` httpOnly si el usuario esta vinculado a un tenant SaaS.
  - `POST /auth/logout` limpia ambas cookies.
  - `POST /auth/refresh` -> emite un token nuevo (opcional).
  - `POST /auth/validate` -> recibe el JWT (via header Authorization o cookie `auth_token`/`saas_token`) y devuelve 200 con los headers que Caddy pasa al upstream: `X-User-Id`, `X-User-Email`, `X-User-Rol`, `X-Tenant-Id`, `X-Rol-Saas`. 401 si invalido. Usado por Caddy via `forward_auth`.
  - `GET /auth/me` -> identidad del token.
  - `GET /auth/health` -> healthcheck ping a Postgres.
- **AS3** — **Una sola BD Postgres** en :5432 (la del Backend). El auth-service escribe en schema `auth`:
  - `auth.users (id, email, password_hash, rol, created_at)`: source-of-truth para login. Seed: `admin@upse.edu.ec` con bcrypt (`123456`), `cliente@upse.edu.ec` con bcrypt.
  - `auth.saas_users (id, user_id FK auth.users, tenant_id FK auth.saas_tenants, email, hashed_password, rol_saas)` (futuro, en GA3 si lo hacemos).
  - `auth.saas_tenants (id, slug UNIQUE, name, created_at)` (opcional, en GA3).
- **AS4** — Reusa `Backend/auth/claims.go` movido a un paquete neutral `auth-service/auth/claims.go` (mismo codigo, paquete propio del servicio).
- **AS5** — Firma HS256 con `AUTH_JWT_SECRET` env (default igual a `JWT_SECRET` del Backend para compatibilidad). Claims B2C: `{sub, email, rol, iat, exp}`. Claims SaaS: `{sub, email, tenant_id, rol_saas, iat, exp}`.
- **AS6** — Cookies httpOnly `auth_token` (B2C) y `saas_token` (SaaS) con SameSite=Lax, Secure=(env!=dev).

### Gateway (Caddy)

- **G1** — `gateway/Caddyfile` con `:8080`:
  - `/api/tx/*` -> `backend:3000` con `forward_auth auth-service:3001`.
  - `/api/saas/*` -> `saas-api:8000` con `forward_auth auth-service:3001`.
  - `/` -> file_server (placeholder hasta U4 con SPA build).
- **G2** — `forward_auth` falla -> 401 directo; pasa -> copia headers `X-User-*` al upstream.

### Backend (Go/Fiber, refactor mínimo)

- **B1** — `Backend/controllers/authController.go` hace HTTP a `auth-service:3001/auth/login` (forwarda el body y devuelve la respuesta tal cual).
- **B2** — `Backend/middleware/auth.go` se mantiene como **defensa en profundidad**: si el gateway está caído, RequireAuth corre su propia validacion HS256 contra `JWT_SECRET`. El gateway es la primera linea; el middleware la segunda.
- **B3** — `routes.go` reemplaza los handlers locales `authCtrl.Login/Logout/Me` con un proxy reverso al `auth-service:3001`.
- **B4** — Los tests existentes del Backend (`controllers/authController_test.go`, `controllers/checkoutController_test.go`) siguen pasando porque el middleware `RequireAuth` no cambia.

### SaaS (FastAPI)

- **SA1** — `Unidad3_SaaS/app/api/auth.py` se convierte en cliente HTTP a `auth-service:3001/auth/login`.
- **SA2** — Tests de smoke de `/health` siguen pasando; nuevos tests de auth usan `httpx` AsyncClient mocking al auth-service.

### docker-compose raíz

- **DC1** — Servicios: `gateway` (Caddy :8080), `auth-service` (:3001), `backend` (:3000), `saas-api` (:8000), `redis` (:6379), `db` (Postgres :5432). **6 contenedores** en total (antes 8). Único puerto expuesto al host: 8080.
- **DC2** — Variables de entorno centralizadas en `.env` raíz: `AUTH_JWT_SECRET`, `BACKEND_DB_*`, `SAAS_DB_*`, `REDIS_*`, `CORS_ORIGINS`, `APP_ENV`.

### Tests E2E

- **E2E1** — `tests/e2e/gateway_e2e.sh`: con `docker compose up -d`, verifica:
  - `curl localhost:8080/api/tx/productos` -> 200 con el catálogo.
  - `curl localhost:8080/api/saas/health` -> 200 con JSON del SaaS.
  - `curl -X POST localhost:8080/api/auth/login ...` -> setea cookie `auth_token` + `saas_token` si aplica.
  - `curl -X POST localhost:8080/api/saas/auth/login ...` -> setea cookie `saas_token`.
  - `curl localhost:8080/api/tx/red` sin cookie -> 401.
  - `curl localhost:8080/api/saas/tenants` sin cookie -> 401.
  - `curl localhost:8080/api/saas/tenants` con `auth_token` B2C -> 401 (wrong token type).
  - `curl localhost:8080/api/saas/tenants` con `saas_token` rol_saas=superadmin -> 200 con lista.

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
- **L6** (operador, 2026-10-16): "aunque orita de momento no seamos purista de microservicio y usemos la misma base de datos o hagamos tablas foraneas entre diferentes base de dtos a ver que tal nos queda, o como manejamos en las otras base de datos de go de monilitico que hace? osea los usuarios como se sabe ahi que usuario hace pago o cosas asi" — decision: BD unica compartida con schemas separados.
- **L7** (operador): opcion tecnica elegida "schemas compartidos".
- **L8** (implementación): `2adf473` refactoriza GA1 para apuntar a la BD compartida en :5432/schema 'auth' en lugar de la BD propia :5436. Sigue siendo GA1 con la nueva decisión arquitectónica.
- **L9** (implementación): PRs #22, #23, #24, #25 — GA1 (scaffold + GA1 update), GA2 (handlers), GA3 (schema), GA4 (gateway + compose) mergeados a main. End-to-end de login via auth-service verificado con curl.
- **L10** (operador, 2026-10-16): la extension de receipt-driven development intento correr review nativo del provider del workflow de GA4 con 2 lenses (review-risk + review-reliability, forecast 2 model runs). El `capture_group` fue rechazado por `capture-group-rejected` ("must be the complete distinct current reviewer group in exact provider order") con los bindings exactos del STATUS. Se intento `capture` individual con forecast successful + ack subsiguiente: `capture-binding-rejected` ("missing or stale for current STATUS") sin un invariante conocido que el operador o el provider pueda afectar. Decision del operador: skip el native review (la lineage queda en `reviewing` sin emitir; la extension recordara de nuevo en futuras mutaciones).
- **L11** (implementación, en curso): `feature/u3-t2-ga5-backend-consumes-auth-service`. PR #25 mergeado. Inicia GA5 (Backend consume auth-service via HTTP).
- **L11.5** (operador, 2026-10-16): patron de skip repetido. La extension volvio a disparar al commitar GA5 (target sha256:b0f2c08caa236b80bbd72c3b6e7f644ab64e33acd6135128eafecaaf2fb8075c, lineage review-ee330992e6496516, 6 paths). El operador decide **skip otra vez**, consolidando la politica: las reviews nativos del provider no son utiles en este repo; la revision de calidad ocurre via PR desde la UI. Las futuras mutaciones probablemente disparan el mismo reminder - aceptar que es asi.
- **L11.6** (operador, 2026-10-16): tras un tercer reminder (sobre el commit de ODD log L11.5, target sha256:3f5c7bafa5ce3254ce57aba593a62f201c0020a350798c420733284a634bb24c), el operador establece **politica permanente**: skip todos los reminders futuros de native review sobre este repo. Yo ya no preguntare mas al respecto ni gastare tokens llamando gentle_review; el reminder del host seguira disparandose por mutaciones pero sera ignorado. Si el operador quiere reactivar native review en algun momento, lo dice explicitamente y se re-evalua.
- **L12** (implementación, pendiente): GA6 (SaaS consume auth-service), GA7 (tests E2E), GA8 (docs finales).
- **L13** (operador, 2026-10-16): PR #26 mergeado (GA5 completo, Backend consume auth-service). Iniciamos **GA6: SaaS consume auth-service**. Spejo de GA5 sobre el lado Python/FastAPI: app/api/auth.py como proxy httpx al auth-service, montado en app/main.py, tests con MockTransport.
- **L14** (operador, 2026-10-16): PR #27 mergeado. Iniciamos **GA7: tests E2E bash/curl**. Cobertura: health (gateway + saas), products publico, red sin auth (401), login admin (cookie), login bad-pwd, me autenticado, me sin cookie, saas-login via proxy, 404. PR con tests/e2e/gateway_e2e.sh + tests/e2e/README.md.
- **L15** (operador, 2026-10-16): PR #28 mergeado (GA7 E2E). Iniciamos **GA8: docs finales** (`docs/architecture.md` con diagrama ASCII del stack + actualizar `docs/PROJECT_MAP.md` y `README.md` raiz).
