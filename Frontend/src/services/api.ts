// src/services/api.ts

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000/api';

export interface LoginResponse {
  token: string;
  email: string;
}

export interface Producto {
  id: number;
  nombre: string;
  precio: number;
  img: string;
}

export const api = {
  async login(email: string, password: string): Promise<LoginResponse> {
    const response = await fetch(`${API_BASE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || 'Error al iniciar sesión');
    }

    return data;
  },

  async getProductos(): Promise<Producto[]> {
    const response = await fetch(`${API_BASE_URL}/productos`, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
      },
    });

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new Error(data.error || 'Error al obtener productos');
    }

    return response.json();
  },
};
