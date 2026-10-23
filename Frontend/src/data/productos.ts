export interface Producto {
  id: number;
  nombre: string;
  descripcion: string;
  precio: number;
  categoria: string;
  img: string;
  galeria: string[];
}


export const categorias = [
  "Serum",
  "Crema",
  "Tonico",
  "Mascarilla",
  "Kit",
]