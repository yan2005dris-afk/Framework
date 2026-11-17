import type { Usuario } from "@/Context/auth-types";

/**
 * Contexto que el router recibe desde `main.tsx`.
 *
 * TanStack Router resuelve los guards en `beforeLoad`, que corre ANTES de
 * montar el componente y no puede usar hooks. Por eso el estado de auth se
 * pasa como contexto: los guards lo leen de ahi y redirigen sin parpadeo.
 */
export interface RouterContext {
  auth: {
    user: Usuario | null;
    isAuthenticated: boolean;
    loading: boolean;
  };
}
