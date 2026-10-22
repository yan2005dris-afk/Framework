Feature: Frontend base with shadcn/ui + TanStack Router, taking shadcn-admin as the
reference for structure (sidebar shell, data-table + form + dialog patterns, Recharts
for Dashboard KPIs) without forking it. Goal: migrate the existing react-router-dom v7
app to TanStack Router, keep AuthProvider/http.ts/rutas intact, and leave the component
base ready for the pending SaaS screens.

## Specs

S1. Migrate the router from react-router-dom v7 to TanStack Router, keeping the exact
   same routes, role guards and auth flow (AuthProvider + httpOnly cookie + /api/me).
S2. Initialize shadcn/ui on Tailwind v4 (already present) with the `@/` alias and the
   primitives the shell and SaaS screens need (button, card, table, form, dialog,
   select, tabs, sheet, badge, input, label, dropdown-menu, tooltip, avatar, sonner).
S3. Search params stay in the URL: `/catalogo?categoria=X` must keep working after the
   migration (react-router used `useSearchParams`).
S4. No behavior change in the flows that already work: login by role (admin -> "/",
   user -> "/tienda"), protected layout, product detail by `:id`, cart, checkout ->
   confirmation, 404 fallback.
S5. Zero remaining imports of react-router-dom in `Frontend/src`; the dependency is
   removed from package.json.

## Tasks

- T1 (S1, S5) Finish migrating the three components that still import react-router-dom:
  `Catalogo.tsx`, `Carrito.tsx`, `Storefront.tsx`. Route: inline. Commit: work-unit 2.
- T2 (S3) Type `/catalogo` search params with `validateSearch` and derive the category
  filter from the URL instead of duplicating it in local state. Route: inline.
  Commit: work-unit 2.
- T3 (S4) Type the checkout -> confirmation `state` handoff without `as never`.
  Route: inline. Commit: work-unit 2.
- T4 (S1, S4) Delete the dead `App.tsx` (react-router route tree) and drop
  `react-router-dom` from dependencies. Route: inline. Commit: work-unit 2.
- T5 (S1-S5) Verify: `tsc -b`, `eslint .`, `vite build`, and a smoke run of the dev
  server on the affected routes. Route: inline. Commit: none (verification only).
- T6 (S2) Port the shadcn-admin shell (collapsible sidebar, header, command palette)
  adapted to the current layout. Route: pending user decision on scope.

## Log

L1. "si otra cosa queria usar esto apra refactorizar mi frontend :
    https://github.com/satnaing/shadcn-admin" (user)
L2. Decision (session): do not fork shadcn-admin. Take the patterns and the component
    base; the repo's TanStack Router and Clerk auth are not a drop-in for an app that
    uses react-router-dom v7 with its own cookie auth.
L3. Decision (session): migrate to TanStack Router (chosen over staying on react-router)
    because it matches the reference repo and keeps file-based routing.
L4. Interrupted mid-migration by a usage limit. Resumed with: shadcn initialized, 17
    primitives in place, route tree written, main.tsx wired to RouterProvider, 6 of 9
    router-consuming components migrated. Remaining: T1-T5.