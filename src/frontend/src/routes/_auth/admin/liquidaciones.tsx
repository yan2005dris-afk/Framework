import { createFileRoute, redirect } from "@tanstack/react-router";
import AdminLiquidaciones from "@/Components/AdminLiquidaciones/AdminLiquidaciones";

export const Route = createFileRoute("/_auth/admin/liquidaciones")({
  beforeLoad: ({ context }) => {
    if (context.auth.user?.rol !== "admin") {
      throw redirect({ to: "/" });
    }
  },
  component: AdminLiquidaciones,
});
