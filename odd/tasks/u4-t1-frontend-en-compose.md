Feature: U4-T1 — el frontend entra al compose de la raiz. Hoy el SPA corre solo en
el host con `pnpm dev` y el compose tiene 6 servicios sin frontend; el Caddyfile ya
tiene escrito el placeholder `root * /srv/frontend` + `file_server`, o sea que la
infraestructura espera que el gateway sirva el build. Esta feature cierra esa pieza:
build de produccion del SPA dentro de la imagen del gateway (todo en :8080, SPA y API
en el mismo origen) mas un servicio de desarrollo con HMR para no obligar a rebuild
de imagen en cada cambio.

## Specs

S1. "creo yo que en el compose que tenemos en la raiz deberia ir tambien el frotnend"

S2. Decision: "Build dentro de la imagen del gateway". El build del SPA ocurre en una
stage de build y el `dist/` queda copiado en la imagen final de Caddy, que sirve `/`
y proxea `/api/*`. Con eso el unico puerto publico es :8080 y SPA + API comparten
origen (que es lo que hace funcionar la cookie `auth_token` sin CORS).

S3. Decision: "En un servicio del compose". Durante el desarrollo el frontend corre
como servicio con Vite en modo dev, con el codigo del host montado y HMR.

S4. Alcance ya especificado por el proyecto (docs/REQUIREMENTS_CHECKLIST.md, U4-T1):
    - "Build de producción: `npm run build` genera `Frontend/dist/` optimizado"
    - "Caddyfile configurado para servir `/` desde `Frontend/dist/` y `/api/*` como reverse proxy"
    - "Imagen Docker multistage: stage build → stage runtime con Caddy"
    - "`docker compose` actualizado para usar la nueva imagen de frontend + Caddy"

S5. El target del proxy de Vite tiene que ser configurable por entorno: en el host es
`http://localhost:8080`, dentro de la red de compose es `http://gateway:8080`.

S6. La SPA servida por Caddy es una app con rutas de cliente (`/catalogo`, `/mi-red`,
    `/producto/:id`). Un refresh o un deep link tienen que devolver el `index.html`, no
    404.

S7. Limpieza: `Frontend/@/` es un directorio literal que creo la CLI de shadcn cuando
    no resolvio el alias `@/` (los archivos se movieron, las carpetas quedaron vacias).
    No debe viajar al build.

## Tasks

- T1 (S1, S4) `gateway/Dockerfile` multi-stage: stage `frontend` con node+pnpm que
  corre `pnpm build`, stage final `caddy:2` con el `dist/` en `/srv/frontend`.
  Route: delegated writer. Commit: work-unit 1.
- T2 (S4) Caddyfile: actualizar el bloque de estaticos (ya no es placeholder) y los
  comentarios que dicen que el volumen no se monta. Route: delegated writer.
  Commit: work-unit 1.
- T3 (S1, S3) Servicio `frontend` de desarrollo en `docker-compose.yml`: Vite dev con
  `host: true`, codigo montado, puerto publicado, `depends_on: gateway`. Route:
  delegated writer. Commit: work-unit 2.
- T4 (S5) `Frontend/vite.config.ts`: target del proxy desde entorno con fallback a
  `http://localhost:8080`. Route: delegated writer. Commit: work-unit 2.
- T5 (S7) Eliminar `Frontend/@/` si queda vacío. Route: inline. Commit: work-unit 2.
- T6 (S6, S1-S5) Verificar de punta a punta: build de la imagen del gateway, `:8080`
  devuelve el SPA, `/api/login` y `/api/productos` funcionan por el mismo origen, deep
  link a `/catalogo` devuelve el index, y el servicio dev sirve el SPA con HMR.
  Route: inline (parent). Commit: ninguno.
- T7 Documentacion: cabecera de `docker-compose.yml` (dice 6 servicios), README del
  gateway, `docs/PROJECT_MAP.md`. Route: delegated writer. Commit: work-unit 2.

## Log

L1. "creo yo que en el compose que tenemos en la raiz deberia ir tambien el frotnend"
    (user)
L2. Decision (user): "Build dentro de la imagen del gateway" frente a las opciones de
    servicio con volumen o contenedor de desarrollo.
L3. Decision (user): "En un servicio del compose" para el flujo de desarrollo.
L4. Contexto: es U4-T1 del checklist (issue #19 abierto), no trabajo nuevo de scope.
    U4-T2 (TLS) y U4-T3 (PWA) quedan fuera.
L5. Previo en esta rama: `f31eee7` dejo la base de API relativa (`/api`), que es lo que
    hace posible que SPA y API compartan origen en :8080.
L6. Implementacion delegada a `gentle-ai-worker` con superficies acotadas (gateway,
    compose, `.dockerignore`, `Frontend/vite.config.ts`, `Frontend/Dockerfile`, docs).
    Reporto S1-S7 cubiertos. Corregi un bug que el writer dejo: al mover el contexto de
    build del gateway a la raiz, `COPY Caddyfile /etc/caddy/Caddyfile` quedo apuntando a
    un path inexistente (en la raiz no hay Caddyfile). Corregido a
    `COPY gateway/Caddyfile /etc/caddy/Caddyfile`.
L7. Supuestos del writer resueltos por el padre: el `.env` de la raiz NO define
    `VITE_DEV_API_TARGET` ni `FRONTEND_PORT` (verificado por grep), asi que no hay riesgo
    de que el contenedor del frontend se proxyee a si mismo; `pnpm-lock.yaml` es
    `lockfileVersion: '9.0'`, compatible con el pin a pnpm@9 que impone
    `.github/workflows/ci.yml` (node 22 + pnpm 9).
L8. T5 quedo sin objeto: `Frontend/@/` ya no existe (limpieza del padre).
L9. Segundo bug del writer, encontrado al verificar: `CMD ["pnpm","run","dev","--","--host"]`
    en Frontend/Dockerfile no aplica el flag. pnpm reenvia el separador y Vite corre
    `vite -- --host`, con `--host` como posicional: el dev server queda en loopback y el
    host no llega a :5173 (logs: "Network: use --host to expose"). Corregido dejando
    `server.host: true` en vite.config.ts (un solo lugar) y el CMD sin flags.
L10. Verificacion con el stack real (DB_PORT=5433, porque 5432 esta ocupado por otro
    proyecto):
    - Gateway (:8080): `/` 200 con assets hasheados; `/catalogo` y `/producto/3` 200 con
      index.html (`try_files`); `/assets/index-*.js` 200 `text/javascript` servido de
      disco; `POST /api/login` 200 + cookie; `/api/me` con cookie 200 y sin cookie 401;
      `/api/red` 200 con el arbol; `/api/productos` 200 con 8 productos. Todo SPA + API en
      el mismo origen.
    - Servicio dev (profile `dev`, :5173): SPA 200 con `@vite/client`; `/catalogo` 200;
      `/api/productos` 200 con 8 productos a traves del proxy; login + cookie + `/api/me`
      200 por el proxy; cliente de HMR presente; `VITE_DEV_API_TARGET=http://gateway:8080`
      dentro del contenedor.
    - `tsc -b` y `eslint .` limpios.
L11. Decision de entrega: un solo commit, no dos. Las dos mitades (build del SPA en la
    imagen del gateway, servicio dev) son interdependientes: partir el `docker-compose.yml`
    entre dos commits deja commits intermedios incoherentes. El diff son469 lineas
    insertadas, ~300 de ellas codigo.