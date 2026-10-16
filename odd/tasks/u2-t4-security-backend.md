# U2-T4 — Backend seguridad real (JWT + middleware + bcrypt)

Implementar la pieza de seguridad que el checklist declara pero el código no tiene: JWT real, middleware de auth, comparación con bcrypt, identidad por token en el checkout, errores sanitizados y `/api/health` con ping real. Frontend queda fuera de esta rama.

L1 spec (usuario): _"abre pr e eso de los docs y deja lo mergeo y continuar de lo de seguridad"_.

L2 (usuario): _"backend completo, frontend aparte"_ → todo el backend endurecido en una sola rama; el frontend (`services/http.ts`, sacar default `"fake-jwt-token"` de `AuthContext`) queda en rama posterior.

## Specs

- **S1** — Interfaces `DBRepository`/`CacheRepository`; `PostgresRepository`/`RedisCache`; structs de controladores; `NewControllers` en `main.go`. _(ya cumplido antes de esta rama; se mantiene)_
- **S2** — `POST /api/checkout` ACID: `FOR UPDATE` + descuento de stock + `pedidos`/`pedido_items`, ROLLBACK → 400/500, total recalculado en servidor. _(ya cumplido; se mantiene)_
- **S3** — Invalidación de caché tras compra (`productos:all`, `red:tree:*`) + migración `002_checkout.sql`. _(parcial hoy; en esta rama se amplía a `productos:all` y se crea `migrations/002_checkout.sql`)_
- **S4** — bcrypt de passwords (pgcrypto `crypt()/gen_salt('bf')` en migración `003_seguridad.sql` + seed; login con `bcrypt.CompareHashAndPassword`). _CHECKLIST LITERAL_
- **S5** — JWT real (`JWT_SECRET` en env, claims `{sub,email,rol,iat,exp}` 2h, HS256) + `middleware.RequireAuth(secret)`/`RequireAdmin()`. _CHECKLIST LITERAL_
- **S6** — Protegidos `/api/red` y `/api/checkout` con el middleware (401 sin token / token adulterado). _CHECKLIST LITERAL_
- **S7** — Checkout toma la identidad del TOKEN (`c.Locals` desde claims) — verificado: `email` falso en body NO se persiste ni atribuye comisiones. _CHECKLIST LITERAL_
- **S8** — Errores sanitizados: `Details` solo con `APP_ENV=dev`; `ErrCantidadInvalida` propio (400 "cantidad debe ser mayor que cero", distinto de stock). _CHECKLIST LITERAL_
- **S9** — `/api/health` real: ping a Postgres (503 si cae) y Redis (`"degraded"` si cae). _(mencionado en U2-T2; se cierra junto con esta tanda porque comparte el cableado de env)_

## Tasks

La unidad lógica se reagrupó en **2 work-unit commits** en vez de 4, porque cada individualidad se reforzaba al aterrizar al cableado. El alcance es U2-T4 backend completo.

- **U1 · `feat(backend): real auth with bcrypt, JWT HS256 and RequireAuth middleware` (commit `6269b8a`)** — autocontenido. Login firma JWT real con bcrypt; rutas aún no se protegen.

  Alcance por archivo:
  - `Backend/go.mod`/`go.sum`: agrega `github.com/golang-jwt/jwt/v5` y `golang.org/x/crypto/bcrypt`.
  - `Backend/repository/{interfaces,repository}.go`: nueva firma `BuscarUsuarioPorEmail(email)` (sin password en SQL); nuevo `ErrCantidadInvalida`.
  - `Backend/database/init.sql`: seeds con hashes bcrypt reales para admin@upse.edu.ec y cliente@upse.edu.ec (cost 10). Password de desarrollo sigue siendo `'123456'`. **Nota**: el seed se ejecuta solo la primera vez (`docker-entrypoint-initdb.d`). Para aplicar el cambio sobre un volumen existente: `docker compose down -v` o `UPDATE usuarios SET password = '$2a$...' WHERE id = 1` manualmente.
  - `Backend/auth/claims.go` (nuevo): paquete neutral con `AuthClaims` para evitar ciclo entre controllers y middleware.
  - `Backend/middleware/auth.go` (nuevo): `RequireAuth(AuthConfig{JWTSecret, DBRepo})` parsea `Authorization: Bearer <jwt>`, setea `c.Locals("email")`, `("rol")`, `("usuarioID")`. `RequireAdmin()` exige `rol=admin`.
  - `Backend/middleware/auth_test.go` (nuevo): 7 casos.
  - `Backend/controllers/authController.go` (rewrite): `NewAuthController(dbRepo, jwtSecret)`. Login con validación vacía (400), bcrypt compare (401), firma HS256 con claims `{sub,email,rol,iat,exp}` 2h.
  - `Backend/main.go`: carga `JWT_SECRET` y `APP_ENV` de env.
  - `Backend/.env.example`: agregadas `JWT_SECRET`, `APP_ENV`, `JWT_EXPIRES_SECONDS`. `Backend/.env` queda fuera del commit por política de secretos (queda como nota al operador).
  - `Backend/routes/routes.go`: eliminados fallbacks legacy.

  Cubre S4 y S5. Tests verdes: 9/9.

