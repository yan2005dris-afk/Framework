# Tests de Integración de Autenticación

Este directorio contiene tests de integración que verifican que el servicio de autenticación funcione correctamente entre los diferentes backends del sistema.

## Arquitectura

```
Frontend/Cliente → Gateway (Caddy :8080)
                          ↓
            ┌─────────────┴─────────────┐
            ↓                           ↓
    Backend Go (:3000)          SaaS FastAPI (:8000)
            ↓                           ↓
    Auth-Service (:3001) ←──────────────┘
            ↓
    PostgreSQL (:5432)
```

## Objetivos de los Tests

1. **Consistencia de JWT_SECRET**: Verificar que ambos backends usen el mismo secreto para firmar/validar tokens
2. **Emisión de tokens**: Verificar que auth-service emita tokens JWT válidos
3. **Compatibilidad de tokens**: Verificar que un token emitido por auth-service sea aceptado tanto por Backend como por SaaS
4. **Proxies funcionales**: Verificar que los proxies HTTP en Backend y SaaS funcionen correctamente
5. **Gateway forward_auth**: Verificar que el gateway Caddy valide tokens correctamente antes de pasar requests a los backends
6. **Aislamiento multi-tenant (RLS)**: Verificar que las políticas de Row-Level Security de Postgres filtran las filas de cada tenant, y que el bypass de superuser está documentado

## Pre-requisitos

1. **Stack levantado**:
   ```bash
   docker compose up -d --build
   ```
2. **Esperar healthchecks** (~15-30 segundos):
   ```bash
   # Verificar que los servicios respondan
   curl http://localhost:8080/api/tx/health
   curl http://localhost:8080/api/saas/health
   ```
3. **Variables de entorno configuradas**: Asegurar que `AUTH_JWT_SECRET` esté configurado en `.env`

## Ejecutar Tests

### Tests de Integración (Go)

```bash
cd tests/integration
go test -v -tags=integration
```

### Tests de Integración (Python)

```bash
cd tests/integration
uv run --project ../fastAPI_SaaS pytest test_auth_integration.py -v
```

### Tests de RLS del SaaS

Aíslan el tenant en la base, no en el código: necesitan solo la BD (el resto de
servicios no importa).

```bash
# 1. Levantar la BD
docker compose up -d db

# 2. Migrar (esto también crea el rol `saas_app`, ver más abajo)
cd ../../fastAPI_SaaS
SAAS_DB_HOST=localhost SAAS_DB_PORT=5432 SAAS_DB_NAME=multicatalogo uv run alembic upgrade head
cd ../..

# 3. Correr los tests (uv usa el venv del servicio SaaS para psycopg2)
cd tests/integration
SAAS_DB_USER=saas_app SAAS_DB_PASSWORD=saas_app \
  uv run --project ../fastAPI_SaaS pytest test_saas_rls_integration.py -v
```

Los tests se saltan (skip) en vez de fallar si la BD no está o si faltan las
políticas RLS. También se saltan si el rol con el que conectan bypasea RLS: en
ese caso no probarían nada.

### Todos los tests

```bash
cd tests/integration
go test -v -tags=integration && \
  uv run --project ../fastAPI_SaaS pytest test_auth_integration.py -v && \
  uv run --project ../fastAPI_SaaS pytest test_saas_rls_integration.py -v
```

Los dos archivos de Python se lanzan por separado a proposito:
`test_auth_integration.py` importa `jwt` (pyjwt), que hoy no declara ningun
proyecto del repo, asi que si se lanzan juntos su error de import corta la
collection y los tests de RLS no llegan a correr.

O el script, que además levanta el stack y espera los healthchecks:

```bash
bash tests/integration/run_integration_tests.sh            # Go + Python
bash tests/integration/run_integration_tests.sh py         # solo Python
bash tests/integration/run_integration_tests.sh --skip-stack py
```

## Casos de Test

### Go (`auth_integration_test.go`)

| Test | Descripción |
|------|-------------|
| `TestJWTSecretConsistency` | Verifica que AUTH_JWT_SECRET sea consistente entre servicios |
| `TestAuthServiceLoginDirect` | Verifica que auth-service emita tokens JWT válidos |
| `TestGatewayLoginViaBackend` | Verifica login a través del gateway → Backend |
| `TestGatewayLoginViaSaaS` | Verifica login a través del gateway → SaaS |
| `TestTokenReuseBetweenBackends` | Verifica que un token funcione en ambos backends |
| `TestGatewayForwardAuth` | Verifica que forward_auth del gateway funcione |

### Python (`test_auth_integration.py`)

| Test | Descripción |
|------|-------------|
| `test_auth_service_login_direct` | Verifica login directo a auth-service |
| `test_token_reuse_backend` | Verifica token en Backend |
| `test_token_reuse_saas` | Verifica token en SaaS |
| `test_gateway_login_backend` | Verifica login vía gateway (Backend) |
| `test_gateway_login_saas` | Verifica login vía gateway (SaaS) |
| `test_cross_service_token` | Verifica compatibilidad cross-service |

### Python (`test_saas_rls_integration.py`)

Todos con el rol de la **app** (`saas_app`, sin superuser), salvo los dos últimos.

