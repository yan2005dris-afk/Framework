import { useState } from "react";
import { Link, useRouterState } from "@tanstack/react-router";
import {
  LayoutDashboard,
  Users,
  CreditCard,
  Building2,
  ChevronLeft,
  ChevronRight,
  ArrowLeft,
} from "lucide-react";
import { Button } from "@/components/ui/button";

interface SaasSidebarProps {
  tenantId: number;
}

export const SaasSidebar = ({ tenantId }: SaasSidebarProps) => {
  const [collapsed, setCollapsed] = useState(false);
  const routerState = useRouterState();
  const currentPath = routerState.location.pathname;

  const menuItems = [
    {
      title: "Resumen SaaS",
      href: "/saas",
      icon: LayoutDashboard,
      active: currentPath === "/saas" || currentPath === "/saas/",
    },
    {
      title: "Planes & Facturación",
      href: "/saas/planes",
      icon: CreditCard,
      active: currentPath.startsWith("/saas/planes"),
    },
    {
      title: "Equipo y Usuarios",
      href: "/saas/usuarios",
      icon: Users,
      active: currentPath.startsWith("/saas/usuarios"),
    },
  ];

  return (
    <aside
      className={`relative flex flex-col border-r bg-card transition-all duration-300 ${
        collapsed ? "w-16" : "w-64"
      }`}
    >
      {/* Header Sidebar */}
      <div className="flex h-16 items-center justify-between px-4 border-b">
        {!collapsed && (
          <div className="flex items-center gap-2 overflow-hidden">
            <div className="p-1.5 rounded-lg bg-indigo-600 text-white flex items-center justify-center">
              <Building2 className="h-5 w-5" />
            </div>
            <div className="truncate">
              <h2 className="text-sm font-semibold leading-tight">Admin SaaS</h2>
              <span className="text-xs text-muted-foreground">Tenant #{tenantId}</span>
            </div>
          </div>
        )}
        {collapsed && (
          <div className="mx-auto p-1.5 rounded-lg bg-indigo-600 text-white">
            <Building2 className="h-5 w-5" />
          </div>
        )}
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 ml-auto"
          onClick={() => setCollapsed(!collapsed)}
        >
          {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </Button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 p-2">
        {menuItems.map((item) => {
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              to={item.href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors ${
                item.active
                  ? "bg-primary text-primary-foreground shadow-sm"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
              } ${collapsed ? "justify-center px-0" : ""}`}
              title={collapsed ? item.title : undefined}
            >
              <Icon className="h-5 w-5 shrink-0" />
              {!collapsed && <span>{item.title}</span>}
            </Link>
          );
        })}
      </nav>

      {/* Footer Return */}
      <div className="p-2 border-t mt-auto">
        <Link
          to="/"
          className={`flex items-center gap-3 px-3 py-2 rounded-md text-xs text-muted-foreground hover:bg-accent hover:text-accent-foreground transition ${
            collapsed ? "justify-center px-0" : ""
          }`}
          title={collapsed ? "Volver a la tienda" : undefined}
        >
          <ArrowLeft className="h-4 w-4 shrink-0" />
          {!collapsed && <span>Volver a Tienda</span>}
        </Link>
      </div>
    </aside>
  );
};
