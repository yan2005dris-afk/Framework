import { createFileRoute } from "@tanstack/react-router";
import { SaasPlanes } from "@/Components/Saas/SaasPlanes";

export const Route = createFileRoute("/_auth/planes")({
  component: SaasPlanes,
});
