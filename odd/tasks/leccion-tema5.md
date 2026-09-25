# Feature: Lección Tema 5 - Catálogo y Red Multinivel con Endpoints de Backend

## Objective
Implementar la lección completa de interfaces interactivas, catálogo y red multinivel, trasladando los datos mock del frontend al backend en Go/Fiber para ser consumidos mediante endpoints REST (`/api/productos`, `/api/productos/:id`, `/api/red`, `/api/login`).

## Tasks
- [ ] `task-1-backend-endpoints`: Actualizar modelos (`Producto`, `Referido`), controladores (`prodController`, `redController`, `authController`) y rutas (`routes.go`) en Go.
- [ ] `task-2-frontend-services`: Implementar `productosService.ts` y `redService.ts` consumiendo los endpoints del backend.
- [ ] `task-3-frontend-auth-context`: Actualizar `AuthContext.tsx` y `CartContext.tsx` con soporte para roles y persistencia por usuario.
- [ ] `task-4-frontend-views-navigation`: Implementar vistas completas (`Storefront`, `Catalogo`, `DetalleProducto` con lightbox, `MiRed`, `Dashboard`, `Carrito`, `Checkout`, `Confirmacion`) y navegación por roles en `Layout`, `Navbar`, `Sidebar` y `App.tsx`.
- [ ] `task-5-verification`: Verificar compilación de Go backend y build/lint de frontend React con TypeScript.
