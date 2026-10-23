import { useState } from "react";
import { Download, Smartphone, X, Globe, CheckCircle2 } from "lucide-react";
import { usePwaInstall } from "@/hooks/usePwaInstall";

/**
 * Floating PWA Install Button (FAB):
 * Botón flotante circular en la esquina inferior izquierda.
 * - Siempre visible mientras la app no esté ya instalada en modo standalone.
 * - Si el navegador soporta `beforeinstallprompt`, al hacer clic dispara el instalador nativo directamente.
 * - Si el navegador no lo soporta (iOS Safari, Zen, Firefox desktop), abre un modal elegante con los pasos exactos.
 */
export function PwaFloatingButton() {
  const { isInstallable, isInstalled, promptInstall } = usePwaInstall();
  const [modalOpen, setModalOpen] = useState(false);

  // Si ya corre instalada como PWA standalone, no se muestra
  if (isInstalled) {
    return null;
  }

  const handleClick = async () => {
    if (isInstallable) {
      const installed = await promptInstall();
      if (!installed) {
        setModalOpen(true);
      }
    } else {
      setModalOpen(true);
    }
  };

  return (
    <>
      {/* Botón flotante circular inferior izquierdo */}
      <div className="fixed bottom-5 left-5 z-50">
        <button
          type="button"
          onClick={() => void handleClick()}
          className="group flex items-center gap-2.5 bg-indigo-600 hover:bg-indigo-700 text-white p-3.5 sm:px-4 sm:py-3 rounded-full shadow-2xl hover:shadow-indigo-500/50 hover:scale-105 active:scale-95 transition-all duration-200 focus:outline-none focus:ring-4 focus:ring-indigo-300"
          title="Descargar / Instalar App"
          aria-label="Descargar o instalar aplicación PWA"
        >
          <div className="relative">
            <Download className="w-6 h-6 animate-bounce" />
          </div>
          <span className="hidden sm:inline font-semibold text-sm tracking-wide pr-1">
            Instalar App
          </span>
        </button>
      </div>

      {/* Modal interactivo de instrucciones para navegadores sin prompt automático */}
      {modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs animate-in fade-in">
          <div className="bg-white rounded-2xl shadow-2xl max-w-sm w-full p-6 relative border border-slate-200">
            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-slate-600 p-1.5 rounded-full hover:bg-slate-100 transition"
              aria-label="Cerrar"
            >
              <X className="w-5 h-5" />
            </button>

            <div className="flex items-center gap-3 mb-4">
              <div className="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
                <Smartphone className="w-6 h-6" />
              </div>
              <div>
                <h3 className="font-bold text-slate-900 text-lg">Instalar MultiCatálogo</h3>
                <p className="text-xs text-slate-500">Úsala como una app nativa en tu dispositivo</p>
              </div>
            </div>

            <div className="space-y-3.5 text-xs text-slate-700 mb-6">
              <div className="p-3 bg-slate-50 rounded-xl border border-slate-100">
                <p className="font-semibold text-slate-900 flex items-center gap-1.5 mb-1 text-sm">
                  <span className="text-base">📱</span> En Android (Chrome / Brave):
                </p>
                <p className="text-slate-600">
                  Toca el menú de <strong>tres puntos (⋮)</strong> arriba a la derecha y selecciona <strong>"Instalar aplicación"</strong> o <em>"Agregar a pantalla principal"</em>.
                </p>
              </div>

              <div className="p-3 bg-slate-50 rounded-xl border border-slate-100">
                <p className="font-semibold text-slate-900 flex items-center gap-1.5 mb-1 text-sm">
                  <span className="text-base">🍎</span> En iPhone / iPad (Safari):
                </p>
                <p className="text-slate-600">
                  Toca el botón <strong>Compartir</strong> (ícono de flecha hacia arriba) y selecciona <strong>"Agregar al inicio"</strong>.
                </p>
              </div>

              <div className="p-3 bg-slate-50 rounded-xl border border-slate-100">
                <p className="font-semibold text-slate-900 flex items-center gap-1.5 mb-1 text-sm">
                  <Globe className="w-4 h-4 text-indigo-600" /> En Computadora:
                </p>
                <p className="text-slate-600">
                  Haz clic en el ícono de <strong>Instalar</strong> en la barra de direcciones de Chrome/Edge o guarda como acceso directo.
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2 text-[11px] text-emerald-700 bg-emerald-50 p-2.5 rounded-lg border border-emerald-200">
              <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-600" />
              <span>Funciona sin conexión a internet y carga al instante.</span>
            </div>

            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="mt-5 w-full bg-slate-900 hover:bg-slate-800 text-white font-medium py-2.5 rounded-xl text-sm transition"
            >
              Entendido
            </button>
          </div>
        </div>
      )}
    </>
  );
}
