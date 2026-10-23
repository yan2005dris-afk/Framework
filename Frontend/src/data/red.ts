export interface Referido {
  id: number;
  nombre: string;
  nivel: number;
  ventas: number;
  hijos?: Referido[];
}

export const TASA_COMISION: Record<number, number> = {
  1: 0.10,
  2: 0.05,
  3: 0.02,
};


export const contarRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijos) => total + 1 + contarRed(hijos), 0);
}


export const sumarVentasRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + hijo.ventas + sumarVentasRed(hijo), 0)
}

export const comisionDeReferido = (ref: Referido): number =>
  ref.ventas * (TASA_COMISION[ref.nivel] ?? 0);

export const sumarComisiones = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce(
    (total, hijo) => total + comisionDeReferido(hijo) + sumarComisiones(hijo), 0
  );
};


export const nivelAlcanzado = (referidosDirectos: number): string => {
  if (referidosDirectos >= 6) return "Diamante";
  if (referidosDirectos >= 4) return "Oro";
  if (referidosDirectos >= 2) return "Plata";
  return "Bronce";
};




