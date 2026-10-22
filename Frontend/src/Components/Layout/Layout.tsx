import { type ReactNode, useState } from "react";
import Sidebar from "./Sidebar";
import Navbar from "./Navbar";

/**
 * Shell de la aplicacion: sidebar + navbar + area de contenido.
 *
 * Recibe el contenido como `children` (no usa `Outlet`) porque el router de
 * TanStack monta el layout desde `_auth.tsx`, que ya renderiza el `<Outlet />`
 * de la ruta hija.
 */
const Layout = ({ children }: { children: ReactNode }) => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  // Toggle para desktop (collapse/expand a 80px) o abrir drawer en móviles
  const handleToggleSidebar = () => {
    if (window.innerWidth < 768) {
      setMobileOpen((prev) => !prev);
    } else {
      setIsCollapsed((prev) => !prev);
    }
  };

  return (
    <div className="flex h-screen bg-slate-50 overflow-hidden">
      {/* Sidebar con soporte para isCollapsed y responsive drawer en móviles */}
      <Sidebar
        isCollapsed={isCollapsed}
        mobileOpen={mobileOpen}
        onCloseMobile={() => setMobileOpen(false)}
      />

      {/* Área de Contenido Principal */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Navbar
          onToggleSidebar={handleToggleSidebar}
          isCollapsed={isCollapsed}
        />
        <main className="flex-1 overflow-x-hidden overflow-y-auto p-4 sm:p-6 md:p-8">
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout;
