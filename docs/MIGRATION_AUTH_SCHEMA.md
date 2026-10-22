# Migración: Eliminación de duplicación de usuarios

## Problema

El Backend tenía una tabla `public.usuarios` que era redundante con `auth.users` del auth-service. Esto causaba:

- Duplicación de datos
- Sincronización manual necesaria
- Fuente de verdad confusa

## Solución

### 1. Cambios en el Backend

**`Backend/repository/repository.go`**
- Cambiado query de `SELECT ... FROM usuarios` a `SELECT ... FROM auth.users`
- El Backend ahora consulta exclusivamente el schema `auth` para autenticación

**`Backend/database/init.sql`**
- Eliminada la creación de tabla `usuarios`
- Eliminados los seeds de usuarios (ahora manejados por auth-service)
- Agregada nota explicativa sobre la migración
- `pedidos.usuario_id` queda **sin** `REFERENCES`: este script corre como
  `docker-entrypoint-initdb.d`, antes de que el auth-service cree `auth.users`.
  La FK la agrega la migración 004 (ver abajo).

**`Backend/database/migrations/002_checkout.sql`**
- Actualizada FK de `pedidos.usuario_id` para referenciar `auth.users` en lugar de `public.usuarios`

### 2. Nueva migración

**`Backend/database/migrations/004_migrate_to_auth_schema.sql`**
- Copia a `auth.users` los usuarios de `public.usuarios` que falten (`ON CONFLICT DO NOTHING`)
- Reapunta la FK de `pedidos.usuario_id` a `auth.users(id) ON DELETE CASCADE`
- Elimina la tabla redundante `public.usuarios` (sin `CASCADE`, a propósito: si algo más dependiera de ella, el error debe verse en el log)
- Idempotente: se puede aplicar las veces que haga falta

### 3. Dockerfile + entrypoint actualizados

**`Backend/Dockerfile`**
- Agregado `postgresql-client` para ejecutar migraciones
- Copia `database/migrations/` dentro de la imagen (sin esto el entrypoint no
  encuentra el directorio y el Backend arrancaría contra un esquema viejo)
- `ENTRYPOINT` pasa a ser `apply_migrations.sh`

**`Backend/scripts/apply_migrations.sh`** (mismo patrón que `auth-service/scripts/entrypoint.sh`)
- Espera activa a que la BD responda (hasta 30s)
- Falla ruidosamente si el directorio de migraciones no existe
- Aplica cada `.sql` con `psql -v ON_ERROR_STOP=1`: si una migración falla, el
  script sale con error y Docker reinicia el contenedor. Nunca queda un Backend
  corriendo contra un esquema a medio migrar
- Recién entonces `exec` del binario

### 4. Orden de arranque (importante)

Las migraciones del Backend referencian `auth.users`, que crea el entrypoint
del **auth-service**. Por eso `backend` depende de `auth-service` en
`docker-compose.yml`. Si el Backend arranca antes, la migración falla con un
error explícito (`auth.users no existe`) y el contenedor reintenta.

## Arquitectura resultante

```
Base de datos: multicatalogo
├── Schema: auth (gestionado por auth-service)
│   └── Tabla: auth.users (FUENTE DE VERDAD para autenticación)
│
└── Schema: public (gestionado por Backend y SaaS)
    ├── productos
    ├── referidos
    ├── pedidos (FK → auth.users)
    ├── pedido_items
    ├── tenants (SaaS)
    └── users_saas (SaaS)
```

## Proceso de migración

Para un entorno existente con datos:

1. Levantar el stack con los cambios
2. La migración `004_migrate_to_auth_schema.sql` se aplicará automáticamente
3. Los datos de `public.usuarios` se migrarán a `auth.users`
4. La tabla `public.usuarios` se eliminará
5. El Backend empezará a usar `auth.users`

## Validación

Para verificar que la migración funcionó correctamente:

```bash
# 1. Verificar que auth.users existe y tiene datos
docker exec -it multicatalogo-db psql -U postgres -d multicatalogo -c "SELECT * FROM auth.users;"

# 2. Verificar que public.usuarios ya no existe
docker exec -it multicatalogo-db psql -U postgres -d multicatalogo -c "\dt public.*"

# 3. Verificar que pedidos tiene la FK correcta
docker exec -it multicatalogo-db psql -U postgres -d multicatalogo -c "\d pedidos"

# 4. Verificar que el Backend'arrancó sin errores de migración
docker compose logs backend | grep -E "\[migrations\]"

# 5. Ejecutar tests E2E
bash tests/e2e/gateway_e2e.sh
```

## Beneficios

- **Single source of truth**: `auth.users` es la única tabla de usuarios
- **Centralización**: auth-service gestiona toda la lógica de autenticación
- **Consistencia**: No hay riesgo de desincronización
- **Mantenibilidad**: Menos código, menos duplicación
