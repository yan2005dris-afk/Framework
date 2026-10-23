import { useState } from "react";
import { Link, useLocation } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { useAuth } from "../../Context/useAuth";
import { Icon, type IconName } from "../Share/Icon";

export type NavPath =
  | "/"
  | "/tienda"
  | "/catalogo"
  | "/mi-red"
  | "/carrito"
  | "/mis-pedidos"
  | "/admin/productos"
  | "/admin/pedidos"
  | "/admin/liquidaciones"
  | "/saas"
  | "/planes"
  | "/usuarios-saas";

interface NavSubItem {
  name: string;
  path: NavPath;
  icon: IconName;
}

interface NavGroupItem {
  name: string;
  icon: IconName;
  soloAdmin?: boolean;
  subItems?: NavSubItem[];
  path?: NavPath;
}

interface SidebarProps {
  isCollapsed: boolean;
  mobileOpen?: boolean;
  onCloseMobile?: () => void;
}

interface SidebarContentProps {
  isCollapsed: boolean;
  onNavigate?: () => void;
  onClose?: () => void;
}

const navMenu: NavGroupItem[] = [
  { name: "Dashboard", path: "/", icon: "layout-dashboard", soloAdmin: true },
  { name: "Gestión Productos", path: "/admin/productos", icon: "store", soloAdmin: true },
  { name: "Gestión Pedidos", path: "/admin/pedidos", icon: "credit-card", soloAdmin: true },
  { name: "Liquidaciones", path: "/admin/liquidaciones", icon: "credit-card", soloAdmin: true },
  {
    name: "Panel SaaS",
    icon: "shield",
    soloAdmin: true,
    subItems: [
      { name: "Métricas & Estado", path: "/saas", icon: "shield" },
      { name: "Planes & Precios", path: "/planes", icon: "credit-card" },
      { name: "Usuarios del Tenant", path: "/usuarios-saas", icon: "users" },
    ],
  },
  { name: "Tienda", path: "/tienda", icon: "shopping-bag" },
  { name: "Catálogo", path: "/catalogo", icon: "store" },
  { name: "Mis Pedidos", path: "/mis-pedidos", icon: "credit-card" },
  { name: "Mi Red MLM", path: "/mi-red", icon: "users", soloAdmin: true },
  { name: "Carrito", path: "/carrito", icon: "shopping-cart" },
];

