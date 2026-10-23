# Dockerfile del gateway (Caddy 2.x) con el build del SPA dentro de la imagen.
#
# Tres stages:
#   1. `frontend`: node + pnpm compila Frontend/ y produce `dist/`.
#   2. `caddy`   : xcaddy compila Caddy con el plugin de rate limit.
#   3. runtime   : imagen final con ese binario, ese `dist/` en /srv/frontend
#                  (la raiz que declara el Caddyfile) y el Caddyfile.
#
# El resultado es que el unico puerto publico del stack es :8080: el SPA y la
# API comparten origen, que es lo que hace funcionar la cookie httpOnly
# `auth_token` sin CORS.
#
# El stage de build necesita Frontend/ y pnpm-lock.yaml, que viven FUERA de
# este directorio. Por eso docker-compose.yml usa `context: .` (raiz del repo)
# con `dockerfile: gateway/Dockerfile`, y por eso la raiz tiene .dockerignore.

# --- Stage 1: build del SPA ----------------------------------------------------
# Node 22 + pnpm 9 son los que fija .github/workflows/ci.yml. bookworm (glibc)
# y no alpine a proposito: Tailwind 4 y Rolldown bajan binarios nativos por
# plataforma y glibc es el mismo que ve CI.
FROM node:22-bookworm-slim AS frontend

# package.json no declara `packageManager`, asi que el default de corepack
# cambiaria con el tiempo. Fijamos pnpm aca para que el build sea reproducible.
ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
RUN corepack enable && corepack prepare pnpm@9 --activate

WORKDIR /build/Frontend

# Solo los manifiestos primero: la capa de `pnpm install` se reusa aunque cambie
# todo el codigo. Notese que NO se copia src/routeTree.gen.ts (esta en
# .gitignore y el .dockerignore de la raiz lo excluye): lo crea `tsr generate`
# dentro de `pnpm build`, asi que la imagen nunca depende del estado del host.
COPY Frontend/package.json Frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY Frontend/ ./
RUN pnpm build

# --- Stage 2: build de Caddy con modulos --------------------------------------
# El rate limiting (#21) es un plugin, no viene en la imagen oficial de Caddy.
# `caddy:2-builder` trae xcaddy y las toolchains para compilarlo. El binario
# resultante reemplaza al de la imagen base en el stage final.
#
# La version del plugin va fija: sin pin, un `docker compose build` de mañana
# podria traer otro codigo y el rate limit cambiaria sin que nadie lo note.
FROM caddy:2-builder AS caddy

RUN xcaddy build \
    --with github.com/mholt/caddy-ratelimit@v0.1.0

# --- Stage 3: runtime con Caddy ------------------------------------------------
FROM caddy:2

# Caddy con el plugin de rate limit compilado.
COPY --from=caddy /usr/bin/caddy /usr/bin/caddy

# Copiamos el Caddyfile en la ruta default que Caddy lee si existe.
# Tambien soportamos /etc/caddy/Caddyfile para volumenes custom.
# Contexto de build = raiz del repo (ver docker-compose.yml), asi que el
# Caddyfile se copia con su ruta real desde la raiz.
COPY gateway/Caddyfile /etc/caddy/Caddyfile

# El SPA compilado en el stage anterior. `root * /srv/frontend` + `try_files`
# en el Caddyfile sirven `/` y hacen fallback a index.html para las rutas de
# cliente (/catalogo, /mi-red, /producto/:id).
COPY --from=frontend /build/Frontend/dist /srv/frontend

# Por defecto Caddy escucha en :2019 (HTTP) y :443 (HTTPS). Cambiamos a 8080
# explicito via env. Si QUEREMOS que el Dockerfile arranque Caddy solo,
# podriamos sobreescribir CMD, pero por defecto heredado es OK con
# 'caddy run --config /etc/caddy/Caddyfile' desde el entrypoint base.
ENV CADDY_PORT=8080

EXPOSE 8080

# Healthcheck: /health es un endpoint dedicado del Caddyfile que responde 200
# sin pasar por forward_auth. Antes apuntaba a /, que con forward_auth sin
# matcher devolvia 401 y dejaba el contenedor unhealthy para siempre.
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]
