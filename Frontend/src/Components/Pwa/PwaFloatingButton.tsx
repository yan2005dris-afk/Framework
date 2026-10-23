import { useState } from "react";
import { Download, Smartphone, X, CheckCircle2, Laptop } from "lucide-react";
import { usePwaInstall } from "@/hooks/usePwaInstall";

/**
 * Floating PWA Install Button (FAB):
 * Botón flotante circular en la esquina inferior DERECHA.
 * - Siempre visible mientras la app no esté ya instalada en modo standalone.
 * - Si el navegador soporta `beforeinstallprompt`, al hacer clic dispara el instalador nativo directamente.
 * - Si el navegador no lo soporta (Zen Browser, Firefox desktop, iOS Safari), abre una tarjeta modal moderna y bien estructurada.
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
      {/* Botón flotante circular inferior DERECHO */}
      <div className="fixed bottom-5 right-5 z-50">
        <button
          type="button"
          onClick={() => void handleClick()}
          className="group flex items-center gap-2.5 bg-indigo-600 hover:bg-indigo-700 text-white p-3.5 sm:px-4 sm:py-3 rounded-full shadow-2xl hover:shadow-indigo-500/50 hover:scale-105 active:scale-95 transition-all duration-200 focus:outline-none focus:ring-4 focus:ring-indigo-300"
          title="Descargar / Instalar Aplicación"
          aria-label="Descargar o instalar aplicación PWA"
        >
          <div className="relative">
            <Download className="w-5 h-5 animate-bounce" />
          </div>
          <span className="hidden sm:inline font-semibold text-sm tracking-wide pr-1">
            Instalar App
          </span>
        </button>
      </div>

      {/* Modal estructurado de instalación */}
      {modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs animate-in fade-in">
          <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6 relative border border-slate-200">
            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-slate-600 p-1.5 rounded-full hover:bg-slate-100 transition"
              aria-label="Cerrar"
            >
              <X className="w-5 h-5" />
            </button>

            <div className="flex items-center gap-3 mb-5">
              <div className="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
                <Smartphone className="w-6 h-6" />
              </div>
              <div>
                <h3 className="font-bold text-slate-900 text-lg">Instalar MultiCatálogo</h3>
                <p className="text-xs text-slate-500">Acceso rápido, modo sin conexión y experiencia nativa</p>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-3 mb-6">
              {/* Opción Celular Android */}
              <div className="flex items-start gap-3 p-3.5 bg-slate-50 rounded-xl border border-slate-200/80">
                <div className="w-8 h-8 rounded-lg bg-emerald-100 text-emerald-700 flex items-center justify-center font-bold text-sm shrink-0">
                  📱
                </div>
                <div>
                  <h4 className="font-semibold text-slate-900 text-sm">En Celular (Android)</h4>
                  <p className="text-xs text-slate-600 mt-0.5">
                    Toca el menú <strong className="text-slate-800">tres puntos (⋮)</strong> en Chrome o tu navegador y elige <strong className="text-slate-800">"Instalar aplicación"</strong> o <em>"Agregar a pantalla principal"</em>.
                  </p>
                </div>
              </div>

              {/* Opción iPhone */}
              <div className="flex items-start gap-3 p-3.5 bg-slate-50 rounded-xl border border-slate-200/80">
                <div className="w-8 h-8 rounded-lg bg-slate-200 text-slate-700 flex items-center justify-center font-bold text-sm shrink-0">
                  🍎
                </div>
                <div>
                  <h4 className="font-semibold text-slate-900 text-sm">En iPhone / iPad (Safari)</h4>
                  <p className="text-xs text-slate-600 mt-0.5">
                    Presiona el botón <strong className="text-slate-800">Compartir</strong> (ícono con flecha arriba) y pulsa <strong className="text-slate-800">"Agregar al inicio"</strong>.
                  </p>
                </div>
              </div>

              {/* Opción Computadora */}
              <div className="flex items-start gap-3 p-3.5 bg-slate-50 rounded-xl border border-slate-200/80">
                <div className="w-8 h-8 rounded-lg bg-indigo-100 text-indigo-600 flex items-center justify-center shrink-0">
                  <Laptop className="w-4 h-4" />
                </div>
                <div>
                  <h4 className="font-semibold text-slate-900 text-sm">En Computadora (Chrome / Edge / Zen)</h4>
                  <p className="text-xs text-slate-600 mt-0.5">
                    En Chrome/Edge haz clic en el ícono de <strong className="text-slate-800">Instalar</strong> en la barra de direcciones. En Zen/Firefox puedes fijar la pestaña como app de trabajo.
                  </p>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-2 text-xs text-emerald-800 bg-emerald-50/80 p-3 rounded-xl border border-emerald-200">
              <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-600" />
              <span>La aplicación se sincronizará y funcionará sin conexión.</span>
            </div>

            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="mt-5 w-full bg-slate-900 hover:bg-slate-800 text-white font-medium py-2.5 rounded-xl text-sm transition shadow-sm"
            >
              Listo
            </button>
          </div>
        </div>
      )}
    </>
  );
}
