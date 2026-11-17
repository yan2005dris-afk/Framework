import { createFileRoute } from "@tanstack/react-router";
import Catalogo from "@/Components/Catalogo/Catalogo";

/**
 * Catálogo. El filtro por categoría vive en la URL para que sea enlazable y
 * sobreviva a un refresh (mismo comportamiento que antes migrando con
 * react-router).
 */
export const Route = createFileRoute("/_auth/catalogo")({
  // `categoria` queda opcional en el tipo para que los `<Link to="/catalogo">`
  // de la app no tengan que pasar `search` (p.ej. desde Tienda o el Sidebar).
  validateSearch: (search: Record<string, unknown>): { categoria?: string } => ({
    ...(typeof search.categoria === "string" ? { categoria: search.categoria } : {}),
  }),
  component: Catalogo,
});
