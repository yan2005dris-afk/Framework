# Feature: Fase 2 - Persistencia de Datos Relacionales con PostgreSQL

## Objective
Implementar la persistencia de datos relacionales utilizando PostgreSQL 16 y el driver pgx en Go/Fiber, reemplazando los datos en memoria por consultas SQL en `repository.go` sin alterar las firmas de los controladores ni la compatibilidad con el frontend.

## Tasks
- [x] `task-1-infra-docker`: Configurar `docker-compose.yml` con PostgreSQL 16 y `.env` en `Backend/`.
- [x] `task-2-db-schema-seed`: Crear `database/init.sql` con esquema de tablas (`usuarios`, `productos`, `referidos`) y datos semilla.
- [x] `task-3-db-connection`: Implementar `config/db.go` para inicializar el pool de conexiones pgx y gestionar cierre ordenado.
- [x] `task-4-repository`: Implementar `repository/repository.go` con `BuscarUsuarioPorCredenciales`, `ObtenerProductos`, `ObtenerProductoPorID` y `ObtenerRed` (construcción jerárquica).
- [x] `task-5-controllers-refactor`: Refactorizar controladores (`authController`, `prodController`, `redController`) y `main.go` para consumir `repository` y conectar con la BD.
- [x] `task-6-verification`: Validar compilación, contenedor de PostgreSQL, datos iniciales y endpoints de la API.

Commit: `a97577e`
