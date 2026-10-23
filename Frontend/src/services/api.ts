import type { Producto } from "../data/productos";
import type { CheckoutRequestDTO, CheckoutResponseDTO } from "@/types/checkout";
import { ApiError, NetworkError, apiFetch } from "./http";
import {
  guardarIntento,
  nuevaIdempotencyKey,
  type IntentoPendiente,
} from "@/lib/checkoutRetry";

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
   *
   * U4-T3: cada intento lleva su propia `Idempotency-Key`. Si la request falla
   * por RED (no hubo respuesta), el intento queda en la cola de reintentos con
   * esa misma key: el Backend devuelve el pedido original en vez de crear un
   * segundo, se haya creado o no el primero.
   */
  checkout(
    payload: CheckoutRequestDTO,
    onUnauthorized?: () => void
  ): Promise<CheckoutResponseDTO> {
    const idempotencyKey = nuevaIdempotencyKey();
    return apiFetch<CheckoutResponseDTO>("/checkout", {
      method: "POST",
      json: payload,
      onUnauthorized,
      headers: { "Idempotency-Key": idempotencyKey },
    }).catch((error: unknown) => {
      // Solo el fallo de RED deja el pedido en un estado indeterminado. Un 4xx
      // o un 5xx son respuestas del servidor: reenviarlos no arregla nada y
      // multiplicaría el problema.
      if (error instanceof NetworkError) {
        guardarIntento({ idempotencyKey, payload, guardadoAt: new Date().toISOString() });
      }
      throw error;
    });
  },
};

/**
 * Reenvía un intento guardado con SU key (U4-T3, S9). No genera una nueva a
 * propósito: esa es justamente la garantía de que no se duplique el pedido.
 */
export const reintentarCheckout = (
  intento: IntentoPendiente
): Promise<CheckoutResponseDTO> =>
  apiFetch<CheckoutResponseDTO>("/checkout", {
    method: "POST",
    json: intento.payload,
    headers: { "Idempotency-Key": intento.idempotencyKey },
  });

export { ApiError };
