import { createContext } from "react";
import type { Usuario } from "./auth-types";

/**
 * Contexto de autenticación.
 * Separado del Provider para satisfacer `react-refresh/only-export-components`.
 *
 * El provider concreta `AuthContextType` (incluye `loading` para el spinner
 * de hidratación); useAuth lo consume via useContext(AuthContext).
 */
export interface AuthContextType {
  isAuthenticated: boolean;
  user: Usuario | null;
  loading: boolean; // true mientras se hidrata la sesion contra /api/me
  login: (usuario: Usuario) => void;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);
