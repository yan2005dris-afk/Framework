import { useEffect, useState, type ReactNode } from "react";
import type { Producto } from "../data/productos";
import { useAuth } from "./useAuth";
import { CartContext, type CartItem, type CartContextType } from "./cart-context";

const CART_STORAGE_PREFIX = "multicatalogo_carrito_";

/**
 * CartProvider mantiene el carrito en estado local y lo persiste en
 * localStorage bajo una clave por email de usuario, garantizando
 * aislamiento entre cuentas.
 */
export const CartProvider = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  const storageKey = user
    ? `${CART_STORAGE_PREFIX}${user.email}`
    : `${CART_STORAGE_PREFIX}anonimo`;

  const [cart, setCart] = useState<CartItem[]>(() => {
    try {
      const saved = localStorage.getItem(storageKey);
      return saved ? (JSON.parse(saved) as CartItem[]) : [];
    } catch {
      return [];
    }
  });

  useEffect(() => {
    localStorage.setItem(storageKey, JSON.stringify(cart));
  }, [cart, storageKey]);

  /**
   * Añade un producto al carrito o incrementa su cantidad si ya existe.
   * @param {Producto} producto - Producto a agregar.
   */
  const addToCart = (producto: Producto) => {
    setCart((prevCart) => {
      const itemExists = prevCart.find((item) => item.id === producto.id);
      if (itemExists) {
        return prevCart.map((item) =>
          item.id === producto.id
            ? { ...item, cantidad: item.cantidad + 1 }
            : item
        );
      }
      return [...prevCart, { ...producto, cantidad: 1 }];
    });
  };

  /**
   * Elimina un producto del carrito por su ID.
   * @param {number} id - ID del producto a eliminar.
   */
  const removeFromCart = (id: number) => {
    setCart((prevCart) => prevCart.filter((item) => item.id !== id));
  };

  /**
   * Incrementa la cantidad de un ítem en el carrito.
   * @param {number} id - ID del producto.
   */
  const incrementQuantity = (id: number) => {
    setCart((prevCart) =>
      prevCart.map((item) =>
        item.id === id ? { ...item, cantidad: item.cantidad + 1 } : item
      )
    );
  };

  /**
   * Disminuye la cantidad de un ítem en el carrito o lo elimina si llega a cero.
   * @param {number} id - ID del producto.
   */
  const decrementQuantity = (id: number) => {
    setCart((prevCart) =>
      prevCart
        .map((item) =>
          item.id === id ? { ...item, cantidad: item.cantidad - 1 } : item
        )
        .filter((item) => item.cantidad > 0)
    );
  };

  /**
   * Vacía todos los ítems del carrito.
   */
  const clearCart = () => {
    setCart([]);
  };

  const totalItems = cart.reduce((total, item) => total + item.cantidad, 0);
  const totalPrice = cart.reduce(
    (total, item) => total + item.precio * item.cantidad,
    0
  );

  const value: CartContextType = {
    cart,
    addToCart,
    removeFromCart,
    incrementQuantity,
    decrementQuantity,
    clearCart,
    totalItems,
    totalPrice,
  };

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
};
