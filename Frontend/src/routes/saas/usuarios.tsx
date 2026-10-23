import { createFileRoute } from "@tanstack/react-router";
import { SaasLayout } from "@/Components/Saas/SaasLayout";
import { SaasUsuarios } from "@/Components/Saas/SaasUsuarios";

export const Route = createFileRoute("/saas/usuarios")({
  component: () => (
    <SaasLayout>
      <SaasUsuarios />
    </SaasLayout>
  ),
});
