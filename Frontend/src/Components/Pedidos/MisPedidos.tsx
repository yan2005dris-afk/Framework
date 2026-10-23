import { useState, useEffect } from "react";
import { Link } from "@tanstack/react-router";
import { api, type PedidoDTO, type PedidoConDetalleDTO } from "../../services/api";

const estadoColor: Record<string, string> = {
  completado: "bg-green-100 text-green-700",
  entregado: "bg-green-100 text-green-700",
  preparando: "bg-blue-100 text-blue-700",
  enviado: "bg-purple-100 text-purple-700",
  pendiente: "bg-amber-100 text-amber-700",
  cancelado: "bg-rose-100 text-rose-700",
};

export const MisPedidos = () => {
  const [pedidos, setPedidos] = useState<PedidoDTO[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>("");
  const [detalleAbierto, setDetalleAbierto] = useState<number | null>(null);
  const [detallePedido, setDetallePedido] = useState<PedidoConDetalleDTO | null>(null);
  const [cargandoDetalle, setCargandoDetalle] = useState<boolean>(false);

  const cargarPedidos = async () => {
    try {
      setLoading(true);
      setError("");
      const res = await api.getPedidosMios();
      setPedidos(res.data || []);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Error al consultar el historial de pedidos");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    cargarPedidos();
  }, []);

  const toggleDetalle = async (id: number) => {
    if (detalleAbierto === id) {
      setDetalleAbierto(null);
      setDetallePedido(null);
      return;
    }

    try {
      setCargandoDetalle(true);
      setDetalleAbierto(id);
      const data = await api.getPedidoDetalle(id);
      setDetallePedido(data);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al obtener detalle del pedido");
    } finally {
      setCargandoDetalle(false);
    }
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">
          Mis Pedidos
        </h1>
        <p className="text-slate-500 text-sm mt-1">
          Historial de compras realizadas y seguimiento del estado de tus envíos
        </p>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-4 rounded-xl text-sm border border-red-200">
          {error}
        </div>
      )}

      {loading ? (
        <div className="bg-white p-8 rounded-xl border border-slate-200 text-center text-slate-400">
          Cargando tus compras...
        </div>
      ) : pedidos.length === 0 ? (
        <div className="bg-white p-12 rounded-xl border border-slate-200 text-center space-y-3">
          <p className="text-slate-600 font-medium">Aún no has realizado ningún pedido.</p>
          <Link
            to="/tienda"
            className="inline-block bg-indigo-600 text-white font-semibold px-4 py-2 rounded-lg text-sm hover:bg-indigo-700 transition"
          >
            Explorar Tienda
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {pedidos.map((p) => {
            const numero = `MC-${p.id.toString().padStart(6, "0")}`;
            const fecha = new Date(p.creadoAt).toLocaleDateString("es-ES", {
              day: "2-digit",
              month: "short",
              year: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            });
            const badgeClass = estadoColor[p.estado] || "bg-slate-100 text-slate-700";
            const estaExpandido = detalleAbierto === p.id;

            return (
              <div
                key={p.id}
                className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden transition"
              >
                <div className="p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div className="space-y-1">
                    <div className="flex items-center gap-3">
                      <span className="font-bold text-slate-900">{numero}</span>
                      <span className={`text-xs px-2.5 py-0.5 rounded-full font-semibold capitalize ${badgeClass}`}>
                        {p.estado}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500">Fecha: {fecha}</p>
                    <p className="text-xs text-slate-500">Pago: {p.metodoPago}</p>
                  </div>

                  <div className="flex items-center justify-between sm:justify-end gap-4 border-t sm:border-t-0 pt-3 sm:pt-0">
                    <div className="text-right">
                      <p className="text-xs text-slate-400">Total</p>
                      <p className="text-lg font-bold text-slate-900">${p.total.toFixed(2)}</p>
                    </div>

                    <button
                      onClick={() => toggleDetalle(p.id)}
                      className="px-3 py-1.5 border border-slate-200 hover:bg-slate-50 text-indigo-600 font-medium text-xs rounded-lg transition"
                    >
                      {estaExpandido ? "Ocultar" : "Ver detalle"}
                    </button>
                  </div>
                </div>

                {/* Desglose Expandible */}
                {estaExpandido && (
                  <div className="bg-slate-50 border-t border-slate-100 p-5 space-y-3">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                      Artículos comprados
                    </h4>

                    {cargandoDetalle ? (
                      <p className="text-xs text-slate-400">Cargando ítems...</p>
                    ) : detallePedido && detallePedido.items?.length > 0 ? (
                      <div className="divide-y divide-slate-200/60">
                        {detallePedido.items.map((item) => (
                          <div key={item.id} className="py-2 flex items-center justify-between text-sm">
                            <div className="flex items-center gap-3">
                              {item.imgProducto && (
                                <img
                                  src={item.imgProducto}
                                  alt={item.nombreProducto}
                                  className="w-10 h-10 rounded object-cover border border-slate-200 bg-white"
                                />
                              )}
                              <div>
                                <p className="font-medium text-slate-800">{item.nombreProducto}</p>
                                <p className="text-xs text-slate-400">
                                  {item.cantidad} x ${item.precioUnitario.toFixed(2)}
                                </p>
                              </div>
                            </div>
                            <span className="font-semibold text-slate-900">
                              ${item.subtotal.toFixed(2)}
                            </span>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="text-xs text-slate-400">No hay detalles registrados.</p>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default MisPedidos;
