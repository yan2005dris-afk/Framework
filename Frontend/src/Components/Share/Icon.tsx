import {
  LayoutDashboard,
  LogOut,
  Menu,
  PanelLeftClose,
  PanelLeftOpen,
  ShoppingBag,
  ShoppingCart,
  Store,
  Users,
  X,
  type LucideIcon,
} from "lucide-react";

/**
 * Mapa estatico de los iconos que la app usa.
 *
 * Antes este componente usaba `DynamicIcon` de `lucide-react/dynamic`, que
 * resuelve el icono con un import dinamico por nombre. El costo: el build
 * empaquetaba los ~1900 iconos de lucide como chunks sueltos (1932 archivos en
 * dist/assets). Con el service worker de U4-T3, eso son 1932 entradas de
 * precache, e instalar la PWA significaba descargar casi dos mil requests para
 * usar diez iconos.
 *
 * Con el mapa estatico solo entra lo que se usa. Agregar un icono nuevo es
 * agregar su entrada aca, y TypeScript marca cualquier `name` que no exista.
 */
const iconos = {
  "layout-dashboard": LayoutDashboard,
  "log-out": LogOut,
  menu: Menu,
  "panel-left-close": PanelLeftClose,
  "panel-left-open": PanelLeftOpen,
  "shopping-bag": ShoppingBag,
  "shopping-cart": ShoppingCart,
  store: Store,
  users: Users,
  x: X,
} satisfies Record<string, LucideIcon>;

export type IconName = keyof typeof iconos;

interface IconProps {
  name: IconName;
  size?: number;
  className?: string;
}


export const Icon = ({ name, size = 20, className }: IconProps) => {
  const Componente = iconos[name];
  // aria-hidden: todos los iconos de la app van junto a texto o dentro de un
  // boton con etiqueta, asi que el lector de pantalla no necesita leerlos.
  return <Componente size={size} className={className} aria-hidden="true" />;
};