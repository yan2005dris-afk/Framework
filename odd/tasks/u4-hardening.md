# U4 hardening: rate limit, TLS, observabilidad y deploy

Objetivo: cerrar las issues abiertas de Fase 4 (#20 TLS, #21 rate limit + observabilidad + deploy).
Rama de trabajo: `feat/u4-hardening` (una por slice). Runner: `go test ./...` (Backend),
`go test -tags=integration` (tests/integration), `pytest` (saas-api), smoke con `curl` contra `:8080`.

## Specs

S1. Rate limiting en Caddy: limitar requests por IP para evitar abuso. (issue #21, punto 1)
S2. Healthchecks reales: cada servicio expone `/health` y Caddy los usa para no rutear a instancias caídas. (issue #21, punto 2)
S3. Logs estructurados JSON en Backend y SaaS, con timestamp unificado. (issue #21, punto 3)
S4. Pipeline de deploy: GitHub Actions que buildea las imagenes y las pushea a un registry. (issue #21, punto 4)
S5. Documentacion de operacion: como arrancar todo con un unico `docker compose up`. (issue #21, punto 5)
S6. TLS con Caddy `auto_https` y Let's Encrypt, con renovacion automatica. (issue #20)
S7. Para desarrollo local con TLS: alternativa local sin depender de un dominio publico. (issue #20)

## Tasks

- T1 (S5) Documentar la operacion actual: servicios, puertos, orden de arranque. Route: inline.
  Hecho en `docs/operacion.md` (10 secciones: arranque, puertos, orden, HMR, TLS, rate
  limit, health checks, logs, despliegue, diagnostico), enlazado desde el README.
  Commit: `b13ef5d`.
- T2 (S1) Rate limit por IP en Caddy. Requiere build con `xcaddy` + `caddy-ratelimit`. Route: inline.
  Commit: `f6a9373`. DONE y verificado en runtime (ver L13).
- T3 (S3) Logs JSON en Backend (Fiber) y SaaS (FastAPI/uvicorn). Route: inline.
  Commit: `bf21b82`. DONE (`go test ./...` 24 passed, `pytest` 3 passed).
- T4 (S2) Health checks activos de Caddy hacia los upstreams. Route: inline.
  Commit: `ee1dc7e`. DONE (`caddy validate` OK; se ven los `health_checker.active` en el log).
- T5 (S4) Workflow de build + push a GHCR. Route: inline. Registry decidido: GHCR.
  Commit: `9381d1b`. Hecho: `.github/workflows/deploy.yml` + un `.dockerignore` por servicio.
  Build de los 4 servicios verificado en local; el push a GHCR no se puede
  verificar sin el PR (no corre en una rama de trabajo).
- T6 (S6-S7) TLS parametrizado por env: `SITE_ADDRESS` + `ACME_EMAIL`. Route: inline.
  Bloqueo resuelto en L6. Commit: `15678e5`. Hecho y validado con `caddy adapt` (ver L9).
- T1 (S5) Documentar la operacion actual. Route: inline. Commit: `b13ef5d`.
  Hecho en `docs/operacion.md` (10 secciones: arranque, puertos, orden, HMR, TLS, rate
  limit, health checks, logs, despliegue, diagnostico), enlazado desde el README.

## Delivery strategy

`ask-on-risk`. Un PR por slice; T2, T3 y T6 tocan `gateway/Caddyfile`, asi que van
secuenciales, nunca en paralelo.

Cortes de slice (el hook de commits avisa al pasar ~400 lineas):

- **Slice 1** - "el gateway deja de ser un proxy abierto y los servicios hablan un
  idioma de log comun": T2 + T3. Commits `f6a9373` y `bf21b82`, 426 lineas.
- **Slice 2** - "resiliencia, despliegue y operacion": T4 + T5 + T6 + T1. El PR
  lleva los dos cambios de infraestructura (workflow de GHCR y TLS) mas los
  health checks y la doc de operacion.

## Log

L1. Pedido del usuario (verbatim): "abriste pr? arreglate esas cosas que encontraste sino arreglalo o quita
    el codigo deprecado o muerto, necesito que termines las issues de la fase 4 y actualices el frotnend"
L2. Hecho antes de este documento: PR #39 (JWT fuera del body del login), PR #41 (quitar
    `AuthController.Refresh` sin ruta), issue #19 cerrada con evidencia (SPA estatico en el gateway).
L3. Correccion de un diagnostico propio: `Frontend/src/routeTree.gen.ts` ya estaba en
    `Frontend/.gitignore` desde `71a6067`. Aparecia como untracked solo porque el checkout local
    estaba en `develop@22e2edb`, previo a esa regla. No hay nada que arreglar ahi.
L4. Hallazgo que bloquea #20: `auto_https` + Let's Encrypt necesita un dominio real y un email de
    contacto, o bien TLS interno para local (que el browser no confia). El repo no declara dominio
    de deploy en ninguna parte. Decision pendiente del usuario.
L5. "Actualizar el frontend" no esta especificado. Se interpreta como alinear el frontend con lo que
    ya existe en el repo (el shell de shadcn-admin ya se porto en el PR #37). Pendiente de confirmar
    alcance con el usuario.
L6. Decisiones del usuario sobre el bloqueo de L4: (a) TLS parametrizado por env, sin dominio hoy:
    el Caddyfile usa `{$SITE_ADDRESS:http://localhost:8080}` y `tls {$ACME_EMAIL:}`; con
    `SITE_ADDRESS=app.dominio.com ACME_EMAIL=...` Caddy pide el certificado a Let's Encrypt y lo
    renueva solo. #20 se cierra con codigo y doc sin esperar a tener un dominio. (b) Registry: GHCR
    (`ghcr.io/yan2005dris-afk/*`) con `GITHUB_TOKEN`. (c) El pipeline solo buildea y pushea; el
    deploy al host es paso manual. Con esto T5 y T6 dejan de estar bloqueados.
L7. El middleware `logger` de Fiber no sirve para JSON: su `Format` interpola `${path}` y `${error}`
    sin escapar, asi que una comilla o un salto de linea inyectan JSON invalido y el agregador
    pierde la linea. Ademas `logger.Config.TimeZone` es `string` en fiber v2.52.15 (no
    `*time.Location`), por lo que un primer intento con `time.UTC` no compilaba. Reemplazado por el
    paquete `Backend/requestlogger`, que serializa con `encoding/json` y tiene tests del escaping.
L8. `/tmp` es un tmpfs de 7.5G al 80% (ocupado por `/tmp/opencode`, de otra herramienta), y ahi es
    donde Go compila: `go test` fallaba con `disk quota exceeded` al enlazar los binarios. Se corre
    con `TMPDIR=$HOME/.cache/go-tmp`. No es un problema del repo, pero hay que acordarlo para que los
    runners no den falsos negativos.
L9. El placeholder del Caddyfile NO admite variable definida y vacia: si `SITE_ADDRESS` existe pero
    esta vacio, el bloque se interpreta como site global y Caddy muere con "unrecognized global
    option". Por eso el compose pasa las variables en forma de lista sin valor (`- SITE_ADDRESS`),
    que hace que Compose las omita cuando no estan definidas. Verificado con `docker compose config`
    (quedan como `null`) y con `caddy adapt`.
L10. Tres modos de TLS, los tres validados con `caddy adapt` (no solo `validate`, que se cuelga con
    el issuer ACME): sin variables -> `:8080` sin TLS; `SITE_ADDRESS=https://localhost:8443` ->
    `:8443` con CA interna (esto es S7, el TLS local sin dominio publico);
    `SITE_ADDRESS=api.tudominio.com ACME_EMAIL=...` -> `:443` con Let's Encrypt. Riesgo asumido y
    documentado: un deploy real sin `ACME_EMAIL` emite un certificado de CA interna que ningun
    browser confia, y eso no falla de forma visible.
L11. No existia `.dockerignore` en `Backend/`, `auth-service/` ni `fastAPI_SaaS/`, y los dos
    Dockerfiles de Go hacen `COPY . .`: el `.env` local de `Backend/` (JWT secret y credenciales de
    Postgres) viajaba como contexto de build. No llegaba a la imagen final, pero queda en el build
    cache de cada maquina que compila. Agregado un `.dockerignore` por servicio.
L12. `ENV CADDY_PORT=8080` en `gateway/Dockerfile` estaba muerto: Caddy no lee ninguna variable con
    ese nombre, el puerto viene del Caddyfile. Eliminado.
L13. Smoke con el stack entero arriba (6/6 healthy): `/health` responde 200; 15 POST seguidos a
    `/api/login` dan 10x401 y despues 5x429, o sea el limite de la zona login se comporta como dice
    el Caddyfile; y los logs de backend, saas-api y gateway salen como JSON valido al parsearlos.
    Diferencia conocida que NO se corrigio: Caddy emite su `ts` como epoch en nanosegundos
    (`1792720524.123`) y no en RFC3339 como Backend y SaaS. Un agregador tiene que convertirlo, pero
    si algun dia se quiere alinear en el origen, hay que hacerlo con `format filter` del Caddyfile.
L14. Todas las tasks del slice 1 y del slice 2 estan commiteadas y verificadas. Lo que falta para
    cerrar #20 y #21 es la parte que decide el humanso: push de la rama, apertura de los PR y
    cierre de las issues una vez mergeado.
