import { createFileRoute, redirect } from "@tanstack/react-router";
import MiRed from "@/Components/MiRed";

/**
 * Mi Red: solo admin.
 */
export const Route = createFileRoute("/_auth/mi-red")({
  beforeLoad: ({ context }) => {
    if (context.auth.user?.rol !== "admin") {
      throw redirect({ to: "/tienda" });
    }
  },
  component: MiRed,
});
