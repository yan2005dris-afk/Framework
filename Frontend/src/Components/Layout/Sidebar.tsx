import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../../Context/AuthContext";
import { Icon } from "../Share/Icon";
import type { IconName } from "lucide-react/dynamic";

interface NavItem {
  name: string;
  path: string;
  icon: IconName;
  soloAdmin?: boolean;
}

interface SidebarProps {
  isCollapsed: boolean;
  mobileOpen?: boolean;
  onCloseMobile?: () => void;
}

const navItems: NavItem[] = [
  { name: "Dashboard", path: "/", icon: "layout-dashboard", soloAdmin: true },
  { name: "Tienda", path: "/tienda", icon: "shopping-bag" },
  { name: "Catálogo", path: "/catalogo", icon: "store" },
  { name: "Mi Red", path: "/mi-red", icon: "users", soloAdmin: true },
  { name: "Carrito", path: "/carrito", icon: "shopping-cart" },
];

/**
 * Componente Sidebar responsive con filtrado de navegación según el rol del usuario.
 */
const Sidebar = ({
  isCollapsed,
  mobileOpen = false,
  onCloseMobile,
}: SidebarProps) => {
  const { user } = useAuth();
  const location = useLocation();

  const itemsPermitidos = navItems.filter(
    (item) => !item.soloAdmin || user?.rol === "admin"
  );

  return (
    <>
      {/* Mobile Backdrop Overlay */}
      {mobileOpen && (
        <div
          onClick={onCloseMobile}
          className="fixed inset-0 bg-black/50 z-40 md:hidden transition-opacity"
          aria-hidden="true"
        />
      )}

      {/* Sidebar Container */}
      <aside
        className={`
          fixed md:static inset-y-0 left-0 z-50
          ${isCollapsed ? "md:w-[80px]" : "md:w-64"}
          w-64 bg-slate-900 text-white flex flex-col
          transition-all duration-300 ease-in-out
          ${mobileOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0"}
        `}
      >
        {/* Header / Brand */}
        <div className="h-16 flex items-center px-4 border-b border-slate-800 justify-between">
          <div
            className={`flex items-center gap-3 overflow-hidden transition-all duration-300 ${
              isCollapsed ? "md:justify-center md:w-full" : ""
            }`}
          >
            <div className="w-8 h-8 rounded-lg bg-indigo-600 flex items-center justify-center flex-shrink-0 font-bold text-white shadow-md">
              MC
            </div>
            {(!isCollapsed || mobileOpen) && (
              <span className="text-lg font-bold tracking-tight whitespace-nowrap md:inline">
                MultiCatálogo
              </span>
            )}
          </div>

          {/* Mobile close button */}
          <button
            onClick={onCloseMobile}
            className="md:hidden p-1 text-slate-400 hover:text-white rounded-lg"
            aria-label="Cerrar menú"
          >
            <Icon name="x" size={22} />
          </button>
        </div>

        {/* Navigation items */}
        <nav className="flex-1 p-3 space-y-1.5 overflow-y-auto">
          {itemsPermitidos.map((item) => {
            const isActive =
              item.path === "/"
                ? location.pathname === "/"
                : location.pathname.startsWith(item.path);

            return (
              <Link
                key={item.path}
                to={item.path}
                onClick={onCloseMobile}
                title={item.name}
                className={`
                  flex items-center gap-3 p-3 rounded-lg transition-colors group
                  ${
                    isActive
                      ? "bg-indigo-600 text-white font-medium shadow-sm"
                      : "text-slate-300 hover:bg-slate-800 hover:text-white"
                  }
                  ${isCollapsed ? "md:justify-center md:px-0" : "px-3"}
                `}
              >
                <div className="flex-shrink-0 flex items-center justify-center">
                  <Icon name={item.icon} size={22} />
                </div>

                {(!isCollapsed || mobileOpen) && (
                  <span className="whitespace-nowrap transition-opacity duration-200">
                    {item.name}
                  </span>
                )}
              </Link>
            );
          })}
        </nav>

        {/* Footer info in sidebar con Rol */}
        <div className="p-3 border-t border-slate-800 text-xs text-slate-400 text-center">
          {isCollapsed ? (
            <span className="uppercase font-semibold text-indigo-400">
              {user?.rol}
            </span>
          ) : (
            <p>
              Conectado como:{" "}
              <span className="font-semibold uppercase text-indigo-400">
                {user?.rol}
              </span>
            </p>
          )}
        </div>
      </aside>
    </>
  );
};

export default Sidebar;
