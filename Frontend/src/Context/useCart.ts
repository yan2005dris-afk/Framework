import { useContext } from "react";
import { CartContext, type CartContextType } from "./cart-context";

/**
 * Hook para acceder al contexto del carrito.
 * Separado del Provider y del Context para satisfacer
 * `react-refresh/only-export-components`.
 */
export const useCart = (): CartContextType => {
  const ctx = useContext(CartContext);
  if (!ctx) {
    throw new Error("useCart debe ser usado dentro de un CartProvider");
  }
  return ctx;
};
