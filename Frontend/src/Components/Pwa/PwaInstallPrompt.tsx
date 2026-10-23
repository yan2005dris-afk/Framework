import { useState } from "react";
import { Download, HelpCircle, X, Smartphone, Globe } from "lucide-react";
import { usePwaInstall } from "@/hooks/usePwaInstall";

/**
 * Componente que muestra el botón de instalación directa (cuando el navegador
 * soporta `beforeinstallprompt`) o una guía clara con instrucciones para
 * dispositivos móviles (Android/iOS) y navegadores de escritorio.
 */
export function PwaInstallPrompt() {
  const { isInstallable, isInstalled, promptInstall } = usePwaInstall();
  const [showInstructions, setShowInstructions] = useState(false);

  // Si ya se está ejecutando como PWA instalada, no mostramos nada
  if (isInstalled) {
    return null;
  }

  return (
    <div className="mt-4 pt-4 border-t border-slate-200">
      {isInstallable ? (
        <button
          type="button"
          onClick={() => void promptInstall()}
          className="w-full min-h-[44px] bg-indigo-50 border-2 border-indigo-600 text-indigo-700 font-semibold py-2.5 px-4 rounded-lg hover:bg-indigo-100 transition flex items-center justify-center gap-2 shadow-sm text-sm"
        >
          <Download className="w-4 h-4 text-indigo-600" />
          <span>Descargar / Instalar App (PWA)</span>
        </button>
      ) : (
        <button
          type="button"
          onClick={() => setShowInstructions(true)}
          className="w-full min-h-[44px] bg-slate-50 border border-slate-300 text-slate-700 font-medium py-2.5 px-4 rounded-lg hover:bg-slate-100 transition flex items-center justify-center gap-2 text-sm"
        >
          <Smartphone className="w-4 h-4 text-slate-600" />
          <span>¿Cómo instalar en mi celular?</span>
          <HelpCircle className="w-4 h-4 text-slate-400 ml-auto" />
        </button>
      )}

      {showInstructions && (
        <div className="mt-3 p-4 bg-slate-50 rounded-lg border border-slate-200 text-xs text-slate-700 relative animate-in fade-in">
          <button
            type="button"
            onClick={() => setShowInstructions(false)}
            className="absolute top-2 right-2 text-slate-400 hover:text-slate-600 p-1"
            aria-label="Cerrar instrucciones"
          >
            <X className="w-4 h-4" />
          </button>
          <p className="font-semibold text-slate-900 mb-2 flex items-center gap-1.5">
            <Globe className="w-4 h-4 text-indigo-600" />
            Cómo instalar esta aplicación:
          </p>
          <ul className="space-y-1.5 list-disc list-inside text-slate-600">
            <li>
              <strong className="text-slate-800">En Android (Chrome):</strong> Toca los tres puntos (⋮) y selecciona <em>"Instalar aplicación"</em> o <em>"Agregar a pantalla principal"</em>.
            </li>
            <li>
              <strong className="text-slate-800">En iPhone (Safari):</strong> Toca el botón Compartir y elige <em>"Agregar a la pantalla de inicio"</em>.
            </li>
            <li>
              <strong className="text-slate-800">En Computadora:</strong> Haz clic en el ícono de instalación en la barra de direcciones de tu navegador.
            </li>
          </ul>
        </div>
      )}
    </div>
  );
}
