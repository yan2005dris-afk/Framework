import { useEffect } from "react";
import { useRouter, useRouterState } from "@tanstack/react-router";
import { useOnlineStatus } from "@/hooks/useOnlineStatus";

/** Rutas donde no se redirige: la propia offline y el login. */
const EXENTAS = new Set(["/offline", "/login"]);

/**
 * Manda a la pagina offline en cuanto se cae la conexion (U4-T3, S7).
 *
 * El service worker resuelve el caso de una navegacion con la red caida, pero
 * no el de una app que ya esta abierta y pierde la conexion: en ese caso cada
 * request va a fallar por separado y el usuario ve errores sueltos en vez de un
 * mensaje claro.
 *
 * El login queda exento porque es justamente donde se entra: obligar a pasar
 * por /offline cuando se pierde la conexion a medio login pierde el formulario
 * que el usuario ya habia escrito.
 */
export const OfflineGuard = () => {
  const isOnline = useOnlineStatus();
  const router = useRouter();
  const pathname = useRouterState({ select: (estado) => estado.location.pathname });

  useEffect(() => {
    if (!isOnline && !EXENTAS.has(pathname)) {
      void router.navigate({ to: "/offline" });
    }
  }, [isOnline, pathname, router]);

  return null;
};