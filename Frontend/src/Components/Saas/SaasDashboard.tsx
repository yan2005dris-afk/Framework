import { useEffect, useState } from "react";
import { saasApi, type PlanSaaS, type SubscriptionSaaS } from "@/services/saasService";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";
import { Users, CreditCard, Layers, AlertCircle } from "lucide-react";

export const SaasDashboard = () => {
  const tenantId = 1;
  const [plans, setPlans] = useState<PlanSaaS[]>([]);
  const [subscription, setSubscription] = useState<SubscriptionSaaS | null>(null);
  const [usersCount, setUsersCount] = useState<number>(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let activo = true;
    Promise.allSettled([
      saasApi.getPlans(),
      saasApi.getSubscription(tenantId),
      saasApi.getUsers(tenantId, { page_size: 1 }),
    ]).then(([resPlans, resSub, resUsers]) => {
      if (!activo) return;

      if (resPlans.status === "fulfilled") setPlans(resPlans.value);
      if (resSub.status === "fulfilled") setSubscription(resSub.value);
      if (resUsers.status === "fulfilled") setUsersCount(resUsers.value.total);

      setLoading(false);
    }).catch(() => {
      if (activo) {
        setLoading(false);
      }
    });

    return () => {
      activo = false;
    };
  }, []);

  const activePlan = plans.find((p) => p.id === subscription?.plan_id);

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center text-muted-foreground text-sm">
        Cargando métricas y suscripción del tenant...
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Panel SaaS</h1>
        <p className="text-muted-foreground">
          Gestión del tenant actual, suscripción a planes y equipo de trabajo.
        </p>
      </div>

      {/* KPI Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Plan Activo</CardTitle>
            <CreditCard className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activePlan?.name || "Sin Suscripción"}</div>
            <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
              {subscription ? (
                <>
                  <span className="inline-block h-2 w-2 rounded-full bg-emerald-500" />
                  Estado: <span className="capitalize font-medium text-emerald-600">{subscription.status}</span>
                </>
              ) : (
                "Ningún plan contratado"
              )}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Usuarios SaaS</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{usersCount}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Miembros con acceso a este tenant
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Planes en Catálogo</CardTitle>
            <Layers className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{plans.length}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Opciones de escalamiento disponibles
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Subscription Quick Info */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card className="flex flex-col">
          <CardHeader>
            <CardTitle>Detalles de Facturación</CardTitle>
            <CardDescription>Información sobre el ciclo de pago actual</CardDescription>
          </CardHeader>
          <CardContent className="flex-1 space-y-4">
            {subscription ? (
              <div className="space-y-3 text-sm">
                <div className="flex justify-between py-1 border-b">
                  <span className="text-muted-foreground">Periodo actual:</span>
                  <span className="font-medium">
                    {new Date(subscription.current_period_start).toLocaleDateString()} -{" "}
                    {new Date(subscription.current_period_end).toLocaleDateString()}
                  </span>
                </div>
                <div className="flex justify-between py-1 border-b">
                  <span className="text-muted-foreground">Renovación automática:</span>
                  <span className="font-medium">
                    {subscription.cancel_at_period_end ? "Cancelada al final del periodo" : "Activa"}
                  </span>
                </div>
                <div className="pt-2">
                  <Link to="/saas/planes">
                    <Button variant="outline" size="sm" className="w-full">
                      Cambiar Plan o Gestionar
                    </Button>
                  </Link>
                </div>
              </div>
            ) : (
              <div className="text-center py-6 space-y-3">
                <AlertCircle className="h-8 w-8 text-amber-500 mx-auto" />
                <p className="text-sm text-muted-foreground">Este tenant aún no tiene un plan activo.</p>
                <Link to="/saas/planes">
                  <Button size="sm">Explorar Planes</Button>
                </Link>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Planes destacados */}
        <Card className="flex flex-col">
          <CardHeader>
            <CardTitle>Planes Recomendados</CardTitle>
            <CardDescription>Escala la capacidad de tu negocio</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {plans.slice(0, 2).map((plan) => (
              <div
                key={plan.id}
                className="flex items-center justify-between p-3 rounded-lg border bg-card/50 hover:bg-accent/40 transition"
              >
                <div>
                  <div className="font-semibold text-sm">{plan.name}</div>
                  <div className="text-xs text-muted-foreground">
                    Hasta {plan.max_users} usuarios • ${plan.price_monthly}/mes
                  </div>
                </div>
                <Badge variant={activePlan?.id === plan.id ? "default" : "secondary"}>
                  {activePlan?.id === plan.id ? "Actual" : "Disponible"}
                </Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};
