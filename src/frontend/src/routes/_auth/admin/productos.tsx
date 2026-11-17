import { createFileRoute, redirect } from "@tanstack/react-router";
import AdminProductos from "@/Components/AdminProductos/AdminProductos";

export const Route = createFileRoute("/_auth/admin/productos")({
  beforeLoad: ({ context }) => {
    if (context.auth.user?.rol !== "admin") {
      throw redirect({ to: "/" });
    }
  },
  component: AdminProductos,
});
