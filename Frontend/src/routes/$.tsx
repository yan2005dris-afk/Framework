import { createFileRoute, Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";

/**
 * Catch-all: cualquier path desconocido vuelve a la raiz, igual que el
 * `<Route path="*" element={<Navigate to="/" replace />} />` anterior.
 */
export const Route = createFileRoute("/$")({
  component: () => (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4">
      <h1 className="text-2xl font-semibold">Pagina no encontrada</h1>
      <p className="text-muted-foreground text-sm">
        La ruta que buscas no existe.
      </p>
      <Button asChild>
        <Link to="/">Volver al inicio</Link>
      </Button>
    </div>
  ),
});
