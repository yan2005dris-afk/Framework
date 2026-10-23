# Operacion del stack

Como levantar, mirar y desplegar MultiCatalogo. Para el diseño del sistema ver
`docs/architecture.md`; para el estado de features, `docs/REQUIREMENTS_CHECKLIST.md`.

## 1. Levantar todo

```bash
cp .env.example .env    # la primera vez
docker compose up -d
```

Eso levanta los 6 servicios y deja la aplicacion completa en
<http://localhost:8080>: el SPA compilado dentro de la imagen del gateway y la
API en el mismo origen. No hay frontend que levantar aparte.

Para reconstruir las imagenes despues de tocar codigo:

```bash
docker compose up -d --build
```

## 2. Servicios y puertos

| Servicio | Puerto host | Puerto interno | Que hace |
|---|---|---|---|
| `gateway` | 8080 | 8080 | Caddy: sirve el SPA y proxea `/api/*` |
| `backend` | 3000 | 3000 | API B2C (catalogo, checkout, MLM) en Go/Fiber |
| `saas-api` | 8000 | 8000 | API SaaS multi-tenant en FastAPI |
| `auth-service` | 3001 | 3001 | Emision y validacion de JWT |
| `db` | 5432 | 5432 | Postgres 16, compartido |
| `redis` | 6379 | 6379 | Cache del Backend |
| `frontend` | 5173 | 5173 | Solo dev (profile `dev`): Vite con HMR |

Los puertos del host son configurables por `.env`: `GATEWAY_PORT`,
`BACKEND_PORT`, `SAAS_API_PORT`, `AUTH_SERVICE_PORT`, `DB_PORT`, `REDIS_PORT`,
`FRONTEND_PORT`.

## 3. Orden de arranque

El orden lo impone `depends_on` con `condition: service_healthy`, no el orden
de escritura del archivo:

```
db -> redis -> auth-service -> backend -> gateway
                saas-api ----^
```

- `db` y `redis` son la base: todo lo demas espera a que su healthcheck pase.
- `auth-service` espera a `db` porque su entrypoint crea el schema `auth`.
- `backend` espera a `auth-service` ademas de a `db` y `redis`: su entrypoint
  aplica las migraciones, y la migracion 004 re-apunta una FK a `auth.users`.
  Sin esa dependencia el Backend arranca antes y muere con un error confuso.
- `gateway` espera a los tres upstreams: sin eso sirve un 502 en el primer
  request.

Comprobacion de estado:

```bash
docker compose ps          # columna Health de cada servicio
docker compose logs -f gateway
```

## 4. Desarrollo con HMR

El build de produccion del SPA lo compila el gateway, asi que para iterar sobre
el frontend conviene el dev server:

```bash
docker compose --profile dev up -d frontend   # http://localhost:5173
```

El dev server proxea `/api` al gateway, de modo que la cookie `auth_token` sigue
viajando en el mismo origen y no hace falta CORS.

Para iterar sobre el backend sin reconstruir imagenes, se pueden correr los
servicios de Go nativos y dejar solo la infra en Docker:

```bash
docker compose up -d db redis auth-service saas-api
cd Backend && go run main.go
cd fastAPI_SaaS && uv run uvicorn app.main:app --host 0.0.0.0 --port 8000
```

## 5. TLS (#20)

La direccion del sitio sale del entorno, nunca del Caddyfile. Hay tres modos y
solo se cambia el `.env`.

**Desarrollo (default).** Sin `SITE_ADDRESS`, el gateway escucha en
`http://localhost:8080` y no hay TLS.

**TLS local, sin dominio publico.** Con el dominio apuntando a `localhost` en
`/etc/hosts`:

```bash
SITE_ADDRESS=https://localhost:8443
```

Caddy emite un certificado de su CA interna. El browser avisa que no la
confia: se acepta una vez por maquina. Sirve para probar cookies `Secure` y
redirecciones sin depender de un dominio real.

**Produccion, con Let's Encrypt.** Requiere que el DNS del dominio apunte al host
y que los puertos 80 y 443 esten abiertos:

```bash
SITE_ADDRESS=api.tudominio.com
ACME_EMAIL=vos@tudominio.com
```

Caddy pide el certificado por HTTP-01 y lo renueva solo. Para exponer los
puertos hay que descomentar las dos lineas de `ports:` del servicio `gateway`
en `docker-compose.yml`.

