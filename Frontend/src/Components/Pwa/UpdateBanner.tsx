import { useRegisterSW } from "virtual:pwa-register/react";
import { Button } from "@/components/ui/button";

/**
 * Banner de actualizacion del service worker (U4-T3).
 *
 * El `registerType: "prompt"` de vite.config.ts deja de lado la version nueva
 * en vez de aplicarla sola. Sin este banner, el usuario se quedaria en la
 * version vieja hasta recargar a mano y no tendria ninguna pista de por que.
 *
 * No va por `sonner` a proposito: el proyecto tiene el componente Toaster pero
 * no lo monta en ninguna parte de la app, y colgar el banner de ahi seria
 * montar un subsistema de notificaciones entero para un mensaje. Es un banner,
 * y va como banner.
 */
export const UpdateBanner = () => {
  const {
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW();

  if (!needRefresh) return null;

  return (
    <div
      role="status"
      className="bg-background animate-in slide-in-from-bottom-4 fixed inset-x-4 bottom-4 z-50 mx-auto flex max-w-md items-center gap-3 rounded-lg border p-3 shadow-lg"
    >
      <p className="text-sm">Hay una version nueva disponible.</p>
      <div className="ml-auto flex shrink-0 gap-2">
        <Button
          variant="ghost"
          size="sm"
          // Descartar el aviso no bloquea la actualizacion: la proxima
          // navegacion vuelve a ofrecerla.
          onClick={() => setNeedRefresh(false)}
        >
          Ahora no
        </Button>
        <Button size="sm" onClick={() => void updateServiceWorker(true)}>
          Recargar
        </Button>
      </div>
    </div>
  );
};