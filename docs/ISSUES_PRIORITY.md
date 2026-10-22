# Análisis y Priorización de Issues

## Estado Actual

Issues abiertas: 7 (#8, #18, #19, #20, #21 y, hasta el merge, #15/#16/#17)
Issues cerradas recientemente: 1 (#30 - U3-T2)

> Las tres primeras de esta lista (#15, #17, #16) están implementadas y
> verificadas en la rama `feature/u3-t3-saas-multitenant` (54 tests
> unitarios/API + 8 de integración RLS contra Postgres). Se cierran al
> mergear el PR; esta tabla describe el estado en `main`.

## Issues por Prioridad

### 🔴 ALTA PRIORIDAD (Críticas para el funcionamiento)

#### #15 - U3-T3: CRUD de users_saas
- **Estado**: IMPLEMENTADO (rama `feature/u3-t3-saas-multitenant`), pendiente de merge
- **Dependencia**: U3-T2 (completada ✅)
- **Descripción**: CRUD de usuarios SaaS multi-tenant
- **Endpoints**:
  - POST /api/saas/tenants/{tenant_id}/users
  - GET /api/saas/tenants/{tenant_id}/users
  - PATCH /api/saas/tenants/{tenant_id}/users/{user_id}
  - DELETE /api/saas/tenants/{tenant_id}/users/{user_id}
- **Impacto**: Necesario para que el SaaS sea funcional
- **Recomendación**: Implementar inmediatamente

#### #17 - U3-T3: middleware de aislamiento multi-tenant
- **Estado**: IMPLEMENTADO (rama `feature/u3-t3-saas-multitenant`), pendiente de merge
- **Dependencia**: U3-T2 (completada ✅)
- **Descripción**: Middleware que inyecta WHERE tenant_id en cada query
- **Opciones**:
  - SQLAlchemy event listener
  - RLS (Row-Level Security) de Postgres
  - ORM-level session
- **Impacto**: Seguridad crítica - sin esto, un tenant puede ver datos de otro
- **Recomendación**: Implementar inmediatamente después de #15

### 🟡 MEDIA PRIORIDAD (Funcionalidades importantes)

#### #16 - U3-T3: planes/suscripciones + billing
- **Estado**: IMPLEMENTADO (rama `feature/u3-t3-saas-multitenant`), pendiente de merge
- **Dependencia**: U3-T3 users_saas (#15)
- **Descripción**: API para planes, suscripciones y webhooks de Stripe
- **Endpoints**:
  - GET /api/saas/plans
  - GET /api/saas/tenants/{tenant_id}/subscription
  - POST /api/saas/tenants/{tenant_id}/subscription
  - DELETE /api/saas/tenants/{tenant_id}/subscription
  - POST /api/saas/webhooks/stripe
- **Impacto**: Monetización del SaaS
- **Recomendación**: Implementar después de #15 y #17

#### #18 - U3-T4+: tests E2E con Playwright
- **Estado**: OPEN
- **Dependencia**: U3-T2, U3-T3 (#15, #16, #17)
- **Descripción**: Tests E2E completos con Playwright
- **Casos**:
  - Login B2C → checkout → pedidos
  - Login SaaS → admin crea tenant → invita user
  - Cross-tenant isolation
- **Impacto**: Calidad y confianza en el sistema
- **Recomendación**: Implementar después de completar U3-T3

### 🟢 BAJA PRIORIDAD (Mejoras y deploy)

#### #19 - U4: build estático SPA + Caddy sirviendo /
- **Estado**: OPEN
- **Dependencia**: U3-T2 (completada ✅)
- **Descripción**: Build de Vite y servir con Caddy
- **Impacto**: Deployment frontend
- **Recomendación**: Implementar cuando se tenga el frontend listo

#### #20 - U4: TLS con Caddy auto_https y Let's Encrypt
- **Estado**: OPEN
- **Dependencia**: U4 (#19)
- **Descripción**: Certificados SSL con Let's Encrypt
- **Impacto**: Seguridad en producción
- **Recomendación**: Implementar solo para producción (no para dev)

#### #21 - U4: rate_limit + observabilidad + deploy pipeline
- **Estado**: OPEN
- **Dependencia**: U4 (#19, #20)
- **Descripción**: Endurecimiento del gateway
- **Componentes**:
  - Rate limiting
  - Healthchecks
  - Logs estructurados
  - Deploy pipeline (GH Actions)
- **Impacto**: Operacionalización
- **Recomendación**: Implementar para producción

### ⚪ REFACTORING (Mejora técnica)

#### #8 - refactor(backend): migrar router de Fiber a Chi
- **Estado**: OPEN
- **Dependencia**: Ninguna
- **Descripción**: Migrar de Fiber (fasthttp) a Chi (net/http)
- **Motivación**:
  - Compatibilidad con ecosistema estándar Go
  - Mejor integración con middlewares
  - Soporte nativo HTTP/2
- **Impacto**: Mejor mantenibilidad a largo plazo
- **Recomendación**: Implementar como mejora técnica, no bloquea otras features

## Roadmap Recomendado

### Fase 1: Completar U3-T3 (SaaS multi-tenant funcional)
1. ✅ #30 - U3-T2 (completada)
2. 🔴 #15 - CRUD de users_saas
3. 🔴 #17 - Middleware de aislamiento multi-tenant
4. 🟡 #16 - Planes/suscripciones + billing

### Fase 2: Tests de calidad
5. 🟡 #18 - Tests E2E con Playwright

### Fase 3: Deployment (U4)
6. 🟢 #19 - Build estático SPA + Caddy
7. 🟢 #20 - TLS con Let's Encrypt
8. 🟢 #21 - Rate limit + observabilidad + deploy pipeline

### Fase 4: Mejoras técnicas (opcional)
9. ⚪ #8 - Migrar Fiber a Chi

## Resumen

- **Críticas para funcionalidad SaaS**: #15, #17
- **Importantes para monetización**: #16
- **Importantes para calidad**: #18
- **Para producción**: #19, #20, #21
- **Mejora técnica**: #8

**Próxima acción**: mergear `feature/u3-t3-saas-multitenant` y cerrar #15, #17
y #16. Después, #19 (build SPA + Caddy) como base de U4, que desbloquea #20,
#21 y la PWA. #8 sigue fuera de alcance por decisión del usuario.
