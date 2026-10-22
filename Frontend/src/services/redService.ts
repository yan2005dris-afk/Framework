import { apiFetch } from "./http";
import type { Referido } from "../data/red";

/**
 * Obtiene el árbol jerárquico de la red multinivel desde el backend.
 * Requiere cookie httpOnly válida (responde 401 sin sesión).
 */
export const getRed = async (
  onUnauthorized?: () => void
): Promise<Referido> => {
  return apiFetch<Referido>("/red", { onUnauthorized });
};
