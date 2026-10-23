# Frontend

SPA de MultiCatalogo: React 19 + Vite 8 + Tailwind 4 + TanStack Router. Habla con la API por `/api` (ruta relativa), nunca por URL absoluta, para que el browser no lo vea como otro origen.

## Como se ejecuta

| Modo | Puerto | Quien lo sirve | Cuando |
|---|---|---|---|
| dev en el host | 5173 | `pnpm dev` (Vite + HMR) | iteracion rapida sin docker |
| dev en compose | 5173 | servicio `frontend` (profile `dev`) | HMR sin pnpm en el host |
| produccion | 8080 | gateway (Caddy), junto a la API | `docker compose up -d` |

## Desarrollo en el host

```bash
pnpm install
pnpm dev          # http://localhost:5173
```

El dev server proxea `/api` al gateway (`:8080` por defecto). El target sale de la variable de entorno **`VITE_DEV_API_TARGET`**:

| Contexto | Valor | Donde se define |
|---|---|---|
| host, gateway en compose | *(sin nada: fallback `http://localhost:8080`)* | default de `vite.config.ts` |
| servicio `frontend` del compose | `http://gateway:8080` | `environment:` del servicio |
| Backend nativo sin gateway | `http://localhost:3000` | `VITE_DEV_API_TARGET=http://localhost:3000 pnpm dev` |

El contrato se lee con `loadEnv`, que suma al `.env` del proyecto lo que venga en `process.env` con prefijo `VITE_`. Por eso el mismo `vite.config.ts` sirve en el host y dentro del contenedor.

## Desarrollo en compose

```bash
docker compose --profile dev up -d frontend    # http://localhost:5173
```

El servicio monta `./Frontend` sobre `/app` para que HMR recompile al guardar, y conserva `node_modules` en un volumen con nombre. `docker compose up -d` a secas **no** lo levanta: en ese caso el frontend lo sirve el gateway en `:8080`.

Si cambian dependencias: `docker compose --profile dev build frontend` y despues `up -d` de nuevo.

## Build de produccion

```bash
pnpm build     # tsr generate && tsc -b && vite build  ->  dist/
pnpm preview   # sirve dist/ localmente
```

El build de produccion **no** pasa por el servicio `frontend`: lo compila el stage `frontend` de `gateway/Dockerfile` y la imagen final de Caddy sirve ese `dist/` en `/srv/frontend`. Ver `gateway/README.md`.

Notas del build:

- `pnpm build` arranca con `tsr generate`, que crea `src/routeTree.gen.ts` a partir de `src/routes/`. El archivo esta en `.gitignore` y en el `.dockerignore` de la raiz: se genera siempre, nunca se copia del host.
- `src/routeTree.gen.ts` se genera antes de `tsc -b` a proposito; el typecheck depende de el.
- No hay que inyectar nada de la API en build: `VITE_API_URL` tiene fallback en `src/services/http.ts` (`DEFAULT_API_BASE_URL = "/api"`).

## Otros comandos

```bash
pnpm typecheck   # tsr generate && tsc -b
pnpm lint        # eslint
```
