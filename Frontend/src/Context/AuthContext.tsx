/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useState, type ReactNode } from "react";

/**
 * Roles disponibles en el sistema.
 */
export type Rol = "admin" | "cliente";

/**
 * Información del usuario autenticado.
 */
export interface Usuario {
  email: string;
  rol: Rol;
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: Usuario | null;
  token: string | null;
  login: (usuario: Usuario, token?: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

/**
 * Hook para acceder al contexto de autenticación.
 * @returns {AuthContextType} Contexto de autenticación.
 */
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

/**
 * Proveedor del contexto de autenticación.
 */
export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem("token")
  );
  const [user, setUser] = useState<Usuario | null>(() => {
    const email = localStorage.getItem("userEmail");
    const rol = localStorage.getItem("userRol") as Rol | null;
    if (email && (rol === "admin" || rol === "cliente")) {
      return { email, rol };
    }
    return null;
  });
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() =>
    Boolean(localStorage.getItem("token") || localStorage.getItem("userEmail"))
  );

  const login = (usuario: Usuario, authToken: string = "fake-jwt-token") => {
    setIsAuthenticated(true);
    setUser(usuario);
    setToken(authToken);
    localStorage.setItem("token", authToken);
    localStorage.setItem("userEmail", usuario.email);
    localStorage.setItem("userRol", usuario.rol);
  };

  const logout = () => {
    setIsAuthenticated(false);
    setUser(null);
    setToken(null);
    localStorage.removeItem("token");
    localStorage.removeItem("userEmail");
    localStorage.removeItem("userRol");
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, user, token, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
};