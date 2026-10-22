import { createFileRoute } from "@tanstack/react-router";
import Confirmacion from "@/Components/Carrito/Confirmacion";

export const Route = createFileRoute("/_auth/confirmacion")({
  component: Confirmacion,
});