| Test | Descripción |
|------|-------------|
| `test_read_only_returns_own_rows` | Con el GUC en tenant A, `users_saas` solo devuelve filas de A |
| `test_subscriptions_are_isolated` | Ídem para `subscriptions` |
| `test_tenants_row_is_isolated` | `tenants` solo deja ver la fila del tenant propio |
| `test_insert_into_other_tenant_is_denied` | Insertar en otro tenant viola la política (`WITH CHECK`) |
| `test_update_other_tenant_row_is_denied` | `UPDATE` de otra fila afecta 0 filas |
| `test_delete_other_tenant_row_is_denied` | `DELETE` de otra fila afecta 0 filas |
| `test_without_guc_nothing_is_visible` | Fail closed: sin `app.current_tenant` no se ve nada |
| `test_superadmin_path_switches_tenant_and_sees_it` | El `superadmin` cruza de tenant cambiando el GUC |
| `test_superuser_sees_rows_of_every_tenant` | Conectado como `postgres`, RLS **no** filtra (documenta el bypass) |
| `test_superuser_sees_every_tenant_row` | Ídem sobre `tenants` |

## Por qué el SaaS no conecta como `postgres`

Las políticas de la migración `fastAPI_SaaS/alembic/versions/0003_rls_policies.py`
solo se aplican a roles que **no** son superuser ni tienen `BYPASSRLS`
(`FORCE ROW LEVEL SECURITY` ata al owner de la tabla, no al superuser). Por eso
la BD se accede con dos identidades:

| Identidad | Variables | Para qué |
|---|---|---|
| **App** (runtime) | `SAAS_DB_USER`, `SAAS_DB_PASSWORD`, `SAAS_DB_APP_ROLE` | Corre uvicorn. Rol sin superuser, creado por la migración `0004_saas_app_role`. RLS lo protege. |
| **Admin** (migraciones) | `SAAS_DB_ADMIN_USER`, `SAAS_DB_ADMIN_PASSWORD` | Solo `alembic upgrade head`. Necesita DDL para crear tablas, rol y políticas. |

El `docker-compose.yml` arranca `alembic` con las credenciales de admin y
`uvicorn` con las de la app. Si se invirtiera, las políticas quedarían inertes y
cada request vería los datos de todos los tenants: eso es exactamente lo que
prueban los dos tests de `TestSuperuserBypassesRLS`.

## Troubleshooting

### Error: "No se pudo conectar al auth-service"

- Verificar que el stack esté levantado: `docker compose ps`
- Esperar a que los healthchecks pasen: `docker compose logs auth-service`
- Verificar puerto: `curl http://localhost:3001/auth/health`

### Error: "JWT_SECRET del Backend no coincide con AUTH_JWT_SECRET"

- Revisar `.env` en la raíz del proyecto
- Asegurar que `AUTH_JWT_SECRET` esté configurado correctamente
- Reiniciar servicios después de cambiar variables: `docker compose up -d --build`

### Error: "Backend rechazó el token"

- Verificar que el Backend tenga acceso al JWT_SECRET correcto
- Revisar logs del Backend: `docker compose logs backend`
- Verificar que el middleware de autenticación esté configurado

### Error: "SaaS rechazó el token"

- Verificar que el SaaS tenga configurada la URL correcta del auth-service
- Revisar logs del SaaS: `docker compose logs saas-api`
- Verificar que el proxy HTTP esté funcionando

### Error: los tests de RLS aparecen como `skipped`

- `Postgres no disponible`: falta `docker compose up -d db`.
- `Faltan politicas RLS`: correr `alembic upgrade head` (ver la receta de arriba).
- `El rol X bypasea RLS`: se están corriendo como superuser. Apuntar
  `SAAS_DB_USER`/`SAAS_DB_PASSWORD` al rol de la app (`saas_app`), que la
  migración `0004_saas_app_role` crea y sincroniza en cada `upgrade head`.
- `El rol admin X no bypasea RLS`: el test del bypass necesita un superuser;
  dejar `SAAS_DB_ADMIN_USER` como el superuser de la BD.

### Error: "No module named jwt" (test_auth_integration.py)

El archivo de auth usa `pyjwt` y ningun proyecto Python del repo lo declara
todavia (el `pyproject.toml` de `fastAPI_SaaS/` esta fuera del alcance de WU3).
Es independiente de los tests de RLS: correlos por separado.

### Error: la app falla con "permission denied for table ..."

Está corriendo con el rol de la app sobre una BD sin la migración `0004`
aplicada. Correr `alembic upgrade head` con las credenciales de admin: es el
paso que crea el rol y le da los permisos.

## Variables de Entorno Relevantes

| Variable | Servicio | Descripción |
|----------|---------|-------------|
| `AUTH_JWT_SECRET` | auth-service, Backend | Secret para firmar/validar JWTs |
| `AUTH_SERVICE_URL` | Backend, SaaS | URL del auth-service para proxies |
| `GATEWAY_PORT` | Gateway | Puerto del gateway Caddy |
| `BACKEND_PORT` | Backend | Puerto del Backend Go |
| `SAAS_API_PORT` | SaaS | Puerto del SaaS FastAPI |
| `SAAS_DB_HOST` / `SAAS_DB_PORT` / `SAAS_DB_NAME` | tests RLS | Dónde está la BD del SaaS |
| `SAAS_DB_APP_ROLE` / `SAAS_DB_USER` / `SAAS_DB_PASSWORD` | tests RLS | Rol de la app para las pruebas de aislamiento (default `saas_app`) |
| `SAAS_DB_ADMIN_USER` / `SAAS_DB_ADMIN_PASSWORD` | tests RLS | Rol admin para el test del bypass (default `postgres`) |

## Próximos Pasos

- [ ] Agregar tests para refresh token
- [ ] Agregar tests para logout
- [ ] Agregar tests para tokens SaaS con tenant_id
- [ ] Agregar tests para roles (admin vs user)
- [ ] Agregar tests de rendimiento (carga de auth)
- [ ] Agregar tests de RLS para las tablas que faltan al catálogo del rol
      (`products`, `referidos`) cuando el SaaS empiece a leerlas
