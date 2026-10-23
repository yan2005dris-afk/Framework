import { createFileRoute, redirect } from "@tanstack/react-router";
import Register from "@/Components/Auth/Register/Register";

interface RegistroSearch {
  ref?: string;
}

export const Route = createFileRoute("/registro")({
  validateSearch: (search: Record<string, unknown>): RegistroSearch => {
    return {
      ref: typeof search.ref === "string" ? search.ref : undefined,
    };
  },
  beforeLoad: ({ context }) => {
    if (context.auth.isAuthenticated) {
      throw redirect({ to: context.auth.user?.rol === "admin" ? "/" : "/tienda" });
    }
  },
  component: Register,
});
