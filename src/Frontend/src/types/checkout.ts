/**
 * Contrato del checkout entre el frontend y el Backend.
 *
 * Vive aca y no en `services/api.ts` porque lo necesita tambien la cola de
 * reintentos (`lib/checkoutRetry.ts`), que guarda el payload en
 * `localStorage`: si el tipo estuviera junto al cliente HTTP, ese archivo
 * importaria a `api.ts` y `api.ts` lo importaria a el.
 */

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