import { API_BASE_URL } from './api';
import type { Referido } from '../data/red';

/**
 * Obtiene el árbol jerárquico de la red multinivel desde el backend.
 * @returns {Promise<Referido>} Promesa con la raíz del árbol de referidos.
 */
export const getRed = async (): Promise<Referido> => {
  const response = await fetch(`${API_BASE_URL}/red`, {
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error('Error al obtener la red multinivel');
  }

  return response.json();
};
