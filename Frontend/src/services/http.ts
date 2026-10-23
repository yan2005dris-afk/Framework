/**
 * Cliente HTTP centralizado.
 *
 * Política:
 * - `credentials: "include"` por defecto para que el browser adjunte la
 *   cookie httpOnly `auth_token` configurada por /api/login. Los endpoints
 *   públicos (/api/login, /api/productos) siguen funcionando sin cookie.
 * - Cuerpo de respuesta: si es JSON, se devuelve parseado; si no, texto.
 * - Errores: se transforma toda respuesta no-2xx en `ApiError` con status,
 *   message y `detalle` opcional (cuando el backend lo expone en dev).
 * - 401 dispara `authOnUnauthorized()` si se pasa en `init`, para que el
 *   frontend pueda limpiar la sesión y redirigir al login.
 */

/**
 * Base por defecto: ruta RELATIVA, no URL absoluta.
 *
 * Con `/api` la peticion sale del mismo origen que el SPA, asi que no hay CORS y
 * la cookie httpOnly `auth_token` (host-only, SameSite=lax) se emite y vuelve en
 * el mismo host. En dev el proxy de Vite la reenvia al gateway; con el build
 * estatico, el gateway sirve SPA y API en el mismo origen.
 *
 * Una URL absoluta como `http://localhost:3000/api` (el Backend directo) rompe:
 * el browser la ve como otro origen y la bloquea por Same Origin Policy.
 */
const DEFAULT_API_BASE_URL = "/api";

export const apiBaseURL =
  (import.meta.env.VITE_API_URL as string | undefined) || DEFAULT_API_BASE_URL;

/**
 * ApiError tipado con `status`, `message` y opcional `detalle`.
 * Reemplaza los `Error` planos que el código legacy lanzaba antes.
 */
export class ApiError extends Error {
  public readonly status: number;
  public readonly detalle?: unknown;

  constructor(status: number, message: string, detalle?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.detalle = detalle;
  }
}

/**
 * Error de RED: no hubo respuesta del servidor (se cortó la conexión, no hay
 * señal, el servidor no levanta). Es distinto de un `ApiError`, que significa
 * que el servidor respondió y respondió con un status.
 *
 * La diferencia importa porque un error de red deja el request en un estado
 * indeterminado: pudo no haber llegado, o haber llegado y haberse perdido la
 * respuesta. Solo en ese caso tiene sentido reintentar (U4-T3, S9).
 */
export class NetworkError extends Error {
  // Campo declarado a mano y no como parameter property: el tsconfig tiene
  // `erasableSyntaxOnly`, que prohibe la sintaxis de TypeScript que no se
  // puede borrar al compilar.
  readonly causa: unknown;

  constructor(causa: unknown) {
    super("No se pudo conectar con el servidor");
    this.name = "NetworkError";
    this.causa = causa;
  }
}

export interface ApiFetchInit extends Omit<RequestInit, "body" | "credentials"> {
  /** Cuerpo JSON. Si está presente, se serializa y se setea Content-Type. */
  json?: unknown;
  /** Cuerpo crudo. Útil si mandás FormData, Blob, etc. */
  body?: BodyInit | null;
  /** Si se pasa, se invoca al recibir 401 (sesión expirada). */
  onUnauthorized?: () => void;
}

const defaultInit: RequestInit = {
  credentials: "include",
};

/**
 * apiFetch es el cliente único del frontend. Por siempre se debe usar este
 * para hablar con el backend; nunca `fetch` directo.
 *
 * @param path Ruta absoluta o relativa a apiBaseURL.
 * @param init Opciones: `json`, `body`, `headers`, `onUnauthorized`, etc.
 *             Por defecto `credentials: "include"` (cookie httpOnly).
 */
export async function apiFetch<T = unknown>(
  path: string,
  init: ApiFetchInit = {}
): Promise<T> {
  const url = path.startsWith("http")
    ? path
    : `${apiBaseURL}${path.startsWith("/") ? "" : "/"}${path}`;

  const headers: HeadersInit = {
    Accept: "application/json",
    ...(init.headers || {}),
  };

  let body: BodyInit | undefined;
  if (init.json !== undefined) {
    body = JSON.stringify(init.json);
    (headers as Record<string, string>)["Content-Type"] = "application/json";
  } else {
    body = init.body ?? undefined;
  }

  let response: Response;
  try {
    response = await fetch(url, {
      ...defaultInit,
      ...init,
      headers,
      body,
      credentials: "include",
    });
  } catch (causa) {
    throw new NetworkError(causa);
  }

  if (response.status === 401) {
    init.onUnauthorized?.();
    // Aún así, parseamos el cuerpo para propagar el mensaje si lo necesitamos.
  }

  const text = await response.text();
  let parsed: unknown = null;
  if (text.length > 0) {
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = text;
    }
  }

  if (!response.ok) {
    const message = extractMessage(parsed) || `HTTP ${response.status}`;
    const detalle = extractDetalle(parsed);
    throw new ApiError(response.status, message, detalle);
  }

  return parsed as T;
}

function extractMessage(body: unknown): string {
  if (body && typeof body === "object") {
    const o = body as Record<string, unknown>;
    if (typeof o.error === "string") return o.error;
    if (typeof o.message === "string") return o.message;
  }
  if (typeof body === "string") return body;
  return "";
}

function extractDetalle(body: unknown): unknown {
  if (body && typeof body === "object") {
    const o = body as Record<string, unknown>;
    if ("detalle" in o) return o.detalle;
  }
  return undefined;
}
