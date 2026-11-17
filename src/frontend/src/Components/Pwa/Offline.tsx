import { useRouter } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { useOnlineStatus } from "@/hooks/useOnlineStatus";

/**
 * Pagina offline (U4-T3, S7).
 *
 * Aca manda el service worker cuando la navegacion no encuentra red: sirve el
 * index.html del build (navigateFallback) y el router resuelve esta ruta. Es
 * una ruta del app y no un HTML suelto, asi que comparte estilos y tipografia
 * con el resto.
 */
const Offline = () => {
  const router = useRouter();
  const isOnline = useOnlineStatus();

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 text-center">
      <h1 className="text-2xl font-semibold">Sin conexion</h1>
      <p className="text-muted-foreground max-w-md text-sm">
        {isOnline
          ? "La conexion volvio. Volve a cargar la pagina para seguir."
          : "No pudimos conectarnos. Revisa tu red y volve a intentar: la app y el catalogo quedan disponibles apenas recuperes la conexion."}
      </p>
      <Button
        onClick={() => {
          // `invalidate` fuerza el router a volver a pedir los datos de la ruta
          // de destino: sin esto volverian los datos del cache, que es
          // justamente lo que el usuario quiere refrescar al reintentar.
          void router.invalidate();
          void router.navigate({ to: "/" });
        }}
      >
        Reintentar
      </Button>
    </div>
  );
};

export default Offline;