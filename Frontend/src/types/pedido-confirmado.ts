/**
 * Pedido que Checkout pasa a Confirmacion por el state del historial.
 *
 * Vive en su propio modulo para que ambos componentes compartan la forma exacta
 * sin importarse entre si.
 */
export interface PedidoConfirmado {
  numero: string;
  items: { nombre: string; cantidad: number; precio: number }[];
  total: number;
  cliente: { nombre: string; email: string; ciudad: string };
  fecha: string;
}

/**
 * Estado de navegacion que Checkout entrega a Confirmacion. El state del
 * historial se pierde con un refresh, igual que antes con react-router.
 */
export interface PedidoConfirmadoState {
  pedido: PedidoConfirmado;
}