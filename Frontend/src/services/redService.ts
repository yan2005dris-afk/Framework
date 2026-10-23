import { apiFetch } from "./http";
import type { Referido } from "../data/red";
import { guardarRedOffline, obtenerRedOffline } from "../lib/offlineDb";

/**
 * Obtiene el árbol jerárquico de la red multinivel desde el backend.
 * Requiere cookie httpOnly válida (responde 401 sin sesión).
 * Si la red o el backend fallan, recupera la última copia guardada en IndexedDB.
 */
export const getRed = async (
  onUnauthorized?: () => void
): Promise<Referido> => {
  try {
    const data = await apiFetch<Referido>("/red", { onUnauthorized });
    void guardarRedOffline(data);
    return data;
  } catch (err) {
    const cached = await obtenerRedOffline();
    if (cached) {
      return cached;
    }
    throw err;
  }
};
