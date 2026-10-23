import { useSyncExternalStore } from "react";

/**
 * Estado de conexion de la app (U4-T3).
 *
 * `useSyncExternalStore` y no un `useState` con listeners a mano: el estado
 * vive en el navegador, no en React, asi que la fuente de verdad tiene que
 * ser `navigator.onLine` y el store solo tiene que suscribirse a los eventos.
 * Asi no hay estado duplicado que pueda quedar desincronizado.
 *
 * El tercer argumento es el valor para renderizado en servidor (SSR): en el
 * servidor no hay conexion que consultar, y arrancar en `true` evita que la
 * primera pintura mande a la pagina offline a un usuario que si tiene red.
 */
const suscribirse = (alCambiar: () => void) => {
  window.addEventListener("online", alCambiar);
  window.addEventListener("offline", alCambiar);
  return () => {
    window.removeEventListener("online", alCambiar);
    window.removeEventListener("offline", alCambiar);
  };
};

export const useOnlineStatus = () =>
  useSyncExternalStore(
    suscribirse,
    () => navigator.onLine,
    () => true
  );