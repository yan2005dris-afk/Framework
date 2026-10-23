import { createFileRoute } from "@tanstack/react-router";
import MisPedidos from "@/Components/Pedidos/MisPedidos";

export const Route = createFileRoute("/_auth/mis-pedidos")({
  component: MisPedidos,
});
