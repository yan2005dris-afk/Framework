import { useSyncExternalStore } from "react";

interface BeforeInstallPromptEvent extends Event {
  readonly platforms: string[];
  readonly userChoice: Promise<{
    outcome: "accepted" | "dismissed";
    platform: string;
  }>;
  prompt(): Promise<void>;
}

declare global {
  interface Window {
    __deferredPrompt?: BeforeInstallPromptEvent | null;
  }
}

// Store externo para sincronizar con React sin setState dentro de useEffect
let deferredPrompt: BeforeInstallPromptEvent | null =
  typeof window !== "undefined" && window.__deferredPrompt ? window.__deferredPrompt : null;

const listeners = new Set<() => void>();

function emitChange() {
  for (const listener of listeners) {
    listener();
  }
}

if (typeof window !== "undefined") {
  window.addEventListener("beforeinstallprompt", (e: Event) => {
    e.preventDefault();
    deferredPrompt = e as BeforeInstallPromptEvent;
    window.__deferredPrompt = deferredPrompt;
    emitChange();
  });

  window.addEventListener("appinstalled", () => {
    deferredPrompt = null;
    window.__deferredPrompt = null;
    emitChange();
  });
}

function subscribe(callback: () => void) {
  listeners.add(callback);
  return () => {
    listeners.delete(callback);
  };
}

function getSnapshot() {
  return deferredPrompt;
}

function getServerSnapshot() {
  return null;
}

function checkIsStandalone(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    (navigator as unknown as { standalone?: boolean }).standalone === true
  );
}

/**
 * Hook para capturar el evento beforeinstallprompt y permitir la instalación
 * manual de la PWA desde la interfaz de usuario.
 */
export function usePwaInstall() {
  const currentPrompt = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const isInstalled = checkIsStandalone();
  const isInstallable = Boolean(currentPrompt) && !isInstalled;

  const promptInstall = async (): Promise<boolean> => {
    if (!currentPrompt) {
      return false;
    }

    await currentPrompt.prompt();
    const choice = await currentPrompt.userChoice;

    if (choice.outcome === "accepted") {
      deferredPrompt = null;
      if (typeof window !== "undefined") {
        window.__deferredPrompt = null;
      }
      emitChange();
      return true;
    }

    return false;
  };

  return {
    isInstallable,
    isInstalled,
    promptInstall,
  };
}