> `ACME_EMAIL` no es opcional en un deploy real: sin ella el gateway emite un
> certificado de CA interna que ningun browser confia. El sitio no da error, no
> abre.

Las variables van en `.env` **sin dejar la linea en blanco**. Una variable
definida y vacia hace que el gateway no arranque: comentala en vez de vaciarla.

Los certificados viven en los volumenes `caddy_data` y `caddy_config`. Sin
ellos, cada reinicio del gateway pide un certificado nuevo y Let's Encrypt
corta el acceso a los 5 por semana que permite por dominio.

## 6. Rate limiting (#21)

El gateway limita por IP antes de tocar cualquier upstream ni el auth-service:

| Zona | Rango | Limite |
|---|---|---|
| `login` | `/api/login` | 10 requests por minuto |
| `api` | `/api/*` | 300 requests por minuto |

La clave es la IP real de conexion (`{remote_host}`), no `X-Forwarded-For`, que
el cliente puede falsificar. Un request bloqueado responde 429 y no llega al
auth-service, asi que tampoco consume CPU de bcrypt.

`/health` y los assets del SPA quedan fuera de los limites: un 429 en la API no
debe impedir que la app cargue.

El rate limiting es un modulo de Caddy (`caddy-ratelimit`), asi que la imagen
del gateway se compila con `xcaddy`. `docker compose build gateway` es
obligatorio la primera vez; usar la imagen oficial de Caddy hace que el gateway
no arranque con un error de sintaxis en el Caddyfile.

## 7. Health checks (#21)

Cada upstream declara por que path Caddy lo interroga cada 10s y lo saca de la
rotacion si no responde:

| Upstream | Path | Que detecta |
|---|---|---|
| `backend` | `/api/health` | Proceso colgado y Postgres caido (503) |
| `saas-api` | `/api/saas/health` | Proceso colgado |

El health del Backend pingea Postgres y Redis: si Postgres cae devuelve 503 y el
gateway lo saca de rotacion; si solo cae Redis devuelve 200 con estado
`degraded` y sigue en rotacion, porque el Backend sirve sin cache.

El health del SaaS no chequea dependencias: solo confirma que el proceso
atiende.

## 8. Logs estructurados (#21)

Los tres servicios escriben una linea JSON por request en stdout, con los
mismos campos base: `ts` (RFC3339 en UTC, sufijo `Z`), `level` y `service`.

```bash
docker compose logs backend --no-log-prefix --tail=1 | jq .
docker compose logs -f gateway saas-api    # stream combined
```

Ejemplo del backend:

```json
{"ts":"2026-01-15T18:42:03.123456Z","level":"info","service":"backend","status":200,"latency":"1.2ms","method":"GET","path":"/api/productos","ip":"172.18.0.1","bytes":812,"error":""}
```

Para leerlos a ojo en local hay dos formas: `docker compose logs backend | jq -r
.msg`, o poner `LOG_FORMAT=text` en el `.env` del SaaS.

## 9. Imagenes y despliegue (#21)

`docker compose up` construye desde el codigo. Para desplegar imagenes ya
construidas:

```bash
docker compose pull     # imagenes publicadas en GHCR
docker compose up -d
```

El workflow `.github/workflows/deploy.yml` construye las 4 imagenes y las
publica en `ghcr.io/yan2005dris-afk/{gateway,backend,saas-api,auth-service}`
en cada push a `main` y en cada tag `v*.*.*`. Usa el `GITHUB_TOKEN` del propio
workflow, asi que no hay secretos que configurar.

El pipeline **no despliega**: publicar una imagen no es ponerla corriendo. El
`docker compose pull && docker compose up -d` en el host es un paso manual y a
proposito.

Las imagenes de produccion se configuran con `SITE_ADDRESS` y `ACME_EMAIL` en
el `.env` del host, nunca en el `.env` versionado.

## 10. Diagnostico

| Sintoma | Causa probable |
|---|---|
| El gateway no arranca con "unrecognized global option" | `SITE_ADDRESS` definida y vacia en el `.env` |
| 502 en `/api/*` | Un upstream caido: `docker compose ps` y `logs` del servicio |
| 429 inesperado | Rate limit de la zona `api` o `login`: 10/min en login |
| El navegador no confia en el TLS local | Es lo esperado con la CA interna de Caddy |
| `disk quota exceeded` al compilar Go | `/tmp` es un tmpfs chico: usar `TMPDIR=$HOME/.cache/go-tmp` |