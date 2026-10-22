# REQUIREMENTS CHECKLIST — MultiCatálogo

> Fuente de verdad del progreso. Leer al inicio de cada sesión; marcar `- [x]`
> y actualizar "Estado General" al completar cualquier tarea.
>
> **Estructura:** el documento se divide en fases por **Unidad/Tema del sílabo**
> (`Silabo_Framework_Programacion_Web.md`). Cada fase trae un comentario con lo
> que se hará en esa clase. El producto integrador es UNO: MultiCatálogo —
> una tienda cosmética con red de ventas multinivel (MLM) que crece como
> **sistema políglota**: React (front) + Go/Fiber (núcleo transaccional, U2)
> + FastAPI (capa SaaS/multitenant, U3), desplegado contenerizado (U4).
>
> **Principios rectores (aplican a toda fase):**
> - Seguridad por defecto: ningún endpoint nuevo nace sin auth ni validación.
> - Inyección de dependencias: interfaces en los bordes; globals solo en el entry point.
> - Funciones puras para el dominio (comisiones, totales, validadores) — testeables sin I/O.
> - Mobile-first: UI a 360px, touch ≥44px, un campo por fila en formularios.

## Estado General

| Ítem | Valor |
|---|---|
| **Hito actual** | **U1 completa y U2 completa** (verificación honesta tras merges #10 + #11, marzo-octubre 2026). U1-T1..T5 y U2-T1..T4 cumplidos. Pendientes menores en U1-T2 (renombrado + .env), U1-T5 (audit mobile-first), U2-T2 (`PORT` default 3000 vs 8080; `models.APIError` literal vs funcionalmente equivalente con helpers), U2-T3 (índices `referidos(parent_id)` y `productos(categoria)` y migración `003_seguridad.sql` por verificar/crear). |
| **Última tarea completada** | Sincronización del checklist contra el código actual tras mergear PR #10 (backend U2-T4) y PR #11 (frontend cookie httpOnly + contexts split + migración 002). |
| **Siguiente tarea pendiente** | 1) Cerrar pendientes chicos: renombrar `package.json`, crear `Frontend/.env` con `VITE_API_URL`, audit mobile-first 375px / botones ≥44px, migración `003_seguridad.sql` con índices faltantes, cambio de `PORT` default a 3000 + uso literal de `models.APIError{status,message,details}` en controllers. 2) Después: **U3-T1 scaffold FastAPI** `Unidad3_SaaS/` (capa SaaS/multitenant, puerto :8000). 3) **U4** containerización con Caddy/Nginx + SSL. |
| **Bloqueadores** | Ninguno funcional. `backend-haskell` detenido (`docker start backend-haskell` restaura); PG propio debe correr en 5434 (ver `Backend/docker-compose.yml`). Recordatorio de seguridad: `Backend/.env` no está versionado — generar `JWT_SECRET` con `openssl rand -hex 48` para deploy y mantener `APP_ENV=prod` ahí. |

## Arquitectura objetivo del semestre

```
React SPA (:5173) ──► Caddy/Nginx (U4: proxy inverso + SSL)
                        ├─ /api/tx/*   → Go/Fiber  :3000  — núcleo transaccional (U2)
                        └─ /api/saas/* → FastAPI   :8000  — capa SaaS/multitenant (U3)
                                            │
                          PostgreSQL 16 (:5434) + Redis (:6379)
```

---

# UNIDAD 1 — FRAMEWORKS FRONTEND (SPA React + Vite + Tailwind)

## U1-T1 · Introducción a frameworks web y arquitectura de soluciones

> Se presenta la pila tecnológica del semestre, la inversión de control y los
> lineamientos del producto integrador MultiCatálogo (tienda + red MLM).

- [x] Pila definida y workspace estructurado. Tras los merges el repo consolidó en `Backend/` y `Frontend/` planos en la raíz (no en `Unidad1_Frontend/Unidad2_Backend/` como sugería este doc). `docs/` y `odd/` separados.

## U1-T2 · Entorno SPA con Vite + Tailwind CSS

> Scaffold del frontend: empaquetador Vite, Tailwind v4, estructura de carpetas
> por responsabilidad (`components/`, `context/`, `services/`, `data/`).

- [x] Entorno SPA con Vite + Tailwind v4 (`@tailwindcss/vite ^4.3.3`, `tailwindcss ^4.3.3`, `vite ^8.3.0` en `Frontend/package.json`)
- [ ] `package.json` renombrado a `multicatalogo-frontend` — sigue como `"name": "proyecto"` (cosmético, no bloquea nada)
- [x] `VITE_API_URL` consumido vía `import.meta.env` — `Frontend/src/services/http.ts` lee `import.meta.env.VITE_API_URL || "http://localhost:8080/api"`. **Falta**: crear `Frontend/.env` versionado con `VITE_API_URL=http://localhost:8080/api` para entornos no-dev.

