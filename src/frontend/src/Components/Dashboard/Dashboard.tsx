import { useEffect, useState } from "react";
import {
  api,
  type ReporteResumenDTO,
  type TopProductoDTO,
  type TopReferidoDTO,
} from "../../services/api";

/**
 * Vista de Dashboard administrativo conectada a los reportes reales de negocio (Fase 6).
 */
const Dashboard = () => {
  const [resumen, setResumen] = useState<ReporteResumenDTO | null>(null);
  const [topProductos, setTopProductos] = useState<TopProductoDTO[]>([]);
  const [topReferidos, setTopReferidos] = useState<TopReferidoDTO[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const cargarMetricas = () => {
    setLoading(true);
    setError(null);

    Promise.all([
      api.getReporteResumen(),
      api.getTopProductos(5),
      api.getTopReferidos(5),
    ])
      .then(([resumenData, prodsData, refsData]) => {
        setResumen(resumenData);
        setTopProductos(prodsData.data || []);
        setTopReferidos(refsData.data || []);
        setLoading(false);
      })
      .catch((err: unknown) => {
        setError(
          err instanceof Error ? err.message : "Error al cargar reportes de negocio"
        );
        setLoading(false);
      });
  };

  useEffect(() => {
    cargarMetricas();
  }, []);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center p-12 text-slate-500 space-y-3">
        <div className="w-8 h-8 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin" />
        <p className="font-medium">Cargando métricas de negocio...</p>
      </div>
    );
  }

  if (error || !resumen) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 p-6 rounded-xl text-center space-y-3">
        <p className="font-semibold">{error || "No se pudo cargar la información de reportes"}</p>
        <button
          onClick={cargarMetricas}
          className="px-4 py-2 bg-red-600 text-white text-sm font-semibold rounded-lg hover:bg-red-700 transition"
        >
          Reintentar
        </button>
      </div>
    );
  }

  const kpis = [
    {
      etiqueta: "Ventas Totales",
      valor: `$${resumen.totalVendido.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      color: "text-indigo-600",
      descripcion: "Volumen acumulado pagado",
    },
    {
      etiqueta: "Pedidos Procesados",
      valor: resumen.numeroPedidos.toString(),
      color: "text-blue-600",
      descripcion: "Total de órdenes registradas",
    },
    {
      etiqueta: "Ticket Promedio",
      valor: `$${resumen.ticketPromedio.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      color: "text-emerald-600",
      descripcion: "Gasto medio por orden",
    },
    {
      etiqueta: "Comisiones Devengadas",
      valor: `$${resumen.comisionesDevengadas.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      color: "text-amber-600",
      descripcion: "Repartido a red de afiliados",
    },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-800">Panel de Control & Reportes</h1>
        <p className="text-sm text-slate-500 mt-1">
          Métricas consolidadas de ventas, volumen transaccional y comisiones MLM.
        </p>
      </div>

      {/* Tarjetas de KPIs */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        {kpis.map((kpi) => (
          <div
            key={kpi.etiqueta}
            className="bg-white p-6 rounded-xl shadow-sm border border-slate-200 flex flex-col justify-between"
          >
            <div>
              <p className="text-xs text-slate-500 uppercase tracking-wider font-bold">
                {kpi.etiqueta}
              </p>
              <p className={`text-3xl font-extrabold mt-2 ${kpi.color}`}>
                {kpi.valor}
              </p>
            </div>
            <p className="text-xs text-slate-400 mt-4 border-t border-slate-100 pt-2">
              {kpi.descripcion}
            </p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Productos */}
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-200">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-slate-800">
              Top Productos Más Vendidos
            </h2>
            <span className="text-xs bg-indigo-50 text-indigo-700 px-2.5 py-1 rounded-full font-semibold">
              Por Ingresos
            </span>
          </div>

          {topProductos.length === 0 ? (
            <div className="text-center py-10 text-slate-400 text-sm">
              No hay ventas registradas aún.
            </div>
          ) : (
            <div className="space-y-3">
              {topProductos.map((prod, idx) => (
                <div
                  key={prod.productoId}
                  className="flex items-center justify-between p-3.5 rounded-lg border border-slate-100 hover:bg-slate-50 transition"
                >
                  <div className="flex items-center gap-3">
                    <span className="w-8 h-8 rounded-full bg-indigo-100 text-indigo-700 font-bold text-sm flex items-center justify-center">
                      {idx + 1}
                    </span>
                    <div>
                      <p className="font-semibold text-slate-800 text-sm">{prod.nombre}</p>
                      <p className="text-xs text-slate-500">
                        {prod.cantidad} {prod.cantidad === 1 ? "unidad vendida" : "unidades vendidas"}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-slate-800 text-sm">
                      ${prod.totalVentas.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </p>
                    <p className="text-xs text-slate-400">Total recaudado</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Top Distribuidores / Referidos */}
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-200">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-slate-800">
              Top Distribuidores de la Red
            </h2>
            <span className="text-xs bg-emerald-50 text-emerald-700 px-2.5 py-1 rounded-full font-semibold">
              Ventas de Red
            </span>
          </div>

          {topReferidos.length === 0 ? (
            <div className="text-center py-10 text-slate-400 text-sm">
              No hay distribuidores con volumen generado aún.
            </div>
          ) : (
            <div className="space-y-3">
              {topReferidos.map((ref, idx) => (
                <div
                  key={ref.referidoId}
                  className="flex items-center justify-between p-3.5 rounded-lg border border-slate-100 hover:bg-slate-50 transition"
                >
                  <div className="flex items-center gap-3">
                    <span className="w-8 h-8 rounded-full bg-emerald-100 text-emerald-700 font-bold text-sm flex items-center justify-center">
                      {idx + 1}
                    </span>
                    <div>
                      <p className="font-semibold text-slate-800 text-sm">{ref.nombre}</p>
                      <p className="text-xs text-slate-500">Nivel {ref.nivel}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-slate-800 text-sm">
                      ${ref.ventas.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </p>
                    <p className="text-xs text-emerald-600 font-semibold">Volumen de Red</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
