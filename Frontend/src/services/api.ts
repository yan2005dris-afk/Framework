import type { Rol } from "../Context/AuthContext";
import type { Producto } from "../data/productos";

export const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

export interface LoginResponse {
  token: string;
  email: string;
  rol: Rol;
}

/**
 * Cliente centralizado para servicios de API.
 */
export const api = {
  /**
   * Realiza la autenticación del usuario mediante credenciales.
   * @param {string} email - Correo del usuario.
   * @param {string} password - Contraseña.
   * @returns {Promise<LoginResponse>} Objeto con token, email y rol.
   */
  async login(email: string, password: string): Promise<LoginResponse> {
    const response = await fetch(`${API_BASE_URL}/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || "Error al iniciar sesión");
    }

    return data;
  },

  /**
   * Obtiene el listado de productos del catálogo.
   * @returns {Promise<Producto[]>} Listado de productos.
   */
  async getProductos(): Promise<Producto[]> {
    const response = await fetch(`${API_BASE_URL}/productos`, {
      method: "GET",
      headers: {
        Accept: "application/json",
      },
    });

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new Error(data.error || "Error al obtener productos");
    }

    return response.json();
  },
};
