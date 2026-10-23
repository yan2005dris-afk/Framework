import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";
import type { RouterContext } from "@/router-context";

/**
 * Ruta raíz. El contexto del router (`RouterContext`) trae el estado de auth
 * para que los guards de las rutas hijas puedan consultarlo sin hooks.
 */
export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => <Outlet />,
});
