import { createFileRoute } from "@tanstack/react-router";
import { SaasDashboard } from "@/Components/Saas/SaasDashboard";

export const Route = createFileRoute("/_auth/saas")({
  component: SaasDashboard,
});
