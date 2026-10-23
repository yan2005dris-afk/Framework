import { useState, useEffect, type FormEvent } from "react";
import { api, type ProductoAdminDTO, type GuardarProductoPayload } from "../../services/api";

/**
 * Panel de administración de productos con búsqueda, edición, alta y borrado lógico.
 */
export const AdminProductos = () => {
  const [productos, setProductos] = useState<ProductoAdminDTO[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [busqueda, setBusqueda] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [modalAbierto, setModalAbierto] = useState<boolean>(false);
  const [editandoId, setEditandoId] = useState<number | null>(null);

  // Form state
  const [nombre, setNombre] = useState<string>("");
  const [descripcion, setDescripcion] = useState<string>("");
  const [precio, setPrecio] = useState<string>("");
  const [categoria, setCategoria] = useState<string>("");
  const [img, setImg] = useState<string>("");
  const [stock, setStock] = useState<string>("100");
  const [activo, setActivo] = useState<boolean>(true);
  const [guardando, setGuardando] = useState<boolean>(false);

  const cargarProductos = async (q?: string) => {
    try {
      setLoading(true);
      setError("");
      const data = await api.getProductosAdmin(q);
      setProductos(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Error al cargar productos");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    cargarProductos(busqueda);
  }, [busqueda]);

  const abrirModalNuevo = () => {
    setEditandoId(null);
    setNombre("");
    setDescripcion("");
    setPrecio("");
    setCategoria("Serum");
    setImg("https://picsum.photos/seed/producto/600");
    setStock("100");
    setActivo(true);
    setModalAbierto(true);
  };

  const abrirModalEditar = (p: ProductoAdminDTO) => {
    setEditandoId(p.id);
    setNombre(p.nombre);
    setDescripcion(p.descripcion);
    setPrecio(p.precio.toString());
    setCategoria(p.categoria);
    setImg(p.img);
    setStock(p.stock.toString());
    setActivo(p.activo);
    setModalAbierto(true);
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setGuardando(true);
    try {
      const payload: GuardarProductoPayload = {
        nombre,
        descripcion,
        precio: parseFloat(precio),
        categoria,
        img: img || "https://picsum.photos/seed/producto/600",
        stock: parseInt(stock, 10) || 0,
        activo,
      };

      if (editandoId) {
        await api.actualizarProducto(editandoId, payload);
      } else {
        await api.crearProducto(payload);
      }

      setModalAbierto(false);
      await cargarProductos(busqueda);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al guardar el producto");
    } finally {
      setGuardando(false);
    }
  };

  const toggleActivo = async (p: ProductoAdminDTO) => {
    const nuevoEstado = !p.activo;
    const confirmMsg = nuevoEstado
      ? `¿Deseas reactivar el producto "${p.nombre}"?`
      : `¿Deseas desactivar el producto "${p.nombre}" del catálogo público?`;

    if (!confirm(confirmMsg)) return;

    try {
      if (!nuevoEstado) {
        await api.eliminarProductoLogico(p.id);
      } else {
        await api.actualizarProducto(p.id, { activo: true });
      }
      await cargarProductos(busqueda);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al actualizar estado");
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">
            Gestión de Productos
          </h1>
          <p className="text-slate-500 text-sm mt-1">
            Administra el catálogo, existencias y estados de disponibilidad
          </p>
        </div>

        <button
          onClick={abrirModalNuevo}
          className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold px-4 py-2.5 rounded-lg text-sm transition flex items-center justify-center gap-2 self-start sm:self-auto min-h-[44px]"
        >
          <span>+ Nuevo Producto</span>
        </button>
      </div>

      {/* Barra de búsqueda */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm flex items-center gap-3">
        <input
          type="text"
          value={busqueda}
          onChange={(e) => setBusqueda(e.target.value)}
          placeholder="Buscar por nombre o categoría (búsqueda server-side ILIKE)..."
          className="w-full px-4 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none text-sm"
        />
        {busqueda && (
          <button
            onClick={() => setBusqueda("")}
            className="text-xs text-slate-500 hover:text-slate-800 px-2 py-1"
          >
            Limpiar
          </button>
        )}
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-4 rounded-lg text-sm border border-red-200">
          {error}
        </div>
      )}

      {/* Tabla de Productos */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-sm">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold text-xs uppercase tracking-wider">
                <th className="py-3.5 px-4">Producto</th>
                <th className="py-3.5 px-4">Categoría</th>
                <th className="py-3.5 px-4">Precio</th>
                <th className="py-3.5 px-4">Stock</th>
                <th className="py-3.5 px-4">Estado</th>
                <th className="py-3.5 px-4 text-right">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 text-slate-800">
              {loading ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-slate-400">
                    Cargando catálogo administrativo...
                  </td>
                </tr>
              ) : productos.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-slate-400">
                    No se encontraron productos con el filtro aplicado.
                  </td>
                </tr>
              ) : (
                productos.map((p) => (
                  <tr key={p.id} className="hover:bg-slate-50/75 transition">
                    <td className="py-3 px-4 flex items-center gap-3">
                      <img
                        src={p.img}
                        alt={p.nombre}
                        className="w-10 h-10 rounded-lg object-cover border border-slate-200 bg-slate-100"
                        onError={(e) => {
                          (e.currentTarget as HTMLImageElement).src =
                            "https://picsum.photos/seed/fallback/100";
                        }}
                      />
                      <div>
                        <p className="font-semibold text-slate-900">{p.nombre}</p>
                        <p className="text-xs text-slate-400 line-clamp-1 max-w-xs">
                          {p.descripcion}
                        </p>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <span className="bg-slate-100 text-slate-700 text-xs px-2.5 py-1 rounded-full font-medium">
                        {p.categoria}
                      </span>
                    </td>
                    <td className="py-3 px-4 font-semibold text-slate-900">
                      ${p.precio.toFixed(2)}
                    </td>
                    <td className="py-3 px-4">
                      <span
                        className={`text-xs font-semibold px-2 py-0.5 rounded ${
                          p.stock > 10
                            ? "bg-slate-100 text-slate-700"
                            : p.stock > 0
                            ? "bg-amber-100 text-amber-800"
                            : "bg-red-100 text-red-800"
                        }`}
                      >
                        {p.stock} unidades
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span
                        className={`text-xs px-2.5 py-1 rounded-full font-semibold ${
                          p.activo
                            ? "bg-green-100 text-green-700"
                            : "bg-rose-100 text-rose-700"
                        }`}
                      >
                        {p.activo ? "Activo" : "Inactivo"}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right space-x-2">
                      <button
                        onClick={() => abrirModalEditar(p)}
                        className="text-indigo-600 hover:text-indigo-900 font-medium text-xs px-2 py-1 rounded hover:bg-indigo-50 transition"
                      >
                        Editar
                      </button>
                      <button
                        onClick={() => toggleActivo(p)}
                        className={`font-medium text-xs px-2 py-1 rounded transition ${
                          p.activo
                            ? "text-rose-600 hover:bg-rose-50"
                            : "text-green-600 hover:bg-green-50"
                        }`}
                      >
                        {p.activo ? "Desactivar" : "Reactivar"}
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Modal Crear / Editar */}
      {modalAbierto && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-lg w-full p-6 border border-slate-200">
            <h3 className="text-lg font-bold text-slate-900 mb-4">
              {editandoId ? "Editar Producto" : "Nuevo Producto"}
            </h3>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">
                  Nombre del Producto
                </label>
                <input
                  type="text"
                  value={nombre}
                  onChange={(e) => setNombre(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-600 outline-none"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">
                  Descripción
                </label>
                <textarea
                  value={descripcion}
                  onChange={(e) => setDescripcion(e.target.value)}
                  rows={2}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-600 outline-none"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">
                    Precio ($)
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    min="0.01"
                    value={precio}
                    onChange={(e) => setPrecio(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-600 outline-none"
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">
                    Stock
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={stock}
                    onChange={(e) => setStock(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-600 outline-none"
                    required
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">
                    Categoría
                  </label>
                  <input
                    type="text"
                    value={categoria}
                    onChange={(e) => setCategoria(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-600 outline-none"
                    placeholder="Serum, Crema, etc."
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">
                    URL Imagen
                  </label>
                  <input
                    type="url"
                    value={img}
                    onChange={(e) => setImg(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-600 outline-none"
                    placeholder="https://..."
                  />
                </div>
              </div>

              <div className="flex items-center gap-2 pt-2">
                <input
                  type="checkbox"
                  id="activoCheckbox"
                  checked={activo}
                  onChange={(e) => setActivo(e.target.checked)}
                  className="rounded text-indigo-600 focus:ring-indigo-500 h-4 w-4"
                />
                <label htmlFor="activoCheckbox" className="text-sm font-medium text-slate-700">
                  Visible en Catálogo Público (Activo)
                </label>
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-slate-100">
                <button
                  type="button"
                  onClick={() => setModalAbierto(false)}
                  className="px-4 py-2 border border-slate-300 rounded-lg text-sm text-slate-700 hover:bg-slate-50 transition"
                  disabled={guardando}
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={guardando}
                  className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-semibold transition"
                >
                  {guardando ? "Guardando..." : "Guardar Producto"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminProductos;
