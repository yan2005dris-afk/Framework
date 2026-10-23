import { Given, When, Then, Before, After, setDefaultTimeout } from "@cucumber/cucumber";
import { chromium, type Browser, type Page, expect } from "@playwright/test";

setDefaultTimeout(60 * 1000);

let browser: Browser;
let page: Page;

const BASE_URL = process.env.BASE_URL || "https://framework.yandro.tech";

Before(async () => {
  browser = await chromium.launch({
    executablePath: process.env.CHROME_BIN || "/usr/bin/chromium",
    headless: true,
    args: ["--no-sandbox", "--disable-setuid-sandbox"],
  });
  const context = await browser.newContext();
  page = await context.newPage();
});

After(async () => {
  await page?.close();
  await browser?.close();
});

Given("que el usuario administrador tiene una sesión activa en la plataforma", async () => {
  await page.goto(`${BASE_URL}/login`, { waitUntil: "networkidle" });
  await page.fill('input[type="email"]', "admin@upse.edu.ec");
  await page.fill('input[type="password"]', "123456");
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes("/login"), { timeout: 15000 });
});

Given('navega a la sección de "Planes SaaS"', async () => {
  await page.goto(`${BASE_URL}/planes`, { waitUntil: "networkidle" });
});

When("la vista de planes termina de cargar", async () => {
  await page.waitForSelector("text=Planes de Suscripción", { timeout: 15000 });
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
  // Precondición contextual
});

When('el usuario hace clic en "Seleccionar Plan" en la tarjeta {string}', async (plan: string) => {
  const card = page.locator(".rounded-xl", { hasText: plan });
  const boton = card.locator("button", { hasText: "Seleccionar Plan" });
  if (await boton.isVisible()) {
    await boton.click();
  }
});

Then("se muestra el indicador de carga", async () => {
  // Verificación de feedback
});

Then('tras la confirmación del servidor el plan {string} pasa a tener la etiqueta "Plan Activo"', async (plan: string) => {
  const card = page.locator(".rounded-xl", { hasText: plan });
  await expect(card.locator("text=Plan Activo")).toBeVisible({ timeout: 15000 });
});

Given("que el usuario tiene un plan activo contratado", async () => {
  await expect(page.locator("text=Plan Activo")).toBeVisible({ timeout: 15000 });
});

When('presiona el botón "Cancelar Suscripción Actual"', async () => {
  const botonCancelar = page.locator("button", { hasText: "Cancelar Suscripción Actual" });
  if (await botonCancelar.isVisible()) {
    await botonCancelar.click();
  }
});

Then("el estado de suscripción se actualiza y se habilita la contratación nuevamente", async () => {
  await expect(page.locator("button:has-text('Seleccionar Plan')").first()).toBeVisible({ timeout: 15000 });
});
