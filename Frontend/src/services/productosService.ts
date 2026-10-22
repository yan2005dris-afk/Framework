import { apiFetch, ApiError } from "./http";
import type { Producto } from "../data/productos";

/**
 * Obtiene la lista completa de productos desde el backend.
 * Usa `apiFetch` (cookie httpOnly + Authorization opcional).
 */
export const getProductos = async (): Promise<Producto[]> => {
  return apiFetch<Producto[]>("/productos");
};

/**
 * Obtiene un producto individual por su identificador numérico.
 * Lanza `ApiError` con status 404 si no existe.
 */
export const getProductoById = async (id: number): Promise<Producto> => {
  try {
    return await apiFetch<Producto>(`/productos/${id}`);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      throw new ApiError(404, `Producto con ID ${id} no encontrado`);
    }
    throw err;
  }
};
