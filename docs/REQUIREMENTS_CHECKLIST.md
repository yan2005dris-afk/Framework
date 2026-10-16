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
| **Hito actual** | **U1 y U2 base completas (verificación honesta)**: SPA con sesión persistida + carrito por usuario + guardas de ruta + backend con capas (routes/controllers/repository), DI, ACID checkout, Goroutine de comisiones, Cache-Aside Redis, docker-compose Postgres+Redis. **Pendiente**: endurecimiento de seguridad real (bcrypt + JWT HS256 + middleware RequireAuth + checkout por token), renombrado de `package.json` y `Frontend/.env`, aplicación consistente de `APIError` en controllers, `/api/health` con ping real a Postgres/Redis. |
| **Última tarea completada** | Recalibración honesta del checklist contra el código real y verificación `go build ./...` → EXIT_BUILD=0 (sanidad del backend). |
| **Siguiente tarea pendiente** | 1) Endurecer U2-T4: `Backend/middleware/` con `RequireAuth`/`RequireAdmin` (HS256, `JWT_SECRET`), bcrypt en `init.sql` seed, login devuelve JWT firmado, `routes.go` aplica middleware a `/api/red` y `/api/checkout`, checkout toma email de `c.Locals` (ignora `emailCliente` del body). 2) Renombrar `package.json` → `multicatalogo-frontend` y crear `Frontend/.env` con `VITE_API_URL=http://localhost:8080/api`. 3) `Frontend/src/services/http.ts` con `apiFetch` que inyecte `Authorization: Bearer` y `ApiError` tipado con `status`. 4) `/api/health` con ping real (503 si Postgres cae, `"degraded"` si Redis cae). 5) Luego U3-T1: scaffold FastAPI `Unidad3_SaaS/`. |
| **Bloqueadores** | Ninguno funcional. `backend-haskell` detenido (`docker start backend-haskell` restaura); `postgres-funcional` en :5433 → Postgres propio debe correr en **5434** (ver `Backend/.env.example`/`docker-compose.yml`). |

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
- [ ] `package.json` renombrado a `multicatalogo-frontend` — sigue como `"name": "proyecto"`
- [ ] `VITE_API_URL` en `Frontend/.env` + `import.meta.env` (cliente único en `services/http.ts`) — `Frontend/.env` no existe; `services/api.ts` usa `import.meta.env.VITE_API_URL || "http://localhost:8080/api"` (consume la env var pero el archivo no está versionado)

## U1-T3 · Componentes y gestión de estado global

> Estado global con React Context: sesión de usuario y carrito por usuario.

- [x] `AuthContext` (sesión `{email, rol}`) en `Frontend/src/Context/AuthContext.tsx` y `CartContext` (carrito en localStorage por email, prefijo `multicatalogo_carrito_<email>`) en `Context/CartContext.tsx`
- [x] Persistencia de sesión en `AuthContext` (claves separadas: `token`, `userEmail`, `userRol` en localStorage; **no** la clave única `multicatalogo_sesion` sugerida originalmente — funciona, difiere el naming)
- [ ] Warnings eslint resueltos con split `context/auth.ts` + `context/cart.ts` (hooks/tipos separados del Provider) y `npm run lint` → 0 errores, 0 warnings — actualmente `Context/AuthContext.tsx` y `CartContext.tsx` son un único archivo cada uno con hooks+tipos+Provider mezclados y usan `/* eslint-disable react-refresh/only-export-components */`. `npm run lint` retorna 0/0 pero **silenciando** reglas, no resolviéndolas con el split.

## U1-T4 · Consumo de API RESTful y enrutamiento

> Promesas con fetch, capa de servicios desacoplada y guardas de ruta por rol.

- [x] `services/productosService.ts` y `services/redService.ts` (desacoplan la capa de datos del consumidor) + guardas `ProtectedRoute`/`AdminRoute` definidas en `App.tsx` con `useAuth()` + `<Outlet/>` y redirección por rol
- [ ] `services/http.ts`: `apiFetch` único con `Authorization: Bearer` automático, `ApiError` tipado con status y helpers de sesión — actualmente la lógica vive en `services/api.ts` como objeto `api = { login, getProductos, checkout }` con `fetch` directo. **No** inyecta Bearer. Devuelve `Error` plano con `data.error`, no `ApiError` con `status`. **No** existe `services/http.ts`.

## U1-T5 · Interfaces full-screen y flujos MLM (Práctica 03)

> Vistas inmersivas del catálogo y flujo completo de compra y red multinivel.

- [x] F8 Login con rol + redirección por rol (`Login.tsx` → guarda rutas y `AdminRoute` envía a `/tienda` cuando rol ≠ admin)
- [x] F1 Storefront `/tienda` (full-screen) | F3 Catálogo con búsqueda/filtros (`Catalogo.tsx`) | F2 Detalle `/producto/:id` (`DetalleProducto.tsx`) | F4 Lightbox (`lightboxAbierto` en `DetalleProducto.tsx`)
- [x] F5 Carrito + checkout + confirmación (rutas en `App.tsx`) | F6 Árbol multinivel (`MiRed.tsx`) | F7 Dashboard KPIs (`Dashboard.tsx`)
- [x] `Checkout.tsx` → `POST ${API_BASE_URL}/checkout` real (sin simulación); **sin** tipo `ApiError` propio: usa `Error` plano propagado por `api.checkout`
- [ ] Audit mobile-first: 375px sin overflow en Dashboard/MiRed/Catálogo; botones de acción ≥44px (`Añadir al Carrito`, steppers de `Carrito`, `Cerrar Sesión`) — el repo no evidencia este audit específico en código. Revisión visual pendiente.

---

# UNIDAD 2 — BACKEND ALTA TRANSACCIONALIDAD (Go/Fiber · "entornos bancarios")

