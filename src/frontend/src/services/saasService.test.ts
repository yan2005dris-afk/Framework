import { describe, it, expect, vi, beforeEach } from "vitest";
import { saasApi } from "./saasService";
import * as http from "./http";

describe("saasApi Service", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("getPlans() debe llamar a /saas/plans y mapear los campos correctamente", async () => {
    const mockPlans = [
      {
        id: 1,
        slug: "free",
        name: "Free",
        description: "Plan gratuito",
        price_monthly: 0,
        price_yearly: 0,
        features: ["1 tenant", "5 usuarios"],
      },
      {
        id: 2,
        slug: "pro",
        name: "Pro",
        description: "Plan profesional",
        price_monthly: 29.9,
        price_yearly: 299.0,
        features: ["10 tenants", "Usuarios ilimitados"],
      },
    ];

    const spy = vi.spyOn(http, "apiFetch").mockResolvedValueOnce(mockPlans);

    const result = await saasApi.getPlans();
    expect(spy).toHaveBeenCalledWith("/saas/plans");
    expect(result).toHaveLength(2);
    expect(result[0].name).toBe("Free");
    expect(result[1].price_monthly).toBe(29.9);
  });

  it("getSubscription(tenantId) debe consultar /saas/tenants/{tenant_id}/subscription", async () => {
    const mockSub = {
      id: 10,
      tenant_id: 1,
      plan_id: 2,
      status: "active",
      cancel_at_period_end: false,
    };

    const spy = vi.spyOn(http, "apiFetch").mockResolvedValueOnce(mockSub);

    const result = await saasApi.getSubscription(1);
    expect(spy).toHaveBeenCalledWith("/saas/tenants/1/subscription");
    expect(result.plan_id).toBe(2);
    expect(result.status).toBe("active");
  });

  it("createSubscription(tenantId, data) debe enviar POST con plan_id", async () => {
    const mockSub = {
      id: 11,
      tenant_id: 1,
      plan_id: 3,
      status: "active",
      cancel_at_period_end: false,
    };

    const spy = vi.spyOn(http, "apiFetch").mockResolvedValueOnce(mockSub);

    const result = await saasApi.createSubscription(1, { plan_id: 3 });
    expect(spy).toHaveBeenCalledWith("/saas/tenants/1/subscription", {
      method: "POST",
      json: { plan_id: 3 },
    });
    expect(result.plan_id).toBe(3);
  });

  it("cancelSubscription(tenantId) debe enviar DELETE a la ruta de suscripción", async () => {
    const spy = vi.spyOn(http, "apiFetch").mockResolvedValueOnce(undefined);

    await saasApi.cancelSubscription(1);
    expect(spy).toHaveBeenCalledWith("/saas/tenants/1/subscription", {
      method: "DELETE",
    });
  });

  it("getUsers(tenantId) debe paginar correctamente", async () => {
    const mockUsersResponse = {
      items: [
        {
          id: 1,
          tenant_id: 1,
          email: "admin@tenant1.com",
          rol: "admin",
          is_active: true,
          created_at: "2026-10-22T00:00:00Z",
        },
      ],
      total: 1,
      page: 1,
      page_size: 10,
    };

    const spy = vi.spyOn(http, "apiFetch").mockResolvedValueOnce(mockUsersResponse);

    const result = await saasApi.getUsers(1, { page: 1, page_size: 10 });
    expect(spy).toHaveBeenCalledWith("/saas/tenants/1/users?page=1&page_size=10");
    expect(result.items).toHaveLength(1);
    expect(result.items[0].email).toBe("admin@tenant1.com");
  });
});
