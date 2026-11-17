import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { api } from "../services/api";
import { AuthContext, type AuthContextType } from "./auth-context";
import type { Usuario } from "./auth-types";

const SESSION_STORAGE_KEY = "mc_session";

/**
 * AuthProvider:
 *  - Al montar, lee la sesión persistida (`mc_session`) y luego confirma
 *    contra el backend (`GET /api/me`). Si /api/me responde 200, mantiene
 *    la sesión; si 401, la limpia.
 *  - login(usuario) la setea en state y en localStorage.
 *  - logout() llama a `POST /api/logout` (limpia cookie httpOnly) y limpia
 *    state + localStorage.
 *
 * No guarda token en ningún storage del cliente: el token vive únicamente
 * en la cookie httpOnly del backend.
 */
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<Usuario | null>(() => {
    try {
      const raw = localStorage.getItem(SESSION_STORAGE_KEY);
      if (!raw) return null;
      const parsed = JSON.parse(raw) as Usuario;
      if (parsed.email && (parsed.rol === "admin" || parsed.rol === "cliente")) {
        return parsed;
      }
    } catch {
      // localStorage corrupto o inaccesible - ignorar.
    }
    return null;
  });
  const [loading, setLoading] = useState(true);
  const unauthorizedFiredRef = useRef(false);

  // Manejador único de 401 - evita duplicar lógica en cada api.xx() call.
  const handleUnauthorized = useCallback(() => {
    if (unauthorizedFiredRef.current) return;
    unauthorizedFiredRef.current = true;
    setUser(null);
    localStorage.removeItem(SESSION_STORAGE_KEY);
    // Reactivar el flag tras un tick para que un próximo 401 vuelva a
    // disparar (p.ej. tras login exitoso).
    queueMicrotask(() => {
      unauthorizedFiredRef.current = false;
    });
  }, []);

  // Hidratación: confirmar la sesión contra el backend cuando haya conexión.
  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const me = await api.me(handleUnauthorized);
        if (active) {
          setUser({ email: me.email, rol: me.rol });
          localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify({ email: me.email, rol: me.rol }));
        }
      } catch {
        // Si falló por red o servidor caído (ej: offline/503), NO limpiamos la sesión
        // preservada en localStorage para permitir acceso offline al catálogo y tienda.
        // Si fue 401 explícito, handleUnauthorized ya limpió el estado y el storage.
      } finally {
        if (active) setLoading(false);
      }
    })();
    return () => {
      active = false;
    };
  }, [handleUnauthorized]);

  const login = (usuario: Usuario) => {
    setUser(usuario);
    localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(usuario));
  };

  const logout = async () => {
    try {
      await api.logout();
    } catch {
      // Aun si la llamada falla, limpiamos local.
    }
    setUser(null);
    localStorage.removeItem(SESSION_STORAGE_KEY);
  };

  const value: AuthContextType = {
    isAuthenticated: user !== null,
    user,
    loading,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
