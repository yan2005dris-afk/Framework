import { Link } from "@tanstack/react-router";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useCart } from "../../Context/useCart";
import { useAuth } from "../../Context/useAuth";
import { Icon } from "../Share/Icon";
import UserMenu from "./UserMenu";

interface NavbarProps {
  /** Escritorio: colapsa/expande el riel del sidebar. */
  onToggleSidebar: () => void;
  /** Móvil: abre el drawer (`ui/sheet`). */
  onOpenMobileNav: () => void;
  isCollapsed?: boolean;
}

/**
 * Barra superior de navegación con toggle del sidebar, badge de rol, contador de
 * carrito y menú de usuario.
 *
 * Sigue `app-header.tsx` de shadcn-admin (acciones a la derecha, trigger a la
 * izquierda) pero con los mismos datos que antes: rol, carrito y usuario.
 */
const Navbar = ({
  onToggleSidebar,
  onOpenMobileNav,
  isCollapsed = false,
}: NavbarProps) => {
  const { totalItems } = useCart();
  const { user } = useAuth();

  const titulo =
    user?.rol === "admin" ? "Panel de Administración" : "Tienda MultiCatálogo";

  return (
    <header className="z-20 flex h-16 shrink-0 items-center justify-between gap-2 border-b border-slate-200 bg-white px-4 sm:px-6 md:px-8">
      <div className="flex min-w-0 items-center gap-1 sm:gap-2">
        {/* Móvil: abre el drawer off-canvas. */}
        <Button
          variant="ghost"
          size="icon"
          className="min-h-11 min-w-11 shrink-0 text-slate-600 md:hidden"
          onClick={onOpenMobileNav}
          aria-label="Abrir menú de navegación"
          aria-haspopup="dialog"
        >
          <Icon name="menu" className="size-5" />
        </Button>

        {/* Escritorio: colapsa/expande el riel de iconos. */}
        <Button
          variant="ghost"
          size="icon"
          className="hidden min-h-11 min-w-11 shrink-0 text-slate-600 md:inline-flex"
          onClick={onToggleSidebar}
          aria-expanded={!isCollapsed}
          aria-controls="app-sidebar"
          aria-label={
            isCollapsed ? "Expandir menú lateral" : "Colapsar menú lateral"
          }
        >
          <Icon
            name={isCollapsed ? "panel-left-open" : "panel-left-close"}
            className="size-5"
          />
        </Button>

        <h2 className="truncate text-base font-semibold text-slate-800 sm:text-lg">
          {titulo}
        </h2>
      </div>

      {/* Acciones: carrito, rol y menú de usuario. */}
      <div className="flex shrink-0 items-center gap-1 sm:gap-3">
        <Button
          asChild
          variant="ghost"
          size="icon"
          className="relative min-h-11 min-w-11 text-slate-700 sm:min-h-9 sm:min-w-9"
        >
          <Link
            to="/carrito"
            title="Ver Carrito"
            aria-label={
              totalItems > 0
                ? `Ver carrito, ${totalItems} artículos`
                : "Ver carrito"
            }
          >
            <Icon name="shopping-cart" className="size-5" />
            {totalItems > 0 && (
              <Badge className="absolute -top-0.5 -right-0.5 size-5 rounded-full bg-indigo-600 p-0 text-[10px] font-bold text-white shadow-sm">
                {totalItems}
              </Badge>
            )}
          </Link>
        </Button>

        {user?.email && (
          <span className="hidden max-w-[150px] truncate text-xs text-slate-500 md:inline sm:text-sm">
            {user.email}
          </span>
        )}

        {user?.rol && (
          <Badge
            className={cn(
              "hidden font-bold uppercase sm:inline-flex",
              user.rol === "admin"
                ? "bg-amber-100 text-amber-800"
                : "bg-indigo-100 text-indigo-800"
            )}
          >
            {user.rol}
          </Badge>
        )}

        <UserMenu />
      </div>
    </header>
  );
};

export default Navbar;
