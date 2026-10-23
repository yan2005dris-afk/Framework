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

Given("que existe el usuario cliente con credenciales válidas en la plataforma", async () => {
  // Verificación de salud del sitio
  const resp = await page.goto(`${BASE_URL}/login`, { waitUntil: "networkidle" });
  expect(resp?.status()).toBeLessThan(400);
});

Given("que el cliente {string} navega a la página de login", async (_nombre: string) => {
  await page.goto(`${BASE_URL}/login`, { waitUntil: "networkidle" });
  await expect(page.locator("text=MultiCatálogo").first()).toBeVisible({ timeout: 15000 });
});

When("ingresa su correo {string} y contraseña {string}", async (email: string, password: string) => {
  const emailInput = page.locator('input[type="email"]');
  const passwordInput = page.locator('input[type="password"]');
  await emailInput.fill(email);
  await passwordInput.fill(password);
});

When("presiona el botón de iniciar sesión", async () => {
  const submitBtn = page.locator('button[type="submit"]');
  await submitBtn.click();
});

Then("es redirigido a la tienda o catálogo de productos", async () => {
  await page.waitForURL((url) => !url.pathname.includes("/login"), { timeout: 15000 });
  const currentPath = new URL(page.url()).pathname;
  expect(["/tienda", "/catalogo", "/"]).toContain(currentPath);
});

Then("visualiza el catálogo de productos disponibles", async () => {
  await page.goto(`${BASE_URL}/catalogo`, { waitUntil: "networkidle" });
  await expect(page.locator("text=Catálogo de Productos")).toBeVisible({ timeout: 15000 });
});

Given("que el cliente está en el catálogo de productos", async () => {
  // Asegurar sesión como cliente activo
  await page.goto(`${BASE_URL}/login`, { waitUntil: "networkidle" });
  await page.fill('input[type="email"]', "cliente@upse.edu.ec");
  await page.fill('input[type="password"]', "123456");
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes("/login"), { timeout: 15000 });
  await page.goto(`${BASE_URL}/catalogo`, { waitUntil: "networkidle" });
  await expect(page.locator("text=Catálogo de Productos")).toBeVisible({ timeout: 15000 });
});

When("busca el producto {string}", async (nombreProducto: string) => {
  const inputBusqueda = page.locator('input[placeholder*="Buscar"]').first();
  await inputBusqueda.fill(nombreProducto);
  await page.waitForTimeout(300);
});

When("añade el producto {string} al carrito de compras", async (nombreProducto: string) => {
  const tarjeta = page.locator(".bg-white", { hasText: nombreProducto }).first();
  const botonAgregar = tarjeta.locator('button:has-text("Añadir al Carrito")');
  await botonAgregar.click();
});

Then("el carrito de compras refleja {int} productos seleccionados", async (cantidad: number) => {
  // Ir al carrito
  await page.goto(`${BASE_URL}/carrito`, { waitUntil: "networkidle" });
  await expect(page.locator("text=Carrito de Compras")).toBeVisible({ timeout: 15000 });
  await expect(page.locator("text=Proceder al Pago")).toBeVisible({ timeout: 10000 });
  const itemsEnCarrito = page.locator('button[aria-label="Eliminar producto"]');
  await expect(itemsEnCarrito).toHaveCount(cantidad, { timeout: 10000 });
});

Given("que el cliente tiene productos en su carrito", async () => {
  // Login y agregar al menos un producto al carrito
  await page.goto(`${BASE_URL}/login`, { waitUntil: "networkidle" });
  await page.fill('input[type="email"]', "cliente@upse.edu.ec");
  await page.fill('input[type="password"]', "123456");
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes("/login"), { timeout: 15000 });

  await page.goto(`${BASE_URL}/catalogo`, { waitUntil: "networkidle" });
  await expect(page.locator("text=Catálogo de Productos")).toBeVisible({ timeout: 15000 });
  const primerBoton = page.locator('button:has-text("Añadir al Carrito")').first();
  await primerBoton.click();
});

Given("procede a la pantalla de finalización de compra", async () => {
  await page.goto(`${BASE_URL}/checkout`, { waitUntil: "networkidle" });
  await expect(page.locator("text=Finalizar Compra")).toBeVisible({ timeout: 15000 });
});

When("completa los datos de envío con:", async (dataTable) => {
  const filas = dataTable.hashes();
  for (const fila of filas) {
    if (fila.campo === "nombre") {
      await page.locator('input[type="text"]').first().fill(fila.valor);
    } else if (fila.campo === "email") {
      await page.locator('input[type="email"]').fill(fila.valor);
    } else if (fila.campo === "direccion") {
      // El tercer input de texto es dirección
      const inputs = page.locator('input[type="text"]');
      await inputs.nth(1).fill(fila.valor);
    } else if (fila.campo === "ciudad") {
      const inputs = page.locator('input[type="text"]');
      await inputs.nth(2).fill(fila.valor);
    } else if (fila.campo === "telefono") {
      await page.locator('input[type="tel"]').fill(fila.valor);
    }
  }
});

When("selecciona el método de pago {string}", async (metodo: string) => {
  const radio = page.locator(`label:has-text("${metodo}") input[type="radio"]`);
  await radio.check();
});

When("confirma el pedido", async () => {
  const btnConfirmar = page.locator('button[type="submit"]:has-text("Confirmar Pedido")');
  await btnConfirmar.click();
});

Then("visualiza la pantalla de confirmación con el mensaje de éxito", async () => {
  await page.waitForURL((url) => url.pathname.includes("/confirmacion"), { timeout: 20000 });
  await expect(page.locator("text=¡Pedido confirmado!").or(page.locator("text=¡Gracias por tu compra!"))).toBeVisible({ timeout: 15000 });
});

Then("puede consultar el pedido en su historial de compras", async () => {
  await page.goto(`${BASE_URL}/mis-pedidos`, { waitUntil: "networkidle" });
  await expect(page.locator("h1:has-text('Mis Pedidos')")).toBeVisible({ timeout: 15000 });
  // Verificar que hay pedidos en la lista
  const tablaOPedidos = page.locator("table, .space-y-4, .rounded-xl");
  await expect(tablaOPedidos.first()).toBeVisible({ timeout: 10000 });
});
