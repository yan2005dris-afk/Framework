import type { CheckoutRequestDTO } from "@/types/checkout";

/**
 * Cola de checkouts pendientes (U4-T3, S9).
 *
 * Cuando el POST /api/checkout falla por red, se guarda el intento con su
 * `Idempotency-Key` y se reintenta solo cuando vuelve la conexion. La key es
 * lo que hace que esto sea seguro: el Backend devuelve el pedido original si
 * esa key ya se uso, asi que reintentar no puede crear un segundo pedido.
 *
 * En localStorage y no en memoria a proposito: el caso que mas importa es el
 * de cerrar la app o que se caiga la pestana con el pedido a medio enviar, y
 * un estado en memoria se pierde con la pestana.
 */

const CLAVE_ALMACEN = "multicatalogo.checkout.pendientes";

export interface IntentoPendiente {
  idempotencyKey: string;
  payload: CheckoutRequestDTO;
  guardadoAt: string;
}

/**
 * Cuantos intentos se guardan como maximo. Es un freno: si el Backend acepta
 * la peticion pero la respuesta nunca llega, cada intento queda en la cola
 * para siempre y sin tope la cola crece sin parar.
 */
const MAX_INTENTOS = 5;

export const leerIntentos = (): IntentoPendiente[] => {
  try {
    const crudo = localStorage.getItem(CLAVE_ALMACEN);
    if (!crudo) return [];
    const parsed = JSON.parse(crudo);
    return Array.isArray(parsed) ? (parsed as IntentoPendiente[]) : [];
  } catch {
    // localStorage puede estar bloqueado (modo privado) o corrupto. Perder la
    // cola no puede romper el checkout, asi que se degrada a "no hay nada".
    return [];
  }
};

const escribir = (intentos: IntentoPendiente[]): void => {
  try {
    if (intentos.length === 0) {
      localStorage.removeItem(CLAVE_ALMACEN);
      return;
    }
    localStorage.setItem(CLAVE_ALMACEN, JSON.stringify(intentos.slice(-MAX_INTENTOS)));
  } catch {
    // Mismo criterio que arriba: la cola es una mejora, no un requisito.
  }
};

/**
 * Genera la key de un intento nuevo. `crypto.randomUUID` existe en todos los
 * navegadores con service worker, o sea exactamente los que llegan aca.
 */
export const nuevaIdempotencyKey = (): string =>
  typeof crypto !== "undefined" && "randomUUID" in crypto
    ? crypto.randomUUID()
    : `k-${Date.now()}-${Math.random().toString(36).slice(2)}`;

export const guardarIntento = (intento: IntentoPendiente): void => {
  const actuales = leerIntentos().filter((i) => i.idempotencyKey !== intento.idempotencyKey);
  escribir([...actuales, intento]);
};

export const borrarIntento = (idempotencyKey: string): void => {
  escribir(leerIntentos().filter((i) => i.idempotencyKey !== idempotencyKey));
};