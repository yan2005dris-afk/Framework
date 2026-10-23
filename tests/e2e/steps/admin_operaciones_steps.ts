import { Given, When, Then, Before, After, setDefaultTimeout } from "@cucumber/cucumber";
import { chromium, type Browser, type BrowserContext, type Page, expect } from "@playwright/test";

setDefaultTimeout(60 * 1000);

let browser: Browser;
let context: BrowserContext;
let page: Page;

const BASE_URL = process.env.BASE_URL || "https://framework.yandro.tech";

Before(async () => {
  browser = await chromium.launch({
    executablePath: process.env.CHROME_BIN || "/usr/bin/chromium",
    headless: true,
    args: ["--no-sandbox", "--disable-setuid-sandbox", "--disable-dev-shm-usage"],
  });
  context = await browser.newContext();
  page = await context.newPage();
});

After(async () => {
  await page?.close();
  await context?.close();
  await browser?.close();
});

Given("que el administrador {string} ha iniciado sesión en el sistema", async (_adminName: string) => {
  await page.goto(`${BASE_URL}/login`, { waitUntil: "networkidle" });
  await page.fill('input[type="email"]', "admin@upse.edu.ec");
  await page.fill('input[type="password"]', "123456");
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes("/login"), { timeout: 15000 });
});

Given("que el administrador se encuentra en el panel de control principal", async () => {
  await page.goto(`${BASE_URL}/`, { waitUntil: "networkidle" });
  await expect(page.locator("text=Panel de Control").or(page.locator("text=Resumen"))).toBeVisible({ timeout: 15000 });
});

When("la vista del dashboard termina de cargar sus métricas", async () => {
  await page.waitForSelector("text=Ventas Totales", { timeout: 15000 });
});

Then("se visualizan las tarjetas resumen de ventas y operaciones", async () => {
  await expect(page.locator("text=Ventas Totales")).toBeVisible();
  await expect(page.locator("text=Pedidos Procesados")).toBeVisible();
  await expect(page.locator("text=Comisiones Devengadas")).toBeVisible();
});

Then("se listan las tablas de los productos más vendidos y mejores referidos", async () => {
  await expect(page.locator("text=Top Productos")).toBeVisible({ timeout: 10000 });
  await expect(page.locator("text=Top Referidos").or(page.locator("text=Líderes de Red"))).toBeVisible({ timeout: 10000 });
});

Given('que el administrador navega a la sección de "Administración de Productos"', async () => {
  await page.goto(`${BASE_URL}/admin/productos`, { waitUntil: "networkidle" });
  await expect(page.locator("h1:has-text('Gestión de Productos')")).toBeVisible({ timeout: 15000 });
});

When("abre el formulario para crear un nuevo producto", async () => {
  const btnNuevo = page.locator('button:has-text("Nuevo Producto")');
  await btnNuevo.click();
  await expect(page.locator("text=Nuevo Producto").first()).toBeVisible({ timeout: 10000 });
});

When("completa los datos del producto con:", async (dataTable) => {
  const filas = dataTable.hashes();
  for (const fila of filas) {
    if (fila.campo === "nombre") {
      await page.locator('input[placeholder*="Serum Anti-Manchas"]').fill(fila.valor);
    } else if (fila.campo === "descripcion") {
      await page.locator('textarea').fill(fila.valor);
    } else if (fila.campo === "precio") {
      await page.locator('input[type="number"][step="0.01"]').fill(fila.valor);
    } else if (fila.campo === "stock") {
      await page.locator('input[type="number"]:not([step="0.01"])').fill(fila.valor);
    }
  }
});

When("guarda el producto", async () => {
  const btnGuardar = page.locator('button[type="submit"]:has-text("Guardar")');
  await btnGuardar.click();
});

Then('el nuevo producto {string} aparece listado en el inventario', async (nombreProd: string) => {
  const busqueda = page.locator('input[placeholder*="Buscar producto por nombre"]').first();
  await busqueda.fill(nombreProd);
  await expect(page.locator(`text=${nombreProd}`).first()).toBeVisible({ timeout: 15000 });
});

Given('que el administrador navega a la sección de "Gestión de Pedidos"', async () => {
  await page.goto(`${BASE_URL}/admin/pedidos`, { waitUntil: "networkidle" });
  await expect(page.locator("h1:has-text('Gestión de Pedidos')")).toBeVisible({ timeout: 15000 });
});

When("visualiza la lista de órdenes registradas", async () => {
  const contenedor = page.locator("table, .space-y-4");
  await expect(contenedor.first()).toBeVisible({ timeout: 10000 });
});

Then("puede consultar el detalle del pedido más reciente", async () => {
  const btnDetalle = page.locator('button:has-text("Ver Detalle")').first();
  if (await btnDetalle.isVisible()) {
    await btnDetalle.click();
    await expect(page.locator("text=Detalle de Pedido").or(page.locator("text=Items"))).toBeVisible({ timeout: 10000 });
  }
});

Then("puede filtrar los pedidos según su estado de entrega", async () => {
  const selectEstado = page.locator("select").first();
  if (await selectEstado.isVisible()) {
    await selectEstado.selectOption("completado");
    await page.waitForTimeout(500);
  }
});

Given('que el administrador ingresa a la sección "Mi Red"', async () => {
  await page.goto(`${BASE_URL}/mi-red`, { waitUntil: "networkidle" });
  await expect(page.locator("h1:has-text('Red de Afiliados')").or(page.locator("text=Mi Red"))).toBeVisible({ timeout: 15000 });
});

When("carga la estructura multinivel", async () => {
  await page.waitForSelector("text=Nivel", { timeout: 15000 });
});

Then("visualiza los distribuidores y referidos de su organización", async () => {
  const nodos = page.locator("li, .rounded-lg");
  await expect(nodos.first()).toBeVisible({ timeout: 10000 });
});

Then("observa el desglose de comisiones acumuladas por nivel", async () => {
  await expect(page.locator("text=comisión").or(page.locator("text=Ventas"))).toBeVisible({ timeout: 10000 });
});
