import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../../Context/CartContext";
import { useAuth } from "../../Context/AuthContext";
import { Icon } from "../Share/Icon";

interface NavbarProps {
  onToggleSidebar: () => void;
  isCollapsed?: boolean;
}

/**
 * Barra superior de navegación con badge de rol, contador de carrito y menú de usuario.
 */
const Navbar = ({ onToggleSidebar }: NavbarProps) => {
  const { totalItems } = useCart();
  const { logout, user } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 sm:px-6 md:px-8 z-10">
      {/* Botón de Toggle y Título */}
      <div className="flex items-center gap-3">
        <button
          onClick={onToggleSidebar}
          aria-label="Toggle Sidebar"
          className="p-2 text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition-colors cursor-pointer"
        >
          <Icon name="menu" size={22} />
        </button>
        <h2 className="text-slate-800 font-semibold text-base sm:text-lg">
          {user?.rol === "admin"
            ? "Panel de Administración"
            : "Tienda MultiCatálogo"}
        </h2>
      </div>

      {/* Acciones del Navbar: Carrito y Perfil */}
      <div className="flex items-center gap-3 sm:gap-6">
        <Link
          to="/carrito"
          title="Ver Carrito"
          className="relative p-2 text-slate-700 hover:bg-slate-100 rounded-full transition-colors flex items-center justify-center"
        >
          <Icon name="shopping-cart" size={20} />
          {totalItems > 0 && (
            <span className="absolute -top-1 -right-1 bg-indigo-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full shadow">
              {totalItems}
            </span>
          )}
        </Link>

        <div className="flex items-center gap-3">
          {user?.email && (
            <span className="hidden md:inline text-xs sm:text-sm text-slate-500 max-w-[150px] truncate">
              {user.email}
            </span>
          )}

          {user?.rol && (
            <span
              className={`hidden sm:inline-block text-xs font-bold px-2.5 py-0.5 rounded-full uppercase ${
                user.rol === "admin"
                  ? "bg-amber-100 text-amber-800"
                  : "bg-indigo-100 text-indigo-800"
              }`}
            >
              {user.rol}
            </span>
          )}

          {/* Menú de Usuario */}
          <div className="relative group cursor-pointer py-1">
            <div className="w-9 h-9 rounded-full bg-slate-200 overflow-hidden border border-slate-300 flex items-center justify-center">
              <img
                src="https://fastly.picsum.photos/id/64/4326/2884.jpg?hmac=9_SzX666YRpR_fOyYStXpfSiJ_edO3ghlSRnH2w09Kg"
                alt="Avatar"
                className="w-full h-full object-cover"
              />
            </div>

            {/* Menú desplegable */}
            <div className="absolute right-0 top-full mt-1 w-44 bg-white border border-slate-200 rounded-lg shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-150 z-50 p-1">
              <div className="px-3 py-2 border-b border-slate-100 md:hidden">
                <p className="text-xs text-slate-400 font-medium">
                  Conectado como:
                </p>
                <p className="text-xs text-slate-700 font-semibold truncate">
                  {user?.email}
                </p>
              </div>
              <button
                onClick={handleLogout}
                className="w-full text-left px-3 py-2 text-sm text-red-600 font-medium hover:bg-red-50 rounded-md transition-colors flex items-center gap-2"
              >
                <Icon name="log-out" size={16} />
                Cerrar Sesión
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Navbar;
