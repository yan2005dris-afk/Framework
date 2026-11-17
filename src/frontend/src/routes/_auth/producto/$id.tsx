import { createFileRoute } from "@tanstack/react-router";
import DetalleProducto from "@/Components/Catalogo/DetalleProducto";

/**
 * El param `id` se valida aca y queda tipado en `Route.useParams()`.
 */
export const Route = createFileRoute("/_auth/producto/$id")({
  component: DetalleProducto,
});
