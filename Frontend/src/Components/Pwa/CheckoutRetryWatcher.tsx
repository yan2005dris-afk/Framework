import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { useOnlineStatus } from "@/hooks/useOnlineStatus";
import { borrarIntento, leerIntentos } from "@/lib/checkoutRetry";
import { NetworkError } from "@/services/http";
import { reintentarCheckout } from "@/services/api";

/**
 * Reintenta los checkouts que quedaron pendientes por falta de red (U4-T3, S9).
 *
 * El recorrido es corto a propósito: cada intento guardado se reenvia con SU
 * `Idempotency-Key`, y el Backend resuelve si crea el pedido o devuelve el que
 * esa key ya creo. Por eso reintentar no puede duplicar nada.
 *
 * Si el reintento falla con un error del servidor (4xx o 5xx, no red), el
 * intento se descarta: ya no es reintentable y dejarlo en la cola lo convertiria
 * en un bucle de requests.
 */
export const CheckoutRetryWatcher = () => {
  const isOnline = useOnlineStatus();
  const [reenviados, setReenviados] = useState(0);

  useEffect(() => {
    if (!isOnline) return;

    const pendientes = leerIntentos();
    if (pendientes.length === 0) return;

    let vigente = true;
    void (async () => {
      for (const intento of pendientes) {
        try {
          await reintentarCheckout(intento);
          borrarIntento(intento.idempotencyKey);
          if (vigente) setReenviados((n) => n + 1);
        } catch (error) {
          if (!(error instanceof NetworkError)) {
            // Respuesta del servidor: reintentar es inútil.
            borrarIntento(intento.idempotencyKey);
          }
        }
      }
    })();

    return () => {
      vigente = false;
    };
  }, [isOnline]);

  if (reenviados === 0) return null;

  return (
    <div
      role="status"
      className="bg-background animate-in slide-in-from-bottom-4 fixed inset-x-4 bottom-4 z-50 mx-auto flex max-w-md items-center gap-3 rounded-lg border p-3 shadow-lg"
    >
      <p className="text-sm">
        {reenviados === 1
          ? "Recuperamos la conexión y confirmamos tu pedido."
          : `Recuperamos la conexión y confirmamos ${reenviados} pedidos.`}
      </p>
      <Button
        variant="ghost"
        size="sm"
        className="ml-auto"
        onClick={() => setReenviados(0)}
      >
        Ok
      </Button>
    </div>
  );
};