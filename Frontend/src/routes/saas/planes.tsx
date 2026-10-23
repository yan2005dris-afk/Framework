import { createFileRoute } from "@tanstack/react-router";
import { SaasLayout } from "@/Components/Saas/SaasLayout";
import { SaasPlanes } from "@/Components/Saas/SaasPlanes";

export const Route = createFileRoute("/saas/planes")({
  component: () => (
    <SaasLayout>
      <SaasPlanes />
    </SaasLayout>
  ),
});
