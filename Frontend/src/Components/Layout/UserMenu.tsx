import { useNavigate } from "@tanstack/react-router";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuth } from "../../Context/useAuth";
import { Icon } from "../Share/Icon";

/** Avatar de ejemplo mientras no exista la foto del usuario en la API. */
const AVATAR_URL =
  "https://fastly.picsum.photos/id/64/4326/2884.jpg?hmac=9_SzX666YRpR_fOyYStXpfSiJ_edO3ghlSRnH2w09Kg";

/**
 * Menú de usuario del header.
 *
 * Antes vivía dentro de `Navbar` y se desplegaba con un `div` `group-hover`,
 * que no era alcanzable por teclado ni se cerraba con Escape. Ahora usa
 * `ui/dropdown-menu`, el patrón de `nav-user.tsx` de shadcn-admin: trigger con
 * `Avatar`, contenido alineado a la derecha y una acción de sesión.
 */
const UserMenu = () => {
  const { logout, user } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate({ to: "/login" });
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="min-h-11 min-w-11 rounded-full sm:min-h-9 sm:min-w-9"
          aria-label="Abrir menú de usuario"
        >
          <Avatar className="size-8 border border-slate-300 sm:size-9">
            <AvatarImage src={AVATAR_URL} alt="Avatar del usuario" />
            <AvatarFallback>
              {user?.email?.charAt(0).toUpperCase() ?? "?"}
            </AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel className="flex flex-col gap-0.5 py-2">
          <span className="text-xs font-medium text-muted-foreground">
            Conectado como:
          </span>
          <span className="truncate text-sm font-semibold text-foreground">
            {user?.email}
          </span>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {/* min-h-11: la auditoría mobile exige 44px de alto en acciones. */}
        <DropdownMenuItem
          variant="destructive"
          className="min-h-11"
          onSelect={handleLogout}
        >
          <Icon name="log-out" className="size-4" />
          Cerrar Sesión
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

export default UserMenu;
