import { type ReactNode } from "react";
import { RouterProvider } from "@tanstack/react-router";
import { AuthProvider } from "@/Context/AuthProvider";
import { CartProvider } from "@/Context/CartProvider";
import { useAuth } from "@/Context/useAuth";
import { UpdateBanner } from "@/Components/Pwa/UpdateBanner";
import { OfflineGuard } from "@/Components/Pwa/OfflineGuard";
import { CheckoutRetryWatcher } from "@/Components/Pwa/CheckoutRetryWatcher";
import { router } from "@/router";

/**
 * Spinner mientras AuthProvider hidrata la sesion contra /api/me.
 * Evita que un usuario con cookie valida parpadee en /login, y evita que los
 * guards de `beforeLoad` corran con `isAuthenticated: false` antes de tiempo.
 */
const HydrationSpinner = () => (
  <div className="bg-muted flex min-h-screen items-center justify-center">
    <div className="text-muted-foreground text-sm">Cargando sesion...</div>
  </div>
);

/**
 * CartBoundary remonta el CartProvider segun el email del usuario para
 * garantizar el aislamiento e integridad de los datos del carrito por cuenta.
 * Se conserva tal cual estaba en el App.tsx previo a la migracion.
 */
const CartBoundary = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  return <CartProvider key={user?.email ?? "anonimo"}>{children}</CartProvider>;
};

/**
 * Conecta el estado de auth con el router.
 *
 * TanStack Router resuelve los guards en `beforeLoad`, que no puede usar
 * hooks. Este componente le inyecta el contexto de auth al router y lo
 * invalida cuando cambia, para que los guards vean el estado real.
 */
const RouterBridge = () => {
  const { user, isAuthenticated, loading } = useAuth();

  if (loading) return <HydrationSpinner />;

  router.update({
    context: {
      auth: { user, isAuthenticated, loading },
    },
  });

  return <RouterProvider router={router} />;
};

/**
 * Arbol de providers de la app. Vive aca y no en `main.tsx` para que ese
 * archivo sea solo el punto de entrada, y para que `react-refresh` pueda
 * recargar los componentes sin recargar el router entero.
 */
export const AppShell = () => (
  <AuthProvider>
    <CartBoundary>
      {/* El banner va fuera del RouterBridge a proposito: ese componente se
          reemplaza por el spinner mientras se hidrata la sesion, y el aviso de
          actualizacion debe sobrevivir a ese estado (U4-T3). */}
      <UpdateBanner />
      {/* Cae a /offline en cuanto se pierde la conexion. */}
      <OfflineGuard />
      {/* Reintenta los checkouts que quedaron a medio enviar (U4-T3). */}
      <CheckoutRetryWatcher />
      <RouterBridge />
    </CartBoundary>
  </AuthProvider>
);