## U1-T3 · Componentes y gestión de estado global

> Estado global con React Context: sesión de usuario y carrito por usuario.

- [x] `AuthProvider` (sesión `{email, rol}`) en `Frontend/src/Context/AuthProvider.tsx` y `CartProvider` (carrito en localStorage por email, prefijo `multicatalogo_carrito_<email>`) en `Context/CartProvider.tsx`
- [x] Persistencia de sesión en AuthContext: clave única `mc_session` con `{email, rol}` en localStorage. **El token NO se persiste localmente** (vive solo en la cookie httpOnly `auth_token` que configura `/api/login`).
- [x] Split `auth-types.ts` + `auth-context.ts` + `AuthProvider.tsx` + `useAuth.ts` y analogamente `cart-context.ts` + `CartProvider.tsx` + `useCart.ts`. `npm run lint` retorna 0 errores, 0 warnings sin necesidad de `eslint-disable react-refresh/only-export-components`.

## U1-T4 · Consumo de API RESTful y enrutamiento

> Promesas con fetch, capa de servicios desacoplada y guardas de ruta por rol.

- [x] `services/productosService.ts` y `services/redService.ts` reescritos para usar `apiFetch` + guardas `ProtectedRoute`/`AdminRoute` en `App.tsx` con `useAuth()` + `<Outlet/>` y spinner mientras se hidrata sesión contra `/api/me`.
- [x] `services/http.ts` con `apiFetch` único (`credentials: "include"`) y `ApiError` tipado con `{status, message, detalle?}`. **Nota**: la implementación real usa cookie httpOnly `auth_token` en lugar de `Authorization: Bearer`. Es estrictamente mejor (mitiga XSS sobre el token); el chequeo de autorización en backend acepta ambos caminos.

## U1-T5 · Interfaces full-screen y flujos MLM (Práctica 03)

> Vistas inmersivas del catálogo y flujo completo de compra y red multinivel.

- [x] F8 Login con rol + redirección por rol (`Login.tsx` → guarda rutas y `AdminRoute` envía a `/tienda` cuando rol ≠ admin)
- [x] F1 Storefront `/tienda` (full-screen) | F3 Catálogo con búsqueda/filtros (`Catalogo.tsx`) | F2 Detalle `/producto/:id` (`DetalleProducto.tsx`) | F4 Lightbox (`lightboxAbierto` en `DetalleProducto.tsx`)
- [x] F5 Carrito + checkout + confirmación (rutas en `App.tsx`) | F6 Árbol multinivel (`MiRed.tsx`) | F7 Dashboard KPIs (`Dashboard.tsx`)
- [x] `Checkout.tsx` → `POST /api/checkout` real (sin simulación) usando `api.checkout` que lanza `ApiError` tipado con `status` si la respuesta no es 2xx.
- [ ] Audit mobile-first: 375px sin overflow en Dashboard/MiRed/Catálogo; botones de acción ≥44px (`Añadir al Carrito`, steppers de `Carrito`, `Cerrar Sesión`). No se ha hecho de forma sistemática todavía. Pendiente.

---

# UNIDAD 2 — BACKEND ALTA TRANSACCIONALIDAD (Go/Fiber · "entornos bancarios")

## U2-T1 · Rendimiento, memoria y concurrencia

> Goroutines y contextos: procesos de misión crítica que no bloquean la respuesta HTTP.

- [x] Goroutine `procesarComisiones` tras el COMMIT del pedido (contexto propio, no el del request): `checkoutController.go:64-72` lanza `go func(total, uID, email)` que invoca `ActualizarComisiones(...)`; no bloquea la respuesta HTTP.

## U2-T2 · Microservicios y APIs robustas (Fiber)

> API REST en capas con contrato de error estándar y configuración por entorno.

- [x] Capas routes → controllers → repository; endpoints `login`, `logout`, `me`, `productos`, `productos/:id`, `red`, `checkout`, `health` registrados en `Backend/routes/routes.go` con DI
- [ ] `APIError{status,message,details}` en todos los errores — `models.APIError` sigue **definido pero los controllers devuelven `fiber.Map{"error":..., "detalle":?}`**. Funcionalmente equivalente via el helper `controllers/helpers.SanitizeErrorDetails(appEnv, err)`. Migración al struct literal es limpieza pendiente.
- [ ] Puerto de escucha configurable (`PORT`, defecto 3000) — **parcial**: `main.go` lee `PORT` con default `"8080"`. Funcional, difiere el default documentado.
- [x] `/api/health` real con ping a Postgres (503 si cae) y Redis (`"degraded"` si cae) — `controllers.NewHealthController(dbPool, redisClient).Health(c)` pingea ambos con timeout 2s. Verificado: `{"status":"ok","components":{"postgres":"up","redis":"up"}}`.

