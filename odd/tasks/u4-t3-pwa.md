# U4-T3: PWA instalable, offline y reintento idempotente

Objetivo: la ultima task abierta de la Fase 4. Transformar el SPA en una PWA
instalable, con service worker, offline degradado y reintento seguro de pedidos.
Rama: `feat/u4-t3-pwa` (base `develop`, que ya tiene el hardening de U4).
Runner: `pnpm lint`, `pnpm build`, `go test ./...`, tests de integracion, y
verificacion de instalabilidad con chromium.

## Specs

S1. La app es instalable: manifest con `name`: "MultiCatálogo",
`short_name`: "MultiCatálogo", `description`: "Tienda de cosméticos con red de
ventas multinivel", `start_url`: "/", `display`: "standalone",
`background_color`: "#ffffff", `theme_color`: "#ffffff",
`orientation`: "portrait-primary", e `icons` de 192x192 y 512x512.
(checklist U4-T3, Fase 1)

S2. `Frontend/index.html` lleva los meta tags PWA:
`<meta name="theme-color" content="#ffffff">`,
`<link rel="manifest" href="/manifest.json">` y
`<link rel="apple-touch-icon" href="/icon-192.png">`. (checklist, Fase 1)

S3. `vite.config.ts` incluye el plugin `vite-plugin-pwa`. (checklist, Fase 1)

S4. Estrategias de cache Workbox. **Decision del usuario**: cachear SOLO el
catalogo publico, que es lo unico igual para todos. `NetworkFirst` para
`/api/productos`, `StaleWhileRevalidate` para `/api/productos/:id`, y
`NetworkOnly` (o sin regla, que es lo mismo) para `/api/login`,
`/api/checkout`, `/api/me` y `/api/red`. El checklist pedia `NetworkFirst` para
todo `/api/*`; se deviate a proposito porque la API usa cookie httpOnly y un
GET cacheado puede mostrar datos de la sesion anterior. (checklist, Fase 2 +
decision del usuario)

S5. El service worker se registra en `Frontend/src/main.tsx` con `registerSW()`.
(checklist, Fase 2)

S6. Hay UI de actualizacion: se muestra un banner cuando el service worker tiene
una version nueva. (checklist, Fase 2)

S7. Pagina offline con mensaje amigable y boton de reintentar, a la que el
service worker manda cuando no hay red. (checklist, Fase 3)

S8. Indicador de conexion online/offline en la UI, via un hook
`useOnlineStatus`. (checklist, Fase 3)

S9. Las peticiones fallidas por red se reintentan solas cuando vuelve la
conexion, **con idempotency key**. **Decision del usuario**: "Reintento
automatico, con idempotency key". El caso que motiva esto: el primer intento
llego al servidor y solo se perdio la respuesta; sin idempotencia el reintento
duplica el pedido. (checklist, Fase 3 + decision del usuario)

S10. El Backend acepta una `Idempotency-Key` en `POST /api/checkout`: si la key
ya se uso, devuelve el pedido ya creado en vez de crear uno nuevo. Persiste la
key en Postgres con restriccion UNIQUE, dentro de la misma transaccion del
pedido. Sin esto, S9 no tiene garantia y el reintento duplica pedidos.
(derivado de S9)

S11. Verificacion de instalabilidad: los criterios PWA de Lighthouse se
cumplen (manifest correcto, instalable, service worker activo, funciona
offline, HTTPS). En dev `localhost` ya es secure context, asi que el service
worker corre sin depender de U4-T2. (checklist, Fase 4)

## Tasks

- T1 (S1-S2-S3) Instalar `vite-plugin-pwa` y `workbox-window`, generar los
  iconos 192 y 512 desde `public/favicon.svg`, escribir el manifest y los meta
  tags. Route: inline.
- T2 (S5-S6) `registerSW` en `main.tsx` y banner de actualizacion con `sonner`,
  que ya esta en el proyecto. Route: inline.
- T3 (S4) `runtimeCaching` de Workbox con las estrategias de S4. Route: inline.
- T4 (S7-S8) Ruta `/offline`, componente `Offline` con boton de reintentar, y
  hook `useOnlineStatus`. Route: inline.
- T5 (S10) Migracion `006_idempotencia.sql` y soporte de la key en
  `CrearPedido` + test que demuestra que dos POST con la misma key dejan un solo
  pedido. Route: inline.
- T6 (S9) Cliente HTTP con reintento al volver la conexion, generando y
  reutilizando la `Idempotency-Key` del mismo intento de checkout. Route: inline.
- T7 (S11) Verificar con chromium que la app es instalable y funciona offline,
  y actualizar `docs/REQUIREMENTS_CHECKLIST.md`. Route: inline.

## Delivery strategy

`ask-on-risk`. T5 toca el Backend y crea pedidos, asi que va en su propio
commit y con test propio. T6 depende de T5 (usa la key que T5 define), asi que
no son paralelizables aunque los archivos no se toquen entre si. T1-T4 van en
un commit: son el mismo comportamiento (la PWA) y separarlos deja la app en un
estado que no se puede instalar.

## Log

L1. Pedido del usuario (verbatim): "ok dale sigue con el u4t3"
L2. Decisiones del usuario sobre las tres preguntas de diseño (verbatim):
    - cache: "Solo catálogo público (recomendado)"
    - reintentos: "Reintento automático, con idempotency key"
    - alcance: "Fases 1 y 2 + offline UX (recomendado)"
L3. Verificado antes de especificar: `vite-plugin-pwa@2.0.0` declara
    `vite: ^3 || ^4 || ^5 || ^6 || ^7 || ^8`, compatible con el Vite 8.3 del
    repo. Pide `workbox-build`, `workbox-window` y `@vite-pwa/assets-generator`
    como peers.
L4. El checklist (U4-T3, Fase 2) pide `NetworkFirst` para todo `/api/*`. Se
    desvia a proposito, ver S4. La razon es que `/api/me` y `/api/red` devuelven
    datos del usuario autenticado por cookie httpOnly: cachearlos puede
    mostrarlos despues de un cambio de sesion en el mismo dispositivo.
L5. La opcion de reintento automatico que eligio el usuario arrastra una parte
    de Backend que no estaba en el checklist (S10). Sin ella, el reintento es
    una bomba de duplicados: el escenario es el request que llego, perdio la
    respuesta y se reintenta.
L6. `crearPedido` del Backend ya es una transaccion ACID con rollback
    (`Backend/repository/repository.go:199`). La idempotencia entra ahi adentro,
    no alrededor: es lo que hace que dos POST concurrentes con la misma key no
    puedan crear dos pedidos.