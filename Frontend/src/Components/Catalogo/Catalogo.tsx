// src/Components/Catalogo/Catalogo.tsx
import { useEffect, useState } from "react";
import { useCart } from "../../Context/CartContext";
import { api, type Producto } from "../../services/api";
import { Icon } from "../Share/Icon";

const Catalogo = () => {
  const { addToCart } = useCart();
  const [productos, setProductos] = useState<Producto[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProductos = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getProductos();
      setProductos(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Error al cargar los productos");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProductos();
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">
            Catálogo de Productos
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Explora nuestro catálogo conectado a la API de Go (Fiber).
          </p>
        </div>

        <button
          onClick={fetchProductos}
          className="self-start sm:self-auto px-3 py-2 text-sm bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg transition-colors flex items-center gap-2"
          title="Recargar productos"
        >
          <Icon name="rotate-cw" size={16} />
          Actualizar
        </button>
      </div>

      {/* Estado de carga */}
      {loading && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((n) => (
            <div
              key={n}
              className="bg-white rounded-xl border border-slate-200 shadow-sm p-4 animate-pulse"
            >
              <div className="w-full h-40 bg-slate-200 rounded-lg mb-4" />
              <div className="h-4 bg-slate-200 rounded w-3/4 mb-2" />
              <div className="h-4 bg-slate-200 rounded w-1/4 mb-4" />
              <div className="h-9 bg-slate-200 rounded" />
            </div>
          ))}
        </div>
      )}

      {/* Estado de error */}
      {error && !loading && (
        <div className="bg-red-50 border border-red-200 text-red-700 p-6 rounded-xl text-center space-y-3">
          <p className="font-semibold">{error}</p>
          <p className="text-sm text-red-500">
            Verifica que el servidor backend de Go esté corriendo en el puerto 3000.
          </p>
          <button
            onClick={fetchProductos}
            className="px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 transition"
          >
            Reintentar
          </button>
        </div>
      )}

      {/* Lista de productos */}
      {!loading && !error && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {productos.map((prod) => (
            <div
              key={prod.id}
              className="bg-white rounded-xl overflow-hidden border border-slate-200 shadow-sm hover:shadow-md transition-shadow flex flex-col"
            >
              <img
                src={prod.img}
                alt={prod.nombre}
                className="w-full h-44 object-cover"
                loading="lazy"
              />
              <div className="p-4 flex flex-col flex-1">
                <h3 className="font-semibold text-slate-800 text-base">
                  {prod.nombre}
                </h3>
                <p className="text-indigo-600 font-bold text-lg mt-2 mb-4">
                  ${prod.precio.toFixed(2)}
                </p>
                <button
                  onClick={() => addToCart(prod)}
                  className="mt-auto w-full bg-slate-900 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-600 transition-colors flex items-center justify-center gap-2 cursor-pointer"
                >
                  <Icon name="shopping-bag" size={16} />
                  Añadir al Carrito
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Catalogo;