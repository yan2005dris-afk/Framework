import { API_BASE_URL } from './api';
import type { Producto } from '../data/productos';

/**
 * Obtiene la lista completa de productos desde el backend.
 * @returns {Promise<Producto[]>} Promesa con la lista de productos.
 */
export const getProductos = async (): Promise<Producto[]> => {
  const response = await fetch(`${API_BASE_URL}/productos`, {
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error('Error al obtener el catálogo de productos');
  }

  return response.json();
};

/**
 * Obtiene un producto individual por su identificador numérico.
 * @param {number} id - Identificador del producto.
 * @returns {Promise<Producto>} Promesa con el producto encontrado.
 */
export const getProductoById = async (id: number): Promise<Producto> => {
  const response = await fetch(`${API_BASE_URL}/productos/${id}`, {
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`Producto con ID ${id} no encontrado`);
  }

  return response.json();
};
