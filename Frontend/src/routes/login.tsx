import { createFileRoute } from "@tanstack/react-router";
import Login from "@/Components/Auth/Login/Login";

/**
 * Ruta publica de login. Si ya hay sesion, redirige al destino segun rol
 * (mismo comportamiento que tenia el guard de react-router).
 */
export const Route = createFileRoute("/login")({
  beforeLoad: ({ context }) => {
    if (context.auth.isAuthenticated) {
      throw { redirectTo: context.auth.user?.rol === "admin" ? "/" : "/tienda" };
    }
  },
  component: Login,
});
