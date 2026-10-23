import type { ReactNode } from "react";
import { SaasSidebar } from "./SaasSidebar";
import { useAuth } from "@/Context/useAuth";

interface SaasLayoutProps {
  children: ReactNode;
}

export const SaasLayout = ({ children }: SaasLayoutProps) => {
  const { user } = useAuth();
  // El tenant por defecto para el administrador principal o demo
  const tenantId = 1;

  return (
    <div className="flex min-h-screen bg-background">
      <SaasSidebar tenantId={tenantId} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top bar */}
        <header className="h-16 border-b bg-card flex items-center justify-between px-6">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-muted-foreground">Panel de Control Multitenant</span>
          </div>
          <div className="flex items-center gap-4 text-sm">
            <div className="flex flex-col text-right">
              <span className="font-medium text-foreground">{user?.email || "Admin SaaS"}</span>
              <span className="text-xs text-muted-foreground capitalize">{user?.rol || "SaaS Admin"}</span>
            </div>
          </div>
        </header>

        {/* Main Content Area */}
        <main className="flex-1 overflow-y-auto p-6 md:p-8 bg-muted/20">
          <div className="max-w-7xl mx-auto space-y-6">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
};
