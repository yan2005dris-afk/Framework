import { type ReactNode, useState } from "react";
import { TooltipProvider } from "@/components/ui/tooltip";
import Sidebar from "./Sidebar";
import Navbar from "./Navbar";

/**
 * Shell de la aplicacion: sidebar + navbar + area de contenido.
 *
 * Recibe el contenido como `children` (no usa `Outlet`) porque el router de
 * TanStack monta el layout desde `_auth.tsx`, que ya renderiza el `<Outlet />`
 * de la ruta hija.
 *
 * El estado del shell son dos banderas independientes, igual que el `Sidebar`
 * collapsible de shadcn-admin: el colapso aplica al panel de escritorio y el
 * estado abierto al drawer movil. El header expone un trigger por viewport
 * (`md:hidden` / `hidden md:inline-flex`) en vez de preguntar por
 * `window.innerWidth`, que se desincroniza al rotar el telefono.
 */
const Layout = ({ children }: { children: ReactNode }) => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    // Necesario para los tooltips del riel angosto; aca se monta una sola vez
    // para las dos instancias del sidebar (panel y drawer).
    <TooltipProvider>
      <div className="flex h-screen overflow-hidden bg-slate-50">
        <Sidebar
          isCollapsed={isCollapsed}
          mobileOpen={mobileOpen}
          onCloseMobile={() => setMobileOpen(false)}
        />

        <div className="flex min-w-0 flex-1 flex-col">
          <Navbar
            isCollapsed={isCollapsed}
            onToggleSidebar={() => setIsCollapsed((prev) => !prev)}
            onOpenMobileNav={() => setMobileOpen(true)}
          />
          <main className="flex-1 overflow-x-hidden overflow-y-auto p-4 sm:p-6 md:p-8">
            {children}
          </main>
        </div>
      </div>
    </TooltipProvider>
  );
};

export default Layout;
