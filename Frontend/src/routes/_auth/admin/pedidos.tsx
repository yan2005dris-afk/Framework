import { createFileRoute, redirect } from "@tanstack/react-router";
import AdminPedidos from "@/Components/Pedidos/AdminPedidos";

export const Route = createFileRoute("/_auth/admin/pedidos")({
  beforeLoad: ({ context }) => {
    if (context.auth.user?.rol !== "admin") {
      throw redirect({ to: "/" });
    }
  },
  component: AdminPedidos,
});
