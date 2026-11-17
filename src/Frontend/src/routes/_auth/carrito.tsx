import { createFileRoute } from "@tanstack/react-router";
import Carrito from "@/Components/Carrito/Carrito";

export const Route = createFileRoute("/_auth/carrito")({
  component: Carrito,
});
