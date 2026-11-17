import { useState, useEffect } from "react";
import { api, type PedidoDTO, type PedidoConDetalleDTO } from "../../services/api";

const estadoColor: Record<string, string> = {
  completado: "bg-green-100 text-green-700",
  entregado: "bg-green-100 text-green-700",
  preparando: "bg-blue-100 text-blue-700",
  enviado: "bg-purple-100 text-purple-700",
  pendiente: "bg-amber-100 text-amber-700",
  cancelado: "bg-rose-100 text-rose-700",
};

export const AdminPedidos = () => {
  const [pedidos, setPedidos] = useState<PedidoDTO[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [filtroEstado, setFiltroEstado] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [detallePedido, setDetallePedido] = useState<PedidoConDetalleDTO | null>(null);
  const [cargandoDetalle, setCargandoDetalle] = useState<boolean>(false);
  const [cambiandoEstado, setCambiandoEstado] = useState<number | null>(null);

  const cargarPedidos = async (estado?: string) => {
    try {
      setLoading(true);
      setError("");
      const res = await api.getPedidosAdmin(estado);
      setPedidos(res.data || []);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Error al cargar pedidos administrativos");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    cargarPedidos(filtroEstado);
  }, [filtroEstado]);

  const verDetalle = async (id: number) => {
    try {
      setCargandoDetalle(true);
      const data = await api.getPedidoDetalle(id);
      setDetallePedido(data);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al consultar detalle");
    } finally {
      setCargandoDetalle(false);
    }
  };

  const handleCambiarEstado = async (id: number, nuevoEstado: string) => {
    if (nuevoEstado === "cancelado") {
      if (!confirm("¿Seguro que deseas cancelar este pedido? Se restaurará el stock y se revertirán comisiones.")) {
        return;
      }
    }

    try {
      setCambiandoEstado(id);
      await api.actualizarEstadoPedido(id, nuevoEstado);
      await cargarPedidos(filtroEstado);
      if (detallePedido && detallePedido.id === id) {
        setDetallePedido({ ...detallePedido, estado: nuevoEstado });
      }
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al actualizar estado");
    } finally {
      setCambiandoEstado(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">
            Gestión de Pedidos
          </h1>
          <p className="text-slate-500 text-sm mt-1">
            Supervisa el ciclo de vida de los pedidos y atiende transiciones de despacho
          </p>
        </div>

        {/* Filtro por estado */}
        <div className="flex items-center gap-2">
          <label className="text-xs font-semibold text-slate-600">Filtrar estado:</label>
          <select
            value={filtroEstado}
            onChange={(e) => setFiltroEstado(e.target.value)}
            className="px-3 py-1.5 bg-white border border-slate-300 rounded-lg text-sm outline-none focus:ring-2 focus:ring-indigo-600"
          >
            <option value="">Todos los estados</option>
            <option value="pendiente">Pendiente</option>
            <option value="preparando">Preparando</option>
            <option value="enviado">Enviado</option>
            <option value="entregado">Entregado</option>
            <option value="cancelado">Cancelado</option>
          </select>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-4 rounded-xl text-sm border border-red-200">
          {error}
        </div>
      )}

      {/* Tabla de Pedidos */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-sm">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold text-xs uppercase tracking-wider">
                <th className="py-3.5 px-4">N° Pedido</th>
                <th className="py-3.5 px-4">Cliente</th>
                <th className="py-3.5 px-4">Fecha</th>
                <th className="py-3.5 px-4">Total</th>
                <th className="py-3.5 px-4">Estado</th>
                <th className="py-3.5 px-4">Cambiar Estado</th>
                <th className="py-3.5 px-4 text-right">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 text-slate-800">
              {loading ? (
                <tr>
                  <td colSpan={7} className="py-8 text-center text-slate-400">
                    Cargando pedidos...
                  </td>
                </tr>
              ) : pedidos.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-8 text-center text-slate-400">
                    No se encontraron pedidos.
                  </td>
                </tr>
              ) : (
                pedidos.map((p) => {
                  const numero = `MC-${p.id.toString().padStart(6, "0")}`;
                  const fecha = new Date(p.creadoAt).toLocaleDateString("es-ES", {
                    day: "2-digit",
                    month: "short",
                    year: "numeric",
                  });
                  const badgeClass = estadoColor[p.estado] || "bg-slate-100 text-slate-700";

                  return (
                    <tr key={p.id} className="hover:bg-slate-50/75 transition">
                      <td className="py-3 px-4 font-bold text-slate-900">{numero}</td>
                      <td className="py-3 px-4">
                        <p className="font-medium text-slate-900">{p.nombreCliente}</p>
                        <p className="text-xs text-slate-400">{p.emailCliente}</p>
                      </td>
                      <td className="py-3 px-4 text-xs text-slate-500">{fecha}</td>
                      <td className="py-3 px-4 font-bold text-slate-900">${p.total.toFixed(2)}</td>
                      <td className="py-3 px-4">
                        <span className={`text-xs px-2.5 py-1 rounded-full font-semibold capitalize ${badgeClass}`}>
                          {p.estado}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <select
                          value={p.estado}
                          disabled={p.estado === "cancelado" || cambiandoEstado === p.id}
                          onChange={(e) => handleCambiarEstado(p.id, e.target.value)}
                          className="text-xs px-2 py-1 bg-white border border-slate-300 rounded font-medium outline-none focus:ring-1 focus:ring-indigo-600 disabled:opacity-50"
                        >
                          <option value="pendiente">Pendiente</option>
                          <option value="preparando">Preparando</option>
                          <option value="enviado">Enviado</option>
                          <option value="entregado">Entregado</option>
                          <option value="cancelado">Cancelar</option>
                        </select>
                      </td>
                      <td className="py-3 px-4 text-right">
                        <button
                          onClick={() => verDetalle(p.id)}
                          className="text-indigo-600 hover:text-indigo-900 font-semibold text-xs px-2 py-1 rounded hover:bg-indigo-50 transition"
                        >
                          Ver ítems
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Modal Detalle */}
      {detallePedido && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-lg w-full p-6 border border-slate-200 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-100 pb-3">
              <div>
                <h3 className="text-lg font-bold text-slate-900">
                  Pedido MC-{detallePedido.id.toString().padStart(6, "0")}
                </h3>
                <p className="text-xs text-slate-500">Cliente: {detallePedido.nombreCliente} ({detallePedido.emailCliente})</p>
              </div>
              <button
                onClick={() => setDetallePedido(null)}
                className="text-slate-400 hover:text-slate-700 text-sm font-bold p-1"
              >
                ✕
              </button>
            </div>

            <div className="divide-y divide-slate-100 max-h-60 overflow-y-auto">
              {cargandoDetalle ? (
                <p className="text-center py-4 text-sm text-slate-400">Cargando...</p>
              ) : (
                detallePedido.items?.map((it) => (
                  <div key={it.id} className="py-2.5 flex items-center justify-between text-sm">
                    <div className="flex items-center gap-3">
                      {it.imgProducto && (
                        <img
                          src={it.imgProducto}
                          alt={it.nombreProducto}
                          className="w-10 h-10 rounded object-cover border border-slate-200"
                        />
                      )}
                      <div>
                        <p className="font-semibold text-slate-800">{it.nombreProducto}</p>
                        <p className="text-xs text-slate-400">
                          {it.cantidad} unidad(es) x ${it.precioUnitario.toFixed(2)}
                        </p>
                      </div>
                    </div>
                    <span className="font-bold text-slate-900">${it.subtotal.toFixed(2)}</span>
                  </div>
                ))
              )}
            </div>

            <div className="border-t border-slate-100 pt-3 flex justify-between items-center text-sm font-bold">
              <span>Total Liquidado:</span>
              <span className="text-lg text-indigo-600">${detallePedido.total.toFixed(2)}</span>
            </div>

            <div className="flex justify-end pt-2">
              <button
                onClick={() => setDetallePedido(null)}
                className="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-800 rounded-lg text-sm font-medium transition"
              >
                Cerrar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminPedidos;
