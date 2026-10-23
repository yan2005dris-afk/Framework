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

interface NavItem {
  name: string;
  path: NavPath;
  icon: IconName;
  soloAdmin?: boolean;
}

/** Rutas del layout protegido. El union mantiene `to={item.path}` tipado por TanStack Router. */
type NavPath = "/" | "/tienda" | "/catalogo" | "/mi-red" | "/carrito" | "/saas";

interface SidebarProps {
  /** Riel angosto (solo iconos) vs panel completo. Afecta solo a escritorio. */
  isCollapsed: boolean;
  /** Estado controlado del drawer móvil. */
  mobileOpen?: boolean;
  onCloseMobile?: () => void;
}

interface SidebarContentProps {
  isCollapsed: boolean;
  /** Cierra el drawer móvil al tocar un ítem de navegación. */
  onNavigate?: () => void;
  /** Solo en el drawer: inyecta el botón de cerrar. */
  onClose?: () => void;
}

const navItems: NavItem[] = [
  { name: "Dashboard", path: "/", icon: "layout-dashboard", soloAdmin: true },
  { name: "Panel SaaS", path: "/saas", icon: "shield", soloAdmin: true },
  { name: "Tienda", path: "/tienda", icon: "shopping-bag" },
  { name: "Catálogo", path: "/catalogo", icon: "store" },
  { name: "Mi Red", path: "/mi-red", icon: "users", soloAdmin: true },
  { name: "Carrito", path: "/carrito", icon: "shopping-cart" },
];

/**
 * Cuerpo del sidebar, compartido entre el panel de escritorio y el drawer móvil.
 *
 * Sigue el patrón de `app-sidebar.tsx` + `nav-group.tsx` de shadcn-admin:
 * header de marca, navegación y footer de rol. Este repo no tiene instalado el
 * primitive `sidebar`, así que el `aside` se arma a mano conservando el lenguaje
 * de color actual: fondo slate, acentos indigo.
 */
const SidebarContent = ({
  isCollapsed,
  onNavigate,
  onClose,
}: SidebarContentProps) => {
  const { user } = useAuth();
  const { pathname } = useLocation();

  const itemsPermitidos = navItems.filter(
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
          )}
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
          const isActive =
            item.path === "/"
              ? pathname === "/"
              : pathname.startsWith(item.path);

          const link = (
            <Link
              key={item.path}
              to={item.path}
              onClick={onNavigate}
              aria-current={isActive ? "page" : undefined}
              // Sin texto visible en el riel: el nombre accesible no puede
              // depender solo del tooltip, que solo se anuncia al abrirlo.
              aria-label={isCollapsed ? item.name : undefined}
              className={cn(
                // min-h-11: la auditoría mobile exige 44px de alto en targets.
                "flex min-h-11 items-center gap-3 rounded-lg p-3 transition-colors",
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

          // En el riel angosto el nombre no cabe: tooltip real en vez de `title`.
          if (!isCollapsed) return link;

          return (
            <Tooltip key={item.path}>
              <TooltipTrigger asChild>{link}</TooltipTrigger>
              <TooltipContent side="right" sideOffset={8}>
                {item.name}
              </TooltipContent>
            </Tooltip>
          );
        })}
      </nav>

      {/* Rol */}
      <div className="shrink-0 border-t border-slate-800 p-3 text-center text-xs text-slate-400">
        {isCollapsed ? (
          <span className="font-semibold uppercase text-indigo-400">
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
    </div>
  );
};

/**
 * Sidebar responsive.
 *
 * Escritorio: `aside` en el flujo, conmutable entre panel completo (16rem) y riel
 * de iconos (5rem) desde el toggle del header. Móvil: `ui/sheet` como drawer
 * off-canvas real, en lugar del overlay a mano y el `-translate-x-full`; así el
 * foco queda atrapado en el drawer y se cierra con Escape, con el overlay o con
 * el botón de cerrar.
 */
const Sidebar = ({
  isCollapsed,
  mobileOpen = false,
  onCloseMobile,
}: SidebarProps) => (
  <>
    <aside
      id="app-sidebar"
      aria-label="Navegación principal"
      className={cn(
        "hidden shrink-0 flex-col bg-slate-900 text-white transition-[width] duration-300 ease-in-out md:flex",
        isCollapsed ? "w-20" : "w-64"
      )}
    >
      <SidebarContent isCollapsed={isCollapsed} />
    </aside>

    <Sheet
      open={mobileOpen}
      onOpenChange={(open) => {
        if (!open) onCloseMobile?.();
      }}
    >
      {/* Se usa el botón de cierre propio de SidebarContent: el de la primitive
          vendida trae un `sr-only` en inglés y queda en 28px de alto. */}
      <SheetContent
        side="left"
        showCloseButton={false}
        className="w-64 gap-0 border-0 bg-slate-900 p-0 text-white data-[side=left]:w-64! data-[side=left]:max-w-64!"
      >
        <SheetHeader className="sr-only">
          <SheetTitle>Menú de navegación</SheetTitle>
          <SheetDescription>
            Secciones disponibles en MultiCatálogo
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

export default Sidebar;