## U2-T1 · Rendimiento, memoria y concurrencia

> Goroutines y contextos: procesos de misión crítica que no bloquean la respuesta HTTP.

- [x] Goroutine `procesarComisiones` tras el COMMIT del pedido (contexto propio, no el del request): `checkoutController.go:64-72` lanza `go func(total, uID, email)` que invoca `ActualizarComisiones(...)`; no bloquea la respuesta HTTP.

## U2-T2 · Microservicios y APIs robustas (Fiber)

> API REST en capas con contrato de error estándar y configuración por entorno.

- [x] Capas routes → controllers → repository; endpoints `login`, `productos`, `productos/:id`, `red`, `checkout`, `health` registrados en `Backend/routes/routes.go` con DI
- [ ] `APIError{status,message,details}` en todos los errores + validación de entrada (400/404) — `models.APIError` está **definido** pero los controllers usan `fiber.Map{"error": "..."}` plano en login/productos/red/checkout. Login no valida campos vacíos explícitamente.
- [ ] Puerto de escucha configurable (`PORT`, defecto 3000) — **parcial**: `main.go` lee `PORT` con default `"8080"` (no 3000 como sugería el doc original). Funcional, difiere el default.
- [ ] `/api/health` real con ping a Postgres (503 si cae) y Redis (`"degraded"` si cae) — **NO**: `healthController.go` actual devuelve estático `{"status":"ok","message":"Servidor Go/Fiber operativo"}` sin tocar PG ni Redis.

## U2-T3 · Persistencia PostgreSQL + caché Redis

> Dockeriza la BD, SQL puro con pgx, reconstrucción del árbol y Cache-Aside.

- [x] `docker-compose` Postgres 16-alpine + Redis-alpine + volumen `postgres_data` + healthchecks (`pg_isready`, `redis-cli ping`)
- [x] `init.sql` (tablas `usuarios`/`productos`/`referidos`/`pedidos`/`pedido_items` + seed), `repository.go` a SQL real con `pgxpool`, árbol de referidos por `parent_id`
- [x] Cache-Aside funcional: `productos:all` con TTL 300s (5 min) en `ProductoController`. `RedController` usa key `"red:tree"` (sin `:email`) — la letra del doc sugería `red:tree:{email}`. Degradación sin caché correcta (try/recatch).
- [ ] Índices: `referidos(parent_id)`, `pedidos(cliente_email)`, `pedido_items(pedido_id)`, `productos(categoria)` — solo existe `init.sql`; **no** existen las migraciones separadas `002_checkout.sql` y `003_seguridad.sql` que menciona este checklist. Definición de índices por verificar dentro de `init.sql`.
- [x] Errores propagados: firmas del repo devuelven `error`; `ErrNotFound` se retorna en 3 paths del repo. Mapeo en controllers **parcial**: checkout mapea no-nil a 500, login mapea `ErrNotFound` a 401; producto/red no siempre mapean `ErrNotFound` a 404 (devuelven 500).

## U2-T4 · Inyección de dependencias y microservicio seguro

> DI por interfaces en toda la cadena + transacción ACID del checkout + hardening.

- [x] Interfaces `DatabaseRepository`/`CacheRepository` (`repository/interfaces.go`); `PostgresRepository`/`RedisCacheRepository` (`repository/repository.go`); structs de controllers con DI; `main.go` cablea con `NewAuthController`/`NewProductoController`/`NewRedController`/`NewCheckoutController`
- [x] `POST /api/checkout` con transacción ACID en `repository.CrearPedido`, `commit`/`rollback` propagado, total recalculado en servidor desde BD; total recalculado confirmado
- [ ] Invalidación de caché tras compra (`productos:all`, `red:tree:*`) — el controller invalida `red:tree` y `red:tree:{email}` (con email tomado del body del checkout, **no del token**). **NO** invalida `productos:all` tras descuento de stock (el endpoint impacta stock pero la caché no se purge). Migración `002_checkout.sql` **no existe** como archivo separado.
- [ ] bcrypt de passwords — **NO implementado**. Login compara con `WHERE email=$1 AND password=$2` en claro. `init.sql` no usa `crypt()`/`gen_salt('bf')` de pgcrypto. Migración `003_seguridad.sql` **no existe**.
- [ ] JWT real (`JWT_SECRET` en env, HS256 2h, claims `{sub,email,rol,iat,exp}`) + `middleware.RequireAuth(secret)`/`RequireAdmin()` — **NO implementado**. `authController.Login` devuelve `"fake-jwt-token-123"` hardcoded. **No existe** el paquete `Backend/middleware/`. No hay import de librería JWT.
- [ ] Rutas `/api/red` y `/api/checkout` protegidas con middleware (401 sin token / token adulterado) — **NO**. `routes.go` las registra sin wrapper de auth; cualquier petición anónima llega a los controllers.
- [ ] Checkout toma la identidad del TOKEN (`c.Locals` desde claims), `email` falso en body NO se persiste ni atribuye comisiones — **NO**. `checkoutController.Checkout` tiene `usuarioID := 2` hardcoded y llama `dbRepo.CrearPedido(usuarioID, req)`. **Sí** persiste cualquier `emailCliente` del body y **sí** atribuye comisiones a `usuario_id=2` aunque el token sea de otro usuario.
- [ ] Errores sanitizados: `Details` solo con `APP_ENV=dev`; `ErrCantidadInvalida` propio (400 "cantidad debe ser mayor que cero", distinto de stock) — **NO**. `checkoutController` expone `detalle: err.Error()` en producción sin gating por env. No existe `ErrCantidadInvalida` (cantidades ≤ 0 no se validan explícitamente; el controller solo verifica `len(req.Items)==0`).

---
