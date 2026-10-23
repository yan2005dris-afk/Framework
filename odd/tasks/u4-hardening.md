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
- T2 (S1) Rate limit por IP en Caddy. Requiere build con `xcaddy` + `caddy-ratelimit`. Route: inline.
- T3 (S3) Logs JSON en Backend (Fiber) y SaaS (FastAPI/uvicorn). Route: inline.
- T4 (S2) Health checks activos de Caddy hacia los upstreams. Route: inline.
- T5 (S4) Workflow de build + push a GHCR. Route: inline. Bloqueado por: decision de registry.
- T6 (S6-S7) TLS. Route: inline. **Bloqueado por: dominio de deploy o decision de TLS local.**

## Delivery strategy

`ask-on-risk`. Un PR por slice; T2, T3 y T6 tocan archivos compartidos y van secuenciales, no en paralelo.

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