## U2-T3 · Persistencia PostgreSQL + caché Redis

> Dockeriza la BD, SQL puro con pgx, reconstrucción del árbol y Cache-Aside.

- [x] `docker-compose` Postgres 16-alpine + Redis-alpine + volumen `postgres_data` + healthchecks (`pg_isready`, `redis-cli ping`)
- [x] `init.sql` (tablas `usuarios`/`productos`/`referidos`/`pedidos`/`pedido_items` + seed) y `Backend/database/migrations/002_checkout.sql` para volúmenes pre-existentes. **Falta**: `database/migrations/003_seguridad.sql` (no usada hasta ahora; la seguridad se aplicó via init.sql actualizado). Migración futura útil para agregar los índices faltantes (`referidos(parent_id)`, `productos(categoria)`) sobre volúmenes ya inicializados.
- [x] `repository.go` a SQL real con `pgxpool`, árbol de referidos por `parent_id`
- [x] Cache-Aside funcional: `productos:all` con TTL 300s (5 min) en `ProductoController`. `RedController` usa key `"red:tree"` (sin `:email`) — la letra del doc sugería `red:tree:{email}`. Degradación sin caché correcta (try/recatch).
- [ ] Índices explícitos: `referidos(parent_id)` y `productos(categoria)` por crear; `pedidos(cliente_email)` y `pedido_items(pedido_id)` ya en `002_checkout.sql`. Pendiente una migración `003_indices.sql` (o equivalente) para los dos primeros.
- [x] Errores propagados: firmas del repo devuelven `error`; `ErrNotFound` se retorna en 3 paths del repo. Mapeo en controllers consistente tras los PR #10/#11: producto/red mapean `ErrNotFound` a 404, login mapea a 401, checkout mapea no-nil a 500.

## U2-T4 · Inyección de dependencias y microservicio seguro

> DI por interfaces en toda la cadena + transacción ACID del checkout + hardening.

- [x] Interfaces `DatabaseRepository`/`CacheRepository` (`repository/interfaces.go`); `PostgresRepository`/`RedisCacheRepository` (`repository/repository.go`); structs de controllers con DI; `main.go` cablea con `NewAuthController(dbRepo, jwtSecret, appEnv)`/`NewProductoController(dbRepo, cacheRepo, appEnv)`/`NewRedController(dbRepo, cacheRepo, appEnv)`/`NewCheckoutController(dbRepo, cacheRepo, appEnv)`/`NewHealthController(dbPool, redisClient)`
- [x] `POST /api/checkout` con transacción ACID en `repository.CrearPedido`, `commit`/`rollback` propagado, total recalculado en servidor desde BD
- [x] Invalidación de caché tras compra: la goroutine en `checkoutController.Checkout` invalida `productos:all` + `red:tree` + `red:tree:{email}` (donde el email es del TOKEN, no del body). Migración `002_checkout.sql` sí existe y aplicada.
- [x] bcrypt de passwords — `authController.Login` compara con `bcrypt.CompareHashAndPassword`. Los seeds en `init.sql` ya están hasheados con bcrypt cost 10 para `admin@upse.edu.ec` y `cliente@upse.edu.ec` (password de dev = `123456`).
- [x] JWT real con cookie httpOnly — `authController.Login` firma HS256 con claims `{sub,email,rol,iat,exp}`, 2h. `JWT_SECRET` viene de env. Cookie `auth_token` con `HTTPOnly=true, SameSite=Lax, Secure=(appEnv != "dev")`. Body JSON ya no incluye token. Paquete `Backend/middleware/` existe con `RequireAuth(AuthConfig{JWTSecret, DBRepo})` y `RequireAdmin()`. También el paquete `Backend/auth/` neutral con `AuthClaims` para evitar ciclo controllers↔middleware.
- [x] Rutas `/api/me`, `/api/red` y `/api/checkout` protegidas por `middleware.RequireAuth(jwtSecret)`. Sin cookie/header válido, devuelven 401.
- [x] Checkout toma la identidad del TOKEN (`middleware.LocalUsuarioID` y `middleware.LocalEmail` desde `c.Locals`). `controllers/checkoutController.Checkout` sobrescribe `req.EmailCliente` con el email del token ANTES de invocar `CrearPedido`. **Verificado vía curl**: body con `atacante@example.com` se persiste como `admin@upse.edu.ec` en `pedidos.email_cliente`. Comisiones se atribuyen a `usuarioID` del token.
- [x] Errores sanitizados: helper `controllers/helpers/SanitizeErrorDetails(appEnv, err)` aplica el principio de mínima divulgación — `detalle` solo se incluye si `APP_ENV=="dev"`. `repository.ErrCantidadInvalida` propio (errors.New("cantidad debe ser mayor que cero")), mapeado a 400 con ese mensaje literal en `checkoutController.Checkout`.

---
