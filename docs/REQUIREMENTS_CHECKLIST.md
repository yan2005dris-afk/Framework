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
| **Hito actual** | **U1, U2 y U3 completas** (verificación honesta tras merges #10 + #11, marzo-octubre 2026). U1-T1..T5, U2-T1..T4 y U3-T1..T3 cumplidos. Pendientes menores en U1-T2 (renombrado + .env), U1-T5 (audit mobile-first), U2-T2 (`PORT` default 3000 vs 8080; `models.APIError` literal vs funcionalmente equivalente con helpers), U2-T3 (índices `referidos(parent_id)` y `productos(categoria)` y migración `003_seguridad.sql` por verificar/crear). |
| **Última tarea completada** | U3-T3: CRUD `users_saas` con identidad del gateway, aislamiento RLS en Postgres y billing de planes/suscripciones (rama `feature/u3-t3-saas-multitenant`, issues #15, #16, #17). |
| **Siguiente tarea pendiente** | **U4**: #19 build estático SPA + Caddy sirviendo `/`, #20 TLS con Let's Encrypt, #21 rate limit + observabilidad + pipeline de deploy. La PWA (U4-T3) es el bloque grande de esa unidad y ya está planificada en detalle más abajo. Pendientes chicos de U1/U2 siguen en pie (ver commit history). |
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

# UNIDAD 3 — CAPA SaaS MULTI-TENANT + API GATEWAY (FastAPI + Go)

## U3-T1 · Scaffold FastAPI `fastAPI_SaaS/`

> Servicio SaaS/multitenant en :8000, proyecto Python propio con `uv`, health
> endpoint y CORS por env.

- [x] Proyecto autocontenido en `fastAPI_SaaS/` (`pyproject.toml`, `uv.lock`, `alembic/`, `tests/`)
- [x] `GET /api/saas/health` + raíz documentada
- [x] Engine async SQLAlchemy 2 lazy y settings tipados (`SAAS_*`)
- [x] Dependencias completas: `sqlalchemy[asyncio]` (greenlet) y `pydantic[email]` (EmailStr). Sin estos extras el servicio ni importaba

## U3-T2 · API gateway (Caddy) + auth-service

> Gateway como entrada única con `forward_auth`, y la autenticación extraída a
> un microservicio Go que es el único emisor/validador de JWT.

- [x] `auth-service/` en Go/Fiber: `POST /auth/login|logout|refresh|validate`, `GET /auth/me`, health
- [x] Claims `{sub, email, rol, tenant_id, rol_saas}` firmados HS256 con `AUTH_JWT_SECRET`
- [x] Schema `auth.users` (dueño: auth-service) con seeds demo
- [x] `gateway/Caddyfile` con `forward_auth` y `copy_headers X-User-*` hacia los upstreams
- [x] Backend y SaaS dejan de firmar JWT: ambos hacen proxy al auth-service
- [x] `tests/e2e/gateway_e2e.sh`: health, catálogo, 401 sin cookie, login, `me`, proxy SaaS
- [x] Usuarios unificados en `auth.users` (`004_migrate_to_auth_schema.sql`): sin `public.usuarios`

## U3-T3 · Multi-tenant funcional: usuarios, aislamiento y billing

> Issues #15, #17 y #16. La identidad llega inyectada por el gateway; el
> aislamiento se sostiene en tres capas: filtro por `tenant_id` en cada query,
> guards de scope/rol en cada endpoint, y Row-Level Security en Postgres.

- [x] Identidad real desde los headers `X-User-*` (`app/deps/principal.py`), sin tenant ni rol por defecto
- [x] `401` sin identidad, `403` fuera de scope o sin rol admin, `404` tenant/usuario inexistente
- [x] CRUD `users_saas` (`POST/GET/PATCH/DELETE /api/saas/tenants/{tenant_id}/users`) con email único por tenant
- [x] Passwords con bcrypt (SHA-256 + base64 previo para no cortar a 72 bytes); el hash nunca sale por la API
- [x] Transacción por request (`get_tenant_db`): un solo commit, rollback ante error, `SET LOCAL app.current_tenant`
- [x] RLS en `tenants`, `users_saas` y `subscriptions` con `FORCE ROW LEVEL SECURITY` (migración `0003_rls_policies`)
- [x] Planes y suscripciones (`0002_billing`) + catálogo semilla free/pro/enterprise
- [x] Webhook de Stripe como stub explícito (no verifica firma todavía)
- [x] Tests: 54 unitarios/API en sqlite + 8 de integración RLS contra Postgres real


# UNIDAD 4 — CONTAINERIZACIÓN + PWA + PRODUCCIÓN

## U4-T1 · Build SPA + Caddy sirviendo estáticos

> Compilación de producción del frontend y configuración de Caddy como servidor
> de archivos estáticos con reverse proxy a servicios backend.

- [ ] Build de producción: `npm run build` genera `Frontend/dist/` optimizado
- [ ] Caddyfile configurado para servir `/` desde `Frontend/dist/` y `/api/*` como reverse proxy
- [ ] Imagen Docker multistage: stage build → stage runtime con Caddy
- [ ] `docker compose` actualizado para usar la nueva imagen de frontend + Caddy

## U4-T2 · TLS con Let's Encrypt (HTTPS obligatorio para PWA)

> Certificados SSL automáticos con Caddy's auto_https en producción.

- [ ] Configuración `auto_https` en Caddyfile para dominio real
- [ ] Certificados Let's Encrypt generados automáticamente en producción
- [ ] HTTP → HTTPS redirect forzado
- [ ] Testing de HTTPS en staging (dominio de prueba)

## U4-T3 · PWA: Progressive Web App

> Transformación de la SPA en PWA instalable: manifest, service worker,
> offline capability, meta tags y Lighthouse PWA criteria.

### Fase 1: Configuración base y manifest
- [ ] Instalar `vite-plugin-pwa` (o `@vite-pwa/assets-generator`): `npm install -D vite-plugin-pwa workbox-window`
- [ ] Crear `Frontend/public/manifest.json` con:
  - `name`: "MultiCatálogo"
  - `short_name`: "MultiCatálogo"
  - `description`: "Tienda de cosméticos con red de ventas multinivel"
  - `start_url`: "/"
  - `display`: "standalone"
  - `background_color`: "#ffffff"
  - `theme_color`: "#ffffff"
  - `orientation`: "portrait-primary"
  - `icons`: [192x192, 512x512] (generar vía @vite-pwa/assets-generator)
- [ ] Actualizar `Frontend/index.html` con meta tags PWA:
  - `<meta name="theme-color" content="#ffffff">`
  - `<link rel="manifest" href="/manifest.json">`
  - `<link rel="apple-touch-icon" href="/icon-192.png">`
- [ ] Configurar `vite.config.ts` con plugin PWA:
  ```ts
  import { VitePWA } from 'vite-plugin-pwa'
  plugins: [react(), tailwindcss(), VitePWA({...})]
  ```

### Fase 2: Service Worker y estrategias de caché
- [ ] Configurar Workbox strategies en `vite.config.ts`:
  - `NetworkFirst` para `/api/*` (peticiones a backend)
  - `CacheFirst` para assets estáticos (JS, CSS, imágenes)
  - `StaleWhileRevalidate` para el HTML index
- [ ] Service worker registrado en `Frontend/src/main.tsx` con `registerSW()`
- [ ] UI de actualización disponible: mostrar banner cuando el SW tiene nueva versión
- [ ] Testing de SW: devtools → Application → Service Workers

### Fase 3: Offline capability y UX
- [ ] Página offline: `Frontend/src/pages/Offline.tsx` con mensaje amigable y botón "reintentar"
- [ ] Service worker intercepta offline requests y sirve página offline
- [ ] Indicador de conexión: mostrar estado online/offline en UI (hook `useOnlineStatus`)
- [ ] Cola de peticiones fallidas: reintentar automáticamente cuando vuelve la conexión (para checkout/pedidos)

### Fase 4: Lighthouse PWA audit
- [ ] Ejecutar Lighthouse audit en Chrome DevTools (categoría PWA)
- [ ] Verificar score ≥90 en todos los criterios PWA:
  - [ ] PWA Optimized (manifest correcto)
  - [ ] Installable (iconos, start_url, display)
  - [ ] Service Worker registrado y activo
  - [ ] Works Offline (caché de assets + página offline)
  - [ ] HTTPS (requisito de U4-T2)
  - [ ] Responsive (viewport meta tag)
  - [ ] Theme color definido
- [ ] Corregir fallos hasta obtener Lighthouse PWA score ≥90

### Fase 5: Instalación y UX de app instalada
- [ ] Botón "Instalar App" (detectar `beforeinstallprompt` event)
- [ ] UX de onboarding post-instalación (primer lanzamiento como app)
- [ ] Ícono de launcher generados (192x192, 512x512)
- [ ] Splash screen configurado (meta tags para iOS/Android)
- [ ] Testing en dispositivos reales (Android Chrome, iOS Safari)

## U4-T4 · Rate limiting + observabilidad + deploy pipeline

> Protección contra abuso, logs estructurados, métricas y CI/CD pipeline.

- [ ] Rate limiting en Caddy (requests/min por IP)
- [ ] Logs estructurados JSON en todos los servicios
- [ ] Métricas Prometheus + Grafana básicos
- [ ] Pipeline CI/CD: tests → build → docker push → deploy

---
