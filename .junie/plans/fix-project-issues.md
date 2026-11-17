---
sessionId: session-261117-075547-15o0
---

# Problemas Identificados

## Problemas que necesitan corrección

Después de revisar el proyecto a fondo, estos son los issues concretos que encontré:

### 1. Go 1.26.0 no existe (crítico — break del build)
Los archivos declaran `go 1.26.0` pero la última versión estable de Go es **1.25.x** (según [go.dev/dl](https://go.dev/dl)). Go 1.26 nunca se lanzó. Esto afecta:
- `Backend/go.mod` — línea `go 1.26.0`
- `auth-service/go.mod` — línea `go 1.26.0`
- `tests/integration/go.mod` — línea `go 1.26.0`
- `Backend/Dockerfile` — `FROM golang:1.26-alpine AS builder`
- `auth-service/Dockerfile` — `FROM golang:1.26-alpine AS builder`

El build de Docker fallará con `golang:1.26-alpine` porque la imagen no existe en Docker Hub.

### 2. Tablas del Backend en schema `public` (acoplamiento)
Las tablas `productos`, `referidos`, `pedidos`, y `pedido_items` viven en `public` sin un schema dedicado `backend.*`. La documentación (`docs/architecture.md`) dice que el Backend debería poseer el schema `backend.*`, pero nunca se creó ni se migraron las tablas.

### 3. Frontend `.env` faltante
Existe `Frontend/.env.example` pero no `Frontend/.env`. Al correr `pnpm dev` localmente, Vite usará el fallback de `http.ts` (`http://localhost:8080/api`), pero es mejor tener el archivo versionado para consistencia.

### 4. `.env.example` sin advertencia de seguridad
El `.env.example` en la raíz tiene `POSTGRES_PASSWORD=postgres` y `AUTH_JWT_SECRET=replace-this-with-a-strong-random-value-in-production`, pero no tiene un comentario explícito advirtiendo que **no se debe usar en producción** con esas credenciales.

### Lo que YA está resuelto (no requiere acción)
- **Índices faltantes**: `003_indices.sql` ya existe con `idx_referidos_parent_id` y `idx_productos_categoria`. ✅
- **`models.APIError`**: Se usa como struct literal en todos los controllers correctamente. ✅
- **`package.json` name**: Ya es `"multicatalogo-frontend"`. ✅

# Delivery Steps

### ✓ Step 1: Corregir versión de Go a 1.27 en todos los módulos y Dockerfiles
Cambiar `go 1.26.0` a `go 1.27` en los 3 archivos `go.mod` y `golang:1.26-alpine` a `golang:1.27-alpine` en los 2 Dockerfiles.
Go 1.27 se lanzó el 19 de agosto de 2026 (versión estable actual: 1.27.2).

- `Backend/go.mod`: `go 1.26.0` → `go 1.27` ✅
- `auth-service/go.mod`: `go 1.26.0` → `go 1.27` ✅
- `tests/integration/go.mod`: `go 1.26.0` → `go 1.27` ✅
- `Backend/Dockerfile`: `golang:1.26-alpine` → `golang:1.27-alpine` ✅
- `auth-service/Dockerfile`: `golang:1.26-alpine` → `golang:1.27-alpine` ✅

### ✓ Step 2: Crear schema `backend` y migrar tablas desde `public`
Crear una nueva migración `010_backend_schema.sql` que crea el schema `backend` y mueve las tablas existentes.

- Crear `CREATE SCHEMA IF NOT EXISTS backend` ✅
- Mover tablas: `ALTER TABLE productos SET SCHEMA backend`, igual para `referidos`, `pedidos`, `pedido_items` ✅
- Actualizar las FK y referencias internas ✅
- Actualizar `init.sql` para crear las tablas en `backend.*` directamente ✅
- Actualizar los queries SQL en `Backend/repository/repository.go` para usar `backend.tabla` ✅
- Migraciones 002-009 actualizadas con `backend.*` ✅

### ✓ Step 3: Crear Frontend/.env y mejorar advertencias de seguridad en .env.example
Agregar el archivo `.env` del frontend y mejorar las advertencias de seguridad.

- Crear `Frontend/.env` con `VITE_API_URL=http://localhost:8080/api` ✅
- `Frontend/.env` ya cubierto por `.gitignore` (línea `.env`) ✅
- En `.env.example` de la raíz, agregar bloque de advertencia al inicio ✅
  - `# ⚠️ NO USAR EN PRODUCCIÓN CON ESTAS CREDENCIALES` ✅
  - Instrucciones para generar `AUTH_JWT_SECRET` con `openssl rand -hex 48` ✅
  - Instrucciones para cambiar `POSTGRES_PASSWORD` ✅