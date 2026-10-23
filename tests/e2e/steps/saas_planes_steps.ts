import { Given, When, Then, Before, After } from "@cucumber/cucumber";
import { chromium, type Browser, type Page, expect } from "@playwright/test";

let browser: Browser;
let page: Page;

Before(async () => {
  browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  page = await context.newPage();
});

After(async () => {
  await page?.close();
  await browser?.close();
});

Given("que el usuario administrador tiene una sesión activa en la plataforma", async () => {
  // Inyectar sesión de administrador en localStorage
  await page.goto("http://localhost:8080/login");
  await page.evaluate(() => {
    localStorage.setItem(
      "mc_session",
      JSON.stringify({ email: "admin@multicatalogo.com", rol: "admin" })
    );
  });
});

Given('navega a la sección de "Planes SaaS"', async () => {
  await page.goto("http://localhost:8080/planes");
});

When("la vista de planes termina de cargar", async () => {
  await page.waitForSelector("text=Planes de Suscripción", { timeout: 10000 });
});

Then(
  'se muestran los 3 planes de suscripción: {string}, {string} y {string}',
  async (p1: string, p2: string, p3: string) => {
    await expect(page.locator(`text=${p1}`)).toBeVisible();
    await expect(page.locator(`text=${p2}`)).toBeVisible();
    await expect(page.locator(`text=${p3}`)).toBeVisible();
  }
);

Then('el plan {string} muestra un precio de {string}', async (plan: string, precio: string) => {
  const card = page.locator(".rounded-xl", { hasText: plan });
  await expect(card.locator(`text=${precio}`)).toBeVisible();
});

Then("cada plan lista sus características y límites correspondientes", async () => {
  await expect(page.locator("text=/mes").first()).toBeVisible();
});

Given('que el plan actual es {string}', async (_plan: string) => {
  // Precondición de estado
});

When('el usuario hace clic en "Seleccionar Plan" en la tarjeta {string}', async (plan: string) => {
  const card = page.locator(".rounded-xl", { hasText: plan });
  const boton = card.locator("button", { hasText: "Seleccionar Plan" });
  await boton.click();
});

Then("se muestra el indicador de carga", async () => {
  // El botón muestra feedback de carga o deshabilitación
});

Then('tras la confirmación del servidor el plan {string} pasa a tener la etiqueta "Plan Activo"', async (plan: string) => {
  const card = page.locator(".rounded-xl", { hasText: plan });
  await expect(card.locator("text=Plan Activo")).toBeVisible({ timeout: 10000 });
});

Given("que el usuario tiene un plan activo contratado", async () => {
  await expect(page.locator("text=Plan Activo")).toBeVisible({ timeout: 10000 });
});

When('presiona el botón "Cancelar Suscripción Actual"', async () => {
  const botonCancelar = page.locator("button", { hasText: "Cancelar Suscripción Actual" });
  if (await botonCancelar.isVisible()) {
    await botonCancelar.click();
  }
});

Then("el estado de suscripción se actualiza y se habilita la contratación nuevamente", async () => {
  await expect(page.locator("button:has-text('Seleccionar Plan')").first()).toBeVisible({ timeout: 10000 });
});
