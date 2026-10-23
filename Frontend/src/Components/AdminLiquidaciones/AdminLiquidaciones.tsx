import { useState, useEffect } from "react";
import {
  api,
  type DeudaDistribuidorDTO,
  type LiquidacionDTO,
} from "../../services/api";

export const AdminLiquidaciones = () => {
  const [deudas, setDeudas] = useState<DeudaDistribuidorDTO[]>([]);
  const [historial, setHistorial] = useState<LiquidacionDTO[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>("");
  const [liquidandoId, setLiquidandoId] = useState<number | null>(null);

  const cargarDatos = async () => {
    try {
      setLoading(true);
      setError("");
      const [resDeudas, resHistorial] = await Promise.all([
        api.getComisionesPendientes(),
        api.getLiquidacionesAdmin(),
      ]);
      setDeudas(resDeudas.data || []);
      setHistorial(resHistorial.liquidaciones || []);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Error al consultar módulo de liquidaciones");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    cargarDatos();
  }, []);

  const handleLiquidar = async (d: DeudaDistribuidorDTO) => {
    if (
      !confirm(
        `¿Confirmas la liquidación de $${d.totalPendiente.toFixed(2)} para ${d.nombre}? Las comisiones se marcarán como pagadas.`
      )
    ) {
      return;
    }

    try {
      setLiquidandoId(d.referidoId);
      const res = await api.liquidarComisiones(d.referidoId);
      alert(`¡${res.mensaje}! Liquidación N° ${res.liquidacion.id} generada.`);
      await cargarDatos();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al procesar liquidación");
    } finally {
      setLiquidandoId(null);
    }
  };

  const totalDeudaGlobal = deudas.reduce((acc, curr) => acc + curr.totalPendiente, 0);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">
            Liquidaciones a Distribuidores
          </h1>
          <p className="text-slate-500 text-sm mt-1">
            Consolida y procesa los pagos de comisiones pendientes devengadas por la red
          </p>
        </div>

        <div className="bg-indigo-50 border border-indigo-200 px-4 py-2.5 rounded-xl text-right">
          <p className="text-xs uppercase font-semibold text-indigo-600">Total Deuda Acumulada</p>
          <p className="text-xl font-bold text-indigo-950">${totalDeudaGlobal.toFixed(2)}</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-4 rounded-xl text-sm border border-red-200">
          {error}
        </div>
      )}

      {/* Tabla 1: Deudas Pendientes de Liquidar */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div className="p-4 border-b border-slate-100 flex items-center justify-between">
          <h2 className="font-bold text-slate-900 text-base">Comisiones Pendientes de Pago</h2>
          <span className="text-xs bg-amber-100 text-amber-800 font-semibold px-2.5 py-0.5 rounded-full">
            {deudas.length} distribuidores por liquidar
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-sm">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold text-xs uppercase tracking-wider">
                <th className="py-3.5 px-4">Distribuidor</th>
                <th className="py-3.5 px-4">N° Comisiones</th>
                <th className="py-3.5 px-4">Saldo Pendiente</th>
                <th className="py-3.5 px-4 text-right">Acción</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 text-slate-800">
              {loading ? (
                <tr>
                  <td colSpan={4} className="py-8 text-center text-slate-400">
                    Calculando saldos pendientes...
                  </td>
                </tr>
              ) : deudas.length === 0 ? (
                <tr>
                  <td colSpan={4} className="py-8 text-center text-slate-400">
                    No hay comisiones pendientes de liquidar. ¡Todo al día!
                  </td>
                </tr>
              ) : (
                deudas.map((d) => (
                  <tr key={d.referidoId} className="hover:bg-slate-50/75 transition">
                    <td className="py-3.5 px-4">
                      <p className="font-semibold text-slate-900">{d.nombre}</p>
                      <p className="text-xs text-slate-400">ID Referido: {d.referidoId}</p>
                    </td>
                    <td className="py-3.5 px-4">
                      <span className="bg-slate-100 text-slate-700 px-2 py-0.5 rounded text-xs font-medium">
                        {d.numeroComisiones} acumuladas
                      </span>
                    </td>
                    <td className="py-3.5 px-4 font-bold text-green-600 text-base">
                      ${d.totalPendiente.toFixed(2)}
                    </td>
                    <td className="py-3.5 px-4 text-right">
                      <button
                        onClick={() => handleLiquidar(d)}
                        disabled={liquidandoId === d.referidoId}
                        className="bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-semibold text-xs px-3.5 py-2 rounded-lg transition min-h-[36px]"
                      >
                        {liquidandoId === d.referidoId ? "Procesando..." : "Liquidar Pago"}
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Tabla 2: Historial de Liquidaciones Procesadas */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div className="p-4 border-b border-slate-100">
          <h2 className="font-bold text-slate-900 text-base">Historial de Liquidaciones</h2>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-sm">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold text-xs uppercase tracking-wider">
                <th className="py-3.5 px-4">N° Liq</th>
                <th className="py-3.5 px-4">ID Distribuidor</th>
                <th className="py-3.5 px-4">Fecha Pago</th>
                <th className="py-3.5 px-4">Monto Liquidado</th>
                <th className="py-3.5 px-4 text-right">Estado</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 text-slate-800">
              {loading ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-slate-400">
                    Cargando historial...
                  </td>
                </tr>
              ) : historial.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-slate-400">
                    No se han registrado pagos liquidados aún.
                  </td>
                </tr>
              ) : (
                historial.map((l) => {
                  const fecha = new Date(l.pagadaEn).toLocaleDateString("es-ES", {
                    day: "2-digit",
                    month: "short",
                    year: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                  });
                  return (
                    <tr key={l.id} className="hover:bg-slate-50/75 transition">
                      <td className="py-3 px-4 font-bold text-slate-900">
                        #{l.id.toString().padStart(5, "0")}
                      </td>
                      <td className="py-3 px-4 text-slate-600">Referido {l.referidoId}</td>
                      <td className="py-3 px-4 text-xs text-slate-500">{fecha}</td>
                      <td className="py-3 px-4 font-bold text-slate-900">${l.total.toFixed(2)}</td>
                      <td className="py-3 px-4 text-right">
                        <span className="bg-green-100 text-green-700 text-xs px-2.5 py-0.5 rounded-full font-semibold capitalize">
                          {l.estado}
                        </span>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default AdminLiquidaciones;
