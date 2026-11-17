import { createFileRoute } from "@tanstack/react-router";
import Checkout from "@/Components/Carrito/Checkout";

export const Route = createFileRoute("/_auth/checkout")({
  component: Checkout,
});
