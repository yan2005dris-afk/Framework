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


export const redInicial: Referido = {
  id: 0,
  nombre: "Tú",
  nivel: 0,
  ventas: 2400,
  hijos: [
    {
      id: 1,
      nombre: "Ana García",
      nivel: 1,
      ventas: 1200,
      hijos: [
        {
          id: 4,
          nombre: "Carlos Ruiz",
          nivel: 2,
          ventas: 500,
          hijos: [
            { id: 7, nombre: "Diana Paz", nivel: 3, ventas: 300 },
          ],
        },
        { id: 5, nombre: "Sofía León", nivel: 2, ventas: 430 },
      ],
    },
    {
      id: 2,
      nombre: "Luis Poveda",
      nivel: 1,
      ventas: 850,
      hijos: [
        { id: 6, nombre: "Marco Díaz", nivel: 2, ventas: 380 },
      ],
    },
    { id: 3, nombre: "Marta Sánchez", nivel: 1, ventas: 430 },
  ],
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




