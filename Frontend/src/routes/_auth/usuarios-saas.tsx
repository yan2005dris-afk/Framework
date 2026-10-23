import { createFileRoute } from "@tanstack/react-router";
import { SaasUsuarios } from "@/Components/Saas/SaasUsuarios";

export const Route = createFileRoute("/_auth/usuarios-saas")({
  component: SaasUsuarios,
});
