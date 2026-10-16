/**
 * Tipos del contexto de autenticación.
 * Separados del Provider para satisfacer la regla
 * `react-refresh/only-export-components` sin necesidad de
 * `eslint-disable`.
 */
export type Rol = "admin" | "cliente";

export interface Usuario {
  email: string;
  rol: Rol;
}
