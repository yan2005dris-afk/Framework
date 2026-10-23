import { apiFetch, ApiError } from "./http";
import type { Producto } from "../data/productos";
import {
  guardarProductosOffline,
  obtenerProductosOffline,
  obtenerProductoPorIdOffline,
} from "../lib/offlineDb";

/**
 * Obtiene la lista completa de productos desde el backend.
 * Si la red o el backend fallan (offline/503/red), recupera desde IndexedDB.
 * Cuando la red responde exitosamente, actualiza IndexedDB en segundo plano.
 */
export const getProductos = async (): Promise<Producto[]> => {
  try {
    const productos = await apiFetch<Producto[]>("/productos");
    // Guardar en IndexedDB para disponibilidad offline
    void guardarProductosOffline(productos);
    return productos;
  } catch (err) {
    const offlineItems = await obtenerProductosOffline();
    if (offlineItems.length > 0) {
      return offlineItems;
    }
    throw err;
  }
};

/**
 * Obtiene un producto individual por su identificador numérico.
 * Si falla la red, intenta resolverlo desde IndexedDB.
 */
export const getProductoById = async (id: number): Promise<Producto> => {
  try {
    const producto = await apiFetch<Producto>(`/productos/${id}`);
    void guardarProductosOffline([producto]);
    return producto;
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      throw new ApiError(404, `Producto con ID ${id} no encontrado`);
    }
    const cached = await obtenerProductoPorIdOffline(id);
    if (cached) {
      return cached;
    }
    throw err;
  }
};
