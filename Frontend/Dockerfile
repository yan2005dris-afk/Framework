# Imagen de DESARROLLO del frontend (Vite con HMR), no de produccion.
#
# El build de produccion del SPA lo compila el stage `frontend` de
# gateway/Dockerfile y el gateway lo sirve en :8080. Esta imagen existe solo
# para el servicio `frontend` de docker-compose.yml (profile `dev`), que
# monta el codigo del host sobre /app para recargar en caliente. El codigo NO
# se copia a la imagen: va siempre en bind mount.
#
# El contexto de build es la raiz del repo (lo define docker-compose.yml), por
# eso los paths de este archivo arrancan en `Frontend/`. Asi el unico
# .dockerignore que hace falta es el de la raiz.

FROM node:22-bookworm-slim

# Misma version de node + pnpm que CI (.github/workflows/ci.yml) y que el
# stage de build de gateway/Dockerfile. package.json no declara
# `packageManager`, asi que fijamos pnpm aca con corepack.
ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
RUN corepack enable && corepack prepare pnpm@9 --activate

WORKDIR /app

# Las deps se instalan al construir la imagen, no en cada arranque: el compose
# las preserva en un volumen con nombre montado sobre /app/node_modules, que
# tapa el node_modules del host. Para actualizarlas hay que rebuildear
# (`docker compose --profile dev build frontend`).
COPY Frontend/package.json Frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

EXPOSE 5173

# Sin flags: `server.host: true` en vite.config.ts es lo que hace que Vite
# escuche en 0.0.0.0 y el host llegue a :5173. Pasarlo por CLI con
# `pnpm run dev -- --host` no sirve: pnpm reenvia el separador y Vite corre
# como `vite -- --host`, con `--host` tratado como posicional.
CMD ["pnpm", "run", "dev"]
