# Task: Arquitectura de Eventos con Valkey (Clientes Nativos Singleton)

Slug: event-driven-valkey-architecture
Status: active
Date: 2026-11-17

## Context

Implementación de arquitectura orientada a eventos con Valkey usando clientes singleton nativos en Go y Python (Streams: consumer groups + ACK) para desacoplar transacciones de catálogo y procesamiento SaaS.

## Specs

S1. Infraestructura Valkey en Docker Compose: Reemplazar imágenes de redis por valkey/valkey:8-alpine en docker-compose.yml y deploy/docker-compose.prod.yml con compatibilidad de puertos y healthcheck.
S2. Cliente Singleton y Publicador en Go (api-catalogo): Crear cliente Valkey Singleton en Go con go-redis (driver estándar nativo compatible con Valkey RESP) y un publicador de eventos tipados en Streams.
S3. Cliente Singleton Asíncrono en Python (api-saas): Crear cliente Valkey Singleton asíncrono en Python usando redis-py nativo en app/shared/valkey.py e integrarlo en lifespan.
S4. Worker Consumidor de Eventos con ACK (api-saas): Implementar Worker de fondo en api-saas con Consumer Group (XREADGROUP + XACK) que escuche 'stream:pedidos' e inserte pedidos en el tenant.
S5. Armonización de Rutas en API Gateway: Ajustar Caddyfile para que /api/saas/tenants/{tenant_id}/pedidos dirija limpiamente a api-saas sin colisionar con rutas B2C de Go.

## Tasks

- [ ] T1: Actualizar docker-compose a Valkey 8: Actualizar imágenes y healthchecks en docker-compose.yml y deploy/docker-compose.prod.yml a valkey/valkey:8-alpine.
- [ ] T2: Implementar Valkey Singleton en api-catalogo: Implementar cliente singleton y helper XADD en src/api-catalogo.
- [ ] T3: Implementar Valkey Async Singleton en api-saas: Añadir redis>=5.0 a pyproject.toml y crear app/shared/valkey.py con cliente async singleton.
- [ ] T4: Implementar Consumer Worker en api-saas: Crear consumer worker con grupo de consumo que procese eventos y registre pedidos multi-tenant.
- [ ] T5: Actualizar Gateway y Documentación de Arquitectura: Verificar rutas en Caddyfile y documentar la arquitectura de eventos en docs/.

## Log

- L1: Task scaffolded for slug "event-driven-valkey-architecture": Arquitectura de Eventos con Valkey (Clientes Nativos Singleton).
