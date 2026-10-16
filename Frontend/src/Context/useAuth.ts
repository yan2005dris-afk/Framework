import { useContext } from "react";
import { AuthContext, type AuthContextType } from "./AuthProvider";

/**
 * Hook para acceder al contexto de autenticación.
 * Separado del Provider para satisfacer `react-refresh/only-export-components`
 * sin necesidad de `eslint-disable`.
 */
export const useAuth = (): AuthContextType => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  }
  return ctx;
};
