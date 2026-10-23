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
  tags. Route: inline. Commit: `925c48f`.
- T2 (S5-S6) `registerSW` en `main.tsx` y banner de actualizacion con `sonner`,
  que ya esta en el proyecto. Route: inline. Commit: `925c48f`.
- T3 (S4) `runtimeCaching` de Workbox con las estrategias de S4. Route: inline.
  Commit: `925c48f`.
- T4 (S7-S8) Ruta `/offline`, componente `Offline` con boton de reintentar, y
  hook `useOnlineStatus`. Route: inline. Commit: `925c48f`.
- T5 (S10) Migracion `006_idempotencia.sql` y soporte de la key en
  `CrearPedido` + test que demuestra que dos POST con la misma key dejan un solo
  pedido. Route: inline. Commit: `c082711`.
- T6 (S9) Cliente HTTP con reintento al volver la conexion, generando y
  reutilizando la `Idempotency-Key` del mismo intento de checkout. Route: inline.
  Commit: `3ff3325`.
- T7 (S11) Verificar con chromium que la app es instalable y funciona offline,
  y actualizar `docs/REQUIREMENTS_CHECKLIST.md`. Route: inline.

Commits fuera de la lista de tareas, encontrados durante T5:

- `2b5343d` fix P0 del gateway (ver L7). Va antes que la idempotencia a proposito:
  sin el, el checkout por el proxy no funcionaba y no habia nada que reintentar.

Los cuatro de la Fase 1 a 4 van en un solo commit (`925c48f`) porque separarlos
deja la app en un estado que no se puede instalar ni probar.

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
L7. **Bug preexistente del gateway, encontrado al probar T5**: todo POST autenticado con body a
    traves del gateway devolvia 502. No era culpa de la idempotencia ni del rate_limit: el
    `forward_auth` reenvia el `Content-Length` del request original al auth-service, que
    respondia sin consumir ese body; el servidor HTTP cortaba la conexion a medias, Caddy
    marcaba el transporte como roto y el body quedaba vacio para el `reverse_proxy` que
    venia despues (`http: ContentLength=159 with Body length 0`). O sea, el checkout nunca
    funciono por el proxy. Fix: `_ = c.Body()` al inicio de `AuthHandler.Validate`. El test
    de integracion `TestCheckoutIdempotencyMismoKeyUnSoloPedido` es el test de regresion: da
    502 sin el fix y 201/200 con el.
L8. T3 fijo un bug de performance que no era de la PWA pero la volvia inviable: `Icon.tsx`
    usaba `DynamicIcon` de `lucide-react/dynamic` y arrastraba ~1900 iconos como chunks
    sueltos. El service worker los precacheaba todos: 1938 entradas. Con el mapa estatico
    quedaron 34.
L9. La eleccion de reintento con idempotency key (L2) obligo a tocar el Backend, y al hacerlo
    aparecio el bug de L7. Sin el fix del gateway, S9 y S10 habrian quedado "verificados" contra
    una ruta que en produccion devolvia 502: los tests de integracion se corrieron contra el
    gateway justamente para no cair en eso.
L10. **Regresion mia, encontrada al intentar correr Lighthouse**: la app no renderizaba, el
    `<body>` llegaba vacio y la consola tiraba `Cannot read properties of null (reading 'stores')`.
    La causa es `OfflineGuard`, que se monta en `AppShell` al lado del `RouterProvider` pero FUERA
    de el, y usaba `useRouter()`/`useRouterState()`: esos hooks leen el contexto del router, que
    en ese punto todavia no existe. `pnpm lint` y `pnpm build` pasan igual, asi que el tipo no lo
    detecta: el fallo es de orden de montaje.
    Se aislo compilando el commit previo (b0c646b) en un worktree: ese renderiza bien, asi que
    la regresion es de U4-T3 y no del shell. El fix es usar el `router` como singleton en vez de
    los hooks.
    Dos lecciones: (1) `pnpm build` en verde NO significa que la app arranque, hace falta abrirla
    en un navegador; (2) el audit de Lighthouse es lo que lo delato, porque NO_FCP es
    precisamente "la pagina no pinto nada".
L11. Resultado de la verificacion (S11) sobre el stack real en `http://localhost:8080`:
    Lighthouse **accessibility 96, best-practices 96, performance 81** (FCP 3,4 s). La categoria
    PWA ya no existe en Lighthouse (se elimino; en la 13.5.0 quedan accessibility,
    best-practices, performance, seo y agentic-browsing), asi que el criterio de "score >=90 en
    PWA" del checklist no se puede cumplir tal como esta escrito: hay que reformularlo.
    Verificado por HTTP: `/manifest.json` con los campos del checklist, `/sw.js` (200,
    text/javascript) y los 3 iconos PNG. Queda pendiente a ojo en DevTools: que el SW quede
    registrado y que la app abra sin red.