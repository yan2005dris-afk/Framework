import { apiFetch } from "./http";

export interface PlanSaaS {
  id: number;
  name: string;
  code: string;
  price_monthly: number;
  price_yearly: number;
  max_users: number;
  max_products: number;
  features: string[];
}

export interface SubscriptionSaaS {
  id: number;
  tenant_id: number;
  plan_id: number;
  status: "active" | "canceled" | "past_due" | "trialing";
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
}

export interface UserSaaS {
  id: number;
  tenant_id: number;
  email: string;
  rol: "superadmin" | "admin" | "member";
  is_active: boolean;
  created_at: string;
}

export interface UserSaaSListResponse {
  items: UserSaaS[];
  total: number;
  page: number;
  page_size: number;
}

export const saasApi = {
  /**
   * GET /api/saas/plans
   * Obtiene el catálogo de planes disponibles para todos los tenants.
   */
  async getPlans(): Promise<PlanSaaS[]> {
    return apiFetch<PlanSaaS[]>("/saas/plans");
  },

  /**
   * GET /api/saas/tenants/{tenant_id}/subscription
   * Obtiene la suscripción activa del tenant actual.
   */
  async getSubscription(tenantId: number): Promise<SubscriptionSaaS> {
    return apiFetch<SubscriptionSaaS>(`/saas/tenants/${tenantId}/subscription`);
  },

  /**
   * POST /api/saas/tenants/{tenant_id}/subscription
   * Crea o actualiza la suscripción del tenant.
   */
  async createSubscription(
    tenantId: number,
    data: { plan_id: number; billing_cycle?: "monthly" | "yearly" }
  ): Promise<SubscriptionSaaS> {
    return apiFetch<SubscriptionSaaS>(`/saas/tenants/${tenantId}/subscription`, {
      method: "POST",
      json: data,
    });
  },

  /**
   * DELETE /api/saas/tenants/{tenant_id}/subscription
   * Cancela la suscripción actual.
   */
  async cancelSubscription(tenantId: number): Promise<void> {
    return apiFetch<void>(`/saas/tenants/${tenantId}/subscription`, {
      method: "DELETE",
    });
  },

  /**
   * GET /api/saas/tenants/{tenant_id}/users
   * Lista los usuarios del tenant con paginación.
   */
  async getUsers(
    tenantId: number,
    params?: { page?: number; page_size?: number }
  ): Promise<UserSaaSListResponse> {
    const search = new URLSearchParams();
    if (params?.page) search.set("page", params.page.toString());
    if (params?.page_size) search.set("page_size", params.page_size.toString());
    const query = search.toString() ? `?${search.toString()}` : "";
    return apiFetch<UserSaaSListResponse>(`/saas/tenants/${tenantId}/users${query}`);
  },

  /**
   * POST /api/saas/tenants/{tenant_id}/users
   * Crea o invita un nuevo usuario al tenant.
   */
  async createUser(
    tenantId: number,
    user: { email: string; password?: string; rol: "admin" | "member" }
  ): Promise<UserSaaS> {
    return apiFetch<UserSaaS>(`/saas/tenants/${tenantId}/users`, {
      method: "POST",
      json: user,
    });
  },

  /**
   * DELETE /api/saas/tenants/{tenant_id}/users/{user_id}
   */
  async deleteUser(tenantId: number, userId: number): Promise<void> {
    return apiFetch<void>(`/saas/tenants/${tenantId}/users/${userId}`, {
      method: "DELETE",
    });
  },
};
