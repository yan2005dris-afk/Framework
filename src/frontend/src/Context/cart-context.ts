import { createContext } from "react";
import type { Producto } from "../data/productos";

/**
 * Ítem del carrito: Producto + cantidad.
 */
export interface CartItem extends Producto {
  cantidad: number;
}

/**
 * Tipo del contexto del carrito.
 * Separado del Provider para satisfacer `react-refresh/only-export-components`.
 */
export interface CartContextType {
  cart: CartItem[];
  addToCart: (producto: Producto) => void;
  removeFromCart: (id: number) => void;
  incrementQuantity: (id: number) => void;
  decrementQuantity: (id: number) => void;
  clearCart: () => void;
  totalItems: number;
  totalPrice: number;
}

export const CartContext = createContext<CartContextType | undefined>(undefined);
