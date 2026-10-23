import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { registerSW } from "virtual:pwa-register";
import "./index.css";
import { AppShell } from "./app/app-shell";

// U4-T3: registro del service worker.
// `immediate: true` registra sin esperar al evento `load`. Sin esto el service
// worker tarda lo que tarda en cargar la pagina en aparecer, y durante esa
// ventana la app no cumple los criterios de instalabilidad.
registerSW({
  immediate: true,
  onRegisterError(error) {
    // Fallar el registro no debe romper la app: sin service worker sigue
    // siendo una web normal, solo que no instalable ni offline.
    console.error("[pwa] no se pudo registrar el service worker:", error);
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppShell />
  </StrictMode>
);