import { createRouter } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";
import type { RouterContext } from "./router-context";

/**
 * Router de la aplicacion.
 *
 * `defaultPreload: "intent"` precarga la ruta al pasar el mouse por un Link,
 * que es el comportamiento por defecto de shadcn-admin.
 */
export const router = createRouter({
  routeTree,
  defaultPreload: "intent",
  // El contexto arranca sin sesion; `main.tsx` lo actualiza cuando
  // AuthProvider hidrata contra /api/me.
  context: {
    auth: {
      user: null,
      isAuthenticated: false,
      loading: true,
    },
  } satisfies RouterContext,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
