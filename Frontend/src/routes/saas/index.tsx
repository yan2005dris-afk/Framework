import { createFileRoute } from "@tanstack/react-router";
import { SaasLayout } from "@/Components/Saas/SaasLayout";
import { SaasDashboard } from "@/Components/Saas/SaasDashboard";

export const Route = createFileRoute("/saas/")({
  component: () => (
    <SaasLayout>
      <SaasDashboard />
    </SaasLayout>
  ),
});