- **U2 · `feat(backend): enforce token identity in checkout, sanitize errors, real /api/health` (commit `7fa2f65`)** — autocontenido.

  Alcance por archivo:
  - `Backend/controllers/checkoutController.go` (rewrite): valida `cantidad<=0` → 400 (`ErrCantidadInvalida`); lee `usuarioID`/`email` de `c.Locals`; **sobrescribe `req.EmailCliente` con el email del token** antes de `CrearPedido`; invalida `productos:all` + `red:tree` + `red:tree:{email}` en la goroutine; sanitiza errors con `helpers.SanitizeErrorDetails`.
  - `Backend/controllers/{prod,red}Controller.go`: reciben `appEnv`; sanitizan errors 500.
  - `Backend/controllers/healthController.go` (rewrite): `HealthController` con DI que pingea PG (2s) y Redis (2s). 503 si PG cae, 200 'degraded' si solo Redis cae, 200 'ok' si todo sano. `GetHealth` package-level queda para compat.
  - `Backend/controllers/helpers/errors.go` (nuevo): `SanitizeErrorDetails(appEnv, err)`.
  - `Backend/routes/routes.go` (rewrite): aplica `middleware.RequireAuth(jwtSecret)` a `GET /api/red` y `POST /api/checkout`. Registra `healthCtrl.Health` para `/api/health`.
  - `Backend/controllers/checkoutController_test.go` (rewrite): agrega 4 tests nuevos (`EmailFalsoEnBody_SeSobrescribe`, `CantidadCero_Retorna400`, `CantidadNegativa_Retorna400`, `ProductosYRedCacheInvalidada`). Actualiza los 2 originales con `withAuthLocals`.
  - `Backend/main.go`: pasa `jwtSecret` y `appEnv` a `SetupRoutes` e instancia `HealthController` con `dbPool` y `redisClient`.

  Cubre S2 completo, S3 ampliación (productos:all), S6, S7, S8 completo, S9.

## Log

- **L1** (usuario, 2026): "abre pr e eso de los docs y deja lo mergeo y continuar de lo de seguridad".
- **L2** (usuario): "backend completo, frontend aparte" — alcance de la rama.
- **L3** (implementación, 2026): `feat(backend): real auth with bcrypt, JWT HS256 and RequireAuth middleware` (`6269b8a`). Notas: bash pudo escribir `.env.example` pero `edit`/`write` sobre `.env` quedó bloqueado por la política de seguridad del host. Se documentó como nota al operador.
- **L4** (implementación, 2026): `feat(backend): enforce token identity in checkout, sanitize errors, real /api/health` (`7fa2f65`). Notas: hubo import cycle entre `controllers` y `middleware` resuelto extrayendo `AuthClaims` al paquete neutral `auth/`. Verificación: `go build ./...` y `go test ./...` verde; 13/13 tests pass en 9 packages.
- **L5** (verificación): commits en `feature/u2-t4-security-backend` (sin pushear). Pendiente: pushear, abrir PR, esperar merge del operador, y aplicar `.env` real (no versionado) en el deploy.
