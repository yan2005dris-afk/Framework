import { useEffect, useState } from "react";
import { saasApi, type PlanSaaS, type SubscriptionSaaS } from "@/services/saasService";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Check, Loader2 } from "lucide-react";
import { toast } from "sonner";

export const SaasPlanes = () => {
  const tenantId = 1;
  const [plans, setPlans] = useState<PlanSaaS[]>([]);
  const [subscription, setSubscription] = useState<SubscriptionSaaS | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<number | null>(null);

  useEffect(() => {
    let activo = true;
    Promise.allSettled([
      saasApi.getPlans(),
      saasApi.getSubscription(tenantId),
    ]).then(([resPlans, resSub]) => {
      if (!activo) return;
      if (resPlans.status === "fulfilled") setPlans(resPlans.value);
      if (resSub.status === "fulfilled") setSubscription(resSub.value);
      setLoading(false);
    });

    return () => {
      activo = false;
    };
  }, []);

  const handleSubscribe = async (planId: number) => {
    try {
      setActionLoading(planId);
      const nuevaSub = await saasApi.createSubscription(tenantId, {
        plan_id: planId,
        billing_cycle: "monthly",
      });
      setSubscription(nuevaSub);
      toast.success("¡Suscripción actualizada con éxito!");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Error al procesar suscripción");
    } finally {
      setActionLoading(null);
    }
  };

  const handleCancel = async () => {
    try {
      setActionLoading(999999);
      await saasApi.cancelSubscription(tenantId);
      setSubscription(null);
      toast.success("Suscripción cancelada.");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Error al cancelar suscripción");
    } finally {
      setActionLoading(null);
    }
  };

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center text-muted-foreground text-sm">
        Cargando catálogo de planes...
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Planes de Suscripción</h1>
          <p className="text-muted-foreground">
            Selecciona el plan que mejor se adapte al volumen de tu organización.
          </p>
        </div>
        {subscription && (
          <Button variant="destructive" size="sm" onClick={handleCancel} disabled={actionLoading !== null}>
            Cancelar Suscripción Actual
          </Button>
        )}
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        {plans.map((plan) => {
          const isCurrent = subscription?.plan_id === plan.id;
          const isPending = actionLoading === plan.id;

          return (
            <Card
              key={plan.id}
              className={`flex flex-col relative overflow-hidden transition-all hover:shadow-md ${
                isCurrent ? "border-primary ring-2 ring-primary/20" : ""
              }`}
            >
              {isCurrent && (
                <div className="absolute top-0 right-0">
                  <Badge className="rounded-none rounded-bl-lg">Plan Activo</Badge>
                </div>
              )}

              <CardHeader>
                <CardTitle className="text-xl">{plan.name}</CardTitle>
                <CardDescription>Código: {plan.code}</CardDescription>
                <div className="mt-4 flex items-baseline gap-1">
                  <span className="text-4xl font-extrabold tracking-tight">${plan.price_monthly}</span>
                  <span className="text-sm font-semibold text-muted-foreground">/mes</span>
                </div>
                <div className="text-xs text-muted-foreground">
                  o ${plan.price_yearly} al año (ahorro anual)
                </div>
              </CardHeader>

              <CardContent className="flex-1 space-y-4">
                <div className="border-t pt-4 space-y-2 text-sm">
                  <div className="flex items-center gap-2">
                    <Check className="h-4 w-4 text-emerald-500 shrink-0" />
                    <span>Hasta <strong>{plan.max_users}</strong> usuarios</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Check className="h-4 w-4 text-emerald-500 shrink-0" />
                    <span>Hasta <strong>{plan.max_products}</strong> productos</span>
                  </div>
                  {plan.features?.map((f, i) => (
                    <div key={i} className="flex items-center gap-2">
                      <Check className="h-4 w-4 text-emerald-500 shrink-0" />
                      <span>{f}</span>
                    </div>
                  ))}
                </div>
              </CardContent>

              <CardFooter className="pt-4 border-t">
                <Button
                  className="w-full"
                  variant={isCurrent ? "outline" : "default"}
                  disabled={isCurrent || isPending}
                  onClick={() => handleSubscribe(plan.id)}
                >
                  {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  {isCurrent ? "Plan Actual" : "Seleccionar Plan"}
                </Button>
              </CardFooter>
            </Card>
          );
        })}
      </div>
    </div>
  );
};
