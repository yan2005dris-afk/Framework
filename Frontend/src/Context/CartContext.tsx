/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import type { Producto } from "../data/productos";
import { useAuth } from "./AuthContext";

/**
 * Elemento del carrito que incluye los datos del producto y su cantidad seleccionada.
 */
export interface CartItem extends Producto {
  cantidad: number;
}

interface CartContextType {
  cart: CartItem[];
  addToCart: (producto: Producto) => void;
  removeFromCart: (id: number) => void;
  incrementQuantity: (id: number) => void;
  decrementQuantity: (id: number) => void;
  clearCart: () => void;
  totalItems: number;
  totalPrice: number;
}

const CART_STORAGE_PREFIX = "multicatalogo_carrito_";

const CartContext = createContext<CartContextType | undefined>(undefined);

/**
 * Hook para acceder al contexto del carrito de compras.
 * @returns {CartContextType} Contexto del carrito.
 */
export const useCart = (): CartContextType => {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error("useCart debe ser usado dentro de un CartProvider");
  }
  return context;
};

interface CartProviderProps {
  children: ReactNode;
}

/**
 * Proveedor del carrito de compras con persistencia aislada por usuario en localStorage.
 */
export const CartProvider = ({ children }: CartProviderProps) => {
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

  return (
    <CartContext.Provider
      value={{
        cart,
        addToCart,
        removeFromCart,
        incrementQuantity,
        decrementQuantity,
        clearCart,
        totalItems,
        totalPrice,
      }}
    >
      {children}
    </CartContext.Provider>
  );
};