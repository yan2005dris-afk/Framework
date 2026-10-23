import { useEffect } from "react";
import { router } from "@/router";
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
 *
 * Importa el `router` como singleton y NO con `useRouter()` a proposito: este
 * componente se monta en `AppShell`, al lado del `RouterProvider` pero FUERA de
 * el. Los hooks de TanStack (`useRouter`, `useRouterState`) leen el contexto
 * del router, que todavia no existe en ese punto: dan `null` y la app entera
 * deja de renderizar con `Cannot read properties of null (reading 'stores')`.
 * El singleton existe desde el import, asi que se puede usar antes y despues.
 */
export const OfflineGuard = () => {
  const isOnline = useOnlineStatus();

  useEffect(() => {
    if (!isOnline && !EXENTAS.has(router.state.location.pathname)) {
      void router.navigate({ to: "/offline" });
    }
  }, [isOnline]);

  return null;
};