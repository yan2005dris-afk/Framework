import { createFileRoute } from "@tanstack/react-router";
import Storefront from "@/Components/Storefront/Storefront";

export const Route = createFileRoute("/_auth/tienda")({
  component: Storefront,
});
