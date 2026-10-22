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
uv run pytest test_auth_integration.py -v
```

### Todos los tests

```bash
cd tests/integration
go test -v -tags=integration && uv run pytest test_auth_integration.py -v
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

## Variables de Entorno Relevantes

| Variable | Servicio | Descripción |
|----------|---------|-------------|
| `AUTH_JWT_SECRET` | auth-service, Backend | Secret para firmar/validar JWTs |
| `AUTH_SERVICE_URL` | Backend, SaaS | URL del auth-service para proxies |
| `GATEWAY_PORT` | Gateway | Puerto del gateway Caddy |
| `BACKEND_PORT` | Backend | Puerto del Backend Go |
| `SAAS_API_PORT` | SaaS | Puerto del SaaS FastAPI |

## Próximos Pasos

- [ ] Agregar tests para refresh token
- [ ] Agregar tests para logout
- [ ] Agregar tests para tokens SaaS con tenant_id
- [ ] Agregar tests para roles (admin vs user)
- [ ] Agregar tests de rendimiento (carga de auth)