const SidebarContent = ({
  isCollapsed,
  onNavigate,
  onClose,
}: SidebarContentProps) => {
  const { user } = useAuth();
  const { pathname } = useLocation();
  const [saasOpen, setSaasOpen] = useState(
    pathname === "/saas" || pathname === "/planes" || pathname === "/usuarios-saas"
  );

  const itemsPermitidos = navMenu.filter(
    (item) => !item.soloAdmin || user?.rol === "admin"
  );

  return (
    <div className="flex h-full flex-col bg-slate-900 text-white">
      {/* Marca */}
      <div className="flex h-16 shrink-0 items-center justify-between gap-2 border-b border-slate-800 px-4">
        <div
          className={cn(
            "flex items-center gap-3",
            isCollapsed && "justify-center"
          )}
        >
          <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-indigo-600 font-bold text-white shadow-md">
            MC
          </div>
          {!isCollapsed && (
            <span className="truncate text-lg font-bold tracking-tight">
              MultiCatálogo
            </span>
          ) && (
              <span className="flex-col"> Yandris Riviera</span>
            )
          }
        </div>

        {onClose && (
          <SheetClose asChild>
            <Button
              variant="ghost"
              size="icon"
              className="min-h-11 min-w-11 shrink-0 text-slate-400 hover:bg-slate-800 hover:text-white"
              aria-label="Cerrar menú"
            >
              <Icon name="x" className="size-5" />
            </Button>
          </SheetClose>
        )}
      </div>

      {/* Navegación */}
      <nav className="min-h-0 flex-1 space-y-1.5 overflow-y-auto p-3">
        {itemsPermitidos.map((item) => {
          // Caso con submenú (ej: Panel SaaS)
          if (item.subItems) {
            const isAnySubActive = item.subItems.some((sub) => pathname === sub.path);

            if (isCollapsed) {
              return (
                <Tooltip key={item.name}>
                  <TooltipTrigger asChild>
                    <Link
                      to={item.subItems[0].path}
                      onClick={onNavigate}
                      className={cn(
                        "flex min-h-11 items-center justify-center rounded-lg p-3 transition-colors",
                        isAnySubActive
                          ? "bg-indigo-600/30 text-indigo-300 font-medium"
                          : "text-slate-300 hover:bg-slate-800 hover:text-white"
                      )}
                      aria-label={item.name}
                    >
                      <Icon name={item.icon} size={22} />
                    </Link>
                  </TooltipTrigger>
                  <TooltipContent side="right">
                    <p className="font-semibold">{item.name}</p>
                    <div className="mt-1 flex flex-col gap-1 text-xs">
                      {item.subItems.map((sub) => (
                        <span key={sub.path}>{sub.name}</span>
                      ))}
                    </div>
                  </TooltipContent>
                </Tooltip>
              );
            }

            return (
              <div key={item.name} className="space-y-1">
                <button
                  type="button"
                  onClick={() => setSaasOpen(!saasOpen)}
                  className={cn(
                    "flex w-full min-h-11 items-center justify-between gap-3 rounded-lg p-3 transition-colors text-sm",
                    isAnySubActive
                      ? "text-indigo-300 font-semibold bg-slate-800/60"
                      : "text-slate-300 hover:bg-slate-800 hover:text-white"
                  )}
                >
                  <div className="flex items-center gap-3">
                    <Icon name={item.icon} size={20} />
                    <span>{item.name}</span>
                  </div>
                  <Icon
                    name={saasOpen ? "chevron-down" : "chevron-right"}
                    size={16}
                    className="text-slate-400"
                  />
                </button>

                {saasOpen && (
                  <div className="ml-5 pl-3 border-l border-slate-700/60 space-y-1">
                    {item.subItems.map((sub) => {
                      const isSubActive = pathname === sub.path;
                      return (
                        <Link
                          key={sub.path}
                          to={sub.path}
                          onClick={onNavigate}
                          className={cn(
                            "flex min-h-9 items-center gap-2.5 rounded-md px-3 py-1.5 text-xs transition-colors",
                            isSubActive
                              ? "bg-indigo-600 font-medium text-white shadow-sm"
                              : "text-slate-400 hover:bg-slate-800 hover:text-white"
                          )}
                        >
                          <Icon name={sub.icon} size={16} />
                          <span>{sub.name}</span>
                        </Link>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          }

          // Ítems estándar de enlace directo
          const itemPath = item.path || "/";
          const isActive =
            itemPath === "/"
              ? pathname === "/"
              : pathname.startsWith(itemPath);

          const link = (
            <Link
              key={itemPath}
              to={itemPath}
              onClick={onNavigate}
              aria-current={isActive ? "page" : undefined}
              aria-label={isCollapsed ? item.name : undefined}
              className={cn(
                "flex min-h-11 items-center gap-3 rounded-lg p-3 transition-colors text-sm",
                "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-400",
                isActive
                  ? "bg-indigo-600 font-medium text-white shadow-sm"
                  : "text-slate-300 hover:bg-slate-800 hover:text-white",
                isCollapsed && "justify-center px-0"
              )}
            >
              <span className="flex size-6 shrink-0 items-center justify-center">
                <Icon name={item.icon} size={22} />
              </span>
              {!isCollapsed && <span className="truncate">{item.name}</span>}
            </Link>
          );

          if (isCollapsed) {
            return (
              <Tooltip key={itemPath}>
                <TooltipTrigger asChild>{link}</TooltipTrigger>
                <TooltipContent side="right">{item.name}</TooltipContent>
              </Tooltip>
            );
          }

          return link;
        })}
      </nav>

      {/* Footer de rol */}
      <div className="shrink-0 border-t border-slate-800 p-3">
        <div
          className={cn(
            "flex items-center gap-3 rounded-lg bg-slate-800/60 p-3 text-xs",
            isCollapsed && "justify-center px-0"
          )}
        >
          <div className="flex size-2 shrink-0 rounded-full bg-emerald-400" />
          {!isCollapsed && (
            <div className="min-w-0 flex-1">
              <p className="truncate font-medium text-slate-200">
                {user?.email ?? "Usuario"}
              </p>
              <p className="truncate uppercase text-slate-400">
                Rol: {user?.rol ?? "cliente"}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export const Sidebar = ({
  isCollapsed,
  mobileOpen = false,
  onCloseMobile,
}: SidebarProps) => {
  return (
    <>
      {/* Panel de escritorio */}
      <aside
        id="app-sidebar"
        aria-label="Navegación principal"
        className={cn(
          "hidden shrink-0 transition-all duration-300 ease-in-out md:block",
          isCollapsed ? "w-16" : "w-64"
        )}
      >
        <SidebarContent isCollapsed={isCollapsed} />
      </aside>

      {/* Drawer móvil con ui/sheet */}
      <Sheet open={mobileOpen} onOpenChange={(open) => !open && onCloseMobile?.()}>
        <SheetContent
          side="left"
          className="w-72 p-0 border-r border-slate-800 bg-slate-900"
          aria-describedby="mobile-nav-description"
        >
          <SheetHeader className="sr-only">
            <SheetTitle>Navegación principal</SheetTitle>
            <SheetDescription id="mobile-nav-description">
              Menú de navegación para acceder a las secciones del catálogo,
              tienda, carrito y administración.
            </SheetDescription>
          </SheetHeader>
          <SidebarContent
            isCollapsed={false}
            onNavigate={onCloseMobile}
            onClose={onCloseMobile}
          />
        </SheetContent>
      </Sheet>
    </>
  );
};

export default Sidebar;
