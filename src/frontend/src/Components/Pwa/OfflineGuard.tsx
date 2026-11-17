import { useEffect } from "react";
import { router } from "@/router";
import { useOnlineStatus } from "@/hooks/useOnlineStatus";

/** Rutas donde no se redirige a pantalla de bloqueo: catálogo, tienda, carrito y login pueden funcionar con datos en caché. */
const EXENTAS = new Set(["/offline", "/login", "/tienda", "/catalogo", "/carrito", "/"]);

/**
 * Manda a la pagina offline únicamente cuando se intenta navegar a vistas
 * que requieren comunicación viva obligatoria con el servidor (como admin o red),
 * permitiendo que la tienda y catálogo sigan siendo navegables offline.
 */
export const OfflineGuard = () => {
  const isOnline = useOnlineStatus();

  useEffect(() => {
    const path = router.state.location.pathname;
    if (!isOnline && !EXENTAS.has(path) && !path.startsWith("/producto/")) {
      void router.navigate({ to: "/offline" });
    }
  }, [isOnline]);

  return null;
};
