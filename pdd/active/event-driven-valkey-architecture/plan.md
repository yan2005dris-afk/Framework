# Plan: Arquitectura de Eventos con Valkey (Clientes Nativos Singleton)

## 1. Descripción y alcance
Implementación de arquitectura orientada a eventos con Valkey usando clientes singleton nativos en Go y Python (Streams: consumer groups + ACK) para desacoplar transacciones de catálogo y procesamiento SaaS.

Tamaño estimado: M (~100-400 líneas).
Estrategia de entrega: Iterative implementation: 1. Infraestructura Valkey, 2. Cliente Singleton Go nativo, 3. Cliente Singleton Python nativo, 4. Event Streams y Worker, 5. Armonización de rutas.

## 2. Dependencias previas
- docker-compose.yml
- src/api-catalogo/config/redis.go
- src/api-saas/pyproject.toml

## 3. Specs numeradas
- **S1**: Infraestructura Valkey en Docker Compose — Reemplazar imágenes de redis por valkey/valkey:8-alpine en docker-compose.yml y deploy/docker-compose.prod.yml con compatibilidad de puertos y healthcheck.
- **S2**: Cliente Singleton y Publicador en Go (api-catalogo) — Crear cliente Valkey Singleton en Go con go-redis (driver estándar nativo compatible con Valkey RESP) y un publicador de eventos tipados en Streams.
- **S3**: Cliente Singleton Asíncrono en Python (api-saas) — Crear cliente Valkey Singleton asíncrono en Python usando redis-py nativo en app/shared/valkey.py e integrarlo en lifespan.
- **S4**: Worker Consumidor de Eventos con ACK (api-saas) — Implementar Worker de fondo en api-saas con Consumer Group (XREADGROUP + XACK) que escuche 'stream:pedidos' e inserte pedidos en el tenant.
- **S5**: Armonización de Rutas en API Gateway — Ajustar Caddyfile para que /api/saas/tenants/{tenant_id}/pedidos dirija limpiamente a api-saas sin colisionar con rutas B2C de Go.

## 4. Contrato de entrada
Eventos de pedido emitidos por api-catalogo al finalizar checkout (Stream 'stream:pedidos' con evento 'pedido.creado').

## 5. Contrato de salida
Consumo confiable con XREADGROUP y XACK en api-saas registrando el pedido en el tenant y actualizando contadores/métricas sin pérdida de eventos.

## 6. Seguridad y resiliencia
Reconexión automática, timeouts estrictos, idempotencia por ID de evento y aislamiento multi-tenant en el worker.
- Destructive Testing Plan:
  Simular caída de Valkey y caída del worker FastAPI durante procesamiento para validar persistencia de stream y reintento por XACK pendiente.

## 7. Archivos afectados
- `docker-compose.yml`
- `deploy/docker-compose.prod.yml`
- `src/api-catalogo/config/valkey.go`
- `src/api-catalogo/events/publisher.go`
- `src/api-saas/app/shared/valkey.py`
- `src/api-saas/app/events/consumer.py`
- `src/api-saas/pyproject.toml`
- `src/api-gateway/Caddyfile`
