import { createFileRoute, redirect } from "@tanstack/react-router";
import Dashboard from "@/Components/Dashboard/Dashboard";

/**
 * Dashboard: solo admin. Equivale al `AdminRoute` de react-router.
 */
export const Route = createFileRoute("/_auth/")({
  beforeLoad: ({ context }) => {
    if (context.auth.user?.rol !== "admin") {
      throw redirect({ to: "/tienda" });
    }
  },
  component: Dashboard,
});
