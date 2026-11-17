import { createFileRoute } from "@tanstack/react-router";
import Offline from "@/Components/Pwa/Offline";

/**
 * Ruta publica para cuando no hay red (U4-T3, S7).
 *
 * Sin `beforeLoad` de auth a proposito: la pagina offline tiene que funcionar
 * justo cuando la sesion no se puede validar contra /api/me, que es el
 * momento en que mas se necesita.
 */
export const Route = createFileRoute("/offline")({
  component: () => <Offline />,
});