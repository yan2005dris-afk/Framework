import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import Layout from "@/Components/Layout/Layout";

/**
 * Layout protegido: exige sesion activa. Equivale al `ProtectedRoute` de
 * react-router, pero resuelto en `beforeLoad` (antes de montar, sin parpadeo).
 *
 * El spinner de hidratacion se maneja en `main.tsx`: mientras `loading` es
 * true el router no se monta, asi que aca `loading` ya es false.
 */
export const Route = createFileRoute("/_auth")({
  beforeLoad: ({ context, location }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({
        to: "/login",
        search: { redirect: location.href },
      });
    }
  },
  component: () => (
    <Layout>
      <Outlet />
    </Layout>
  ),
});
