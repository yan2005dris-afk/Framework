import type { Producto } from "../data/productos";
import { ApiError, apiFetch } from "./http";

export interface LoginResponse {
  email: string;
  rol: "admin" | "cliente";
  expires: number;
}

export interface MeResponse {
  id: number;
  email: string;
  rol: "admin" | "cliente";
}

export interface CheckoutItemDTO {
  productoId: number;
  cantidad: number;
  precio: number;
}

export interface CheckoutRequestDTO {
  nombreCliente: string;
  emailCliente: string;
  direccion?: string;
  ciudad?: string;
  telefono?: string;
  metodoPago: string;
  items: CheckoutItemDTO[];
}

export interface CheckoutResponseDTO {
  pedidoId: number;
  numero: string;
  total: number;
  mensaje: string;
}

/**
 * Cliente de API. Toda llamada pasa por `apiFetch` con credenciales para
 * que el browser adjunte automáticamente la cookie `auth_token`.
 *
 * El token YA NO se devuelve en el body: viaja únicamente en la cookie
 * httpOnly configurada por el backend. Eso elimina la posibilidad de
 * robo por XSS.
 */
export const api = {
  /**
   * POST /api/login
   * Devuelve {email, rol, expires}. La cookie httpOnly llega en el header
   * Set-Cookie del response.
   */
  login(email: string, password: string, onUnauthorized?: () => void): Promise<LoginResponse> {
    return apiFetch<LoginResponse>("/login", {
      method: "POST",
      json: { email, password },
      onUnauthorized,
    });
  },

  /**
   * POST /api/logout
   * Limpia la cookie httpOnly en el backend. Idempotente: si no había
   * cookie, sigue respondiendo 200 OK.
   */
  logout(onUnauthorized?: () => void): Promise<{ status: string }> {
    return apiFetch<{ status: string }>("/logout", {
      method: "POST",
      onUnauthorized,
    });
  },

  /**
   * GET /api/me
   * Endpoint protegido que devuelve la identidad leída desde la cookie.
   * El frontend lo llama al montar la app para hidratar la sesión.
   * Si la cookie no está o expiró, lanza ApiError con status 401.
   */
  me(onUnauthorized?: () => void): Promise<MeResponse> {
    return apiFetch<MeResponse>("/me", {
      method: "GET",
      onUnauthorized,
    });
  },

  /**
   * GET /api/productos
   * Catálogo completo desde el backend con Cache-Aside.
   */
  getProductos(): Promise<Producto[]> {
    return apiFetch<Producto[]>("/productos", { method: "GET" });
  },

  /**
   * GET /api/red
   * Árbol multinivel. Requiere cookie httpOnly válida.
   */
  getRed(onUnauthorized?: () => void): Promise<unknown> {
    return apiFetch<unknown>("/red", { method: "GET", onUnauthorized });
  },

  /**
   * POST /api/checkout
   * Checkout transaccional ACID. La identidad del usuario se toma del JWT
   * en la cookie; el campo emailCliente del body es sobrescrito por el
   * backend con el email del token antes de la transacción.
   */
  checkout(
    payload: CheckoutRequestDTO,
    onUnauthorized?: () => void
  ): Promise<CheckoutResponseDTO> {
    return apiFetch<CheckoutResponseDTO>("/checkout", {
      method: "POST",
      json: payload,
      onUnauthorized,
    });
  },
};

export { ApiError };
