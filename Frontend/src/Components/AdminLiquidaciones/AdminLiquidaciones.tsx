import { useState, useEffect } from "react";
import {
  api,
  type DeudaDistribuidorDTO,
  type LiquidacionDTO,
} from "../../services/api";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AlertCircle, CheckCircle2, DollarSign, Loader2 } from "lucide-react";

export const AdminLiquidaciones = () => {
  const [deudas, setDeudas] = useState<DeudaDistribuidorDTO[]>([]);
  const [historial, setHistorial] = useState<LiquidacionDTO[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>("");
  const [liquidandoId, setLiquidandoId] = useState<number | null>(null);

  // Modal confirmación y éxito
  const [selectedDeuda, setSelectedDeuda] = useState<DeudaDistribuidorDTO | null>(null);
  const [confirmOpen, setConfirmOpen] = useState<boolean>(false);
  const [modalFeedback, setModalFeedback] = useState<{
    open: boolean;
    title: string;
    description: string;
    isError?: boolean;
  }>({ open: false, title: "", description: "" });

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

  const handleOpenConfirm = (d: DeudaDistribuidorDTO) => {
    setSelectedDeuda(d);
    setConfirmOpen(true);
  };

  const handleConfirmarLiquidacion = async () => {
    if (!selectedDeuda) return;
    const target = selectedDeuda;

    try {
      setLiquidandoId(target.referidoId);
      const res = await api.liquidarComisiones(target.referidoId);
      setConfirmOpen(false);
      setSelectedDeuda(null);
      setModalFeedback({
        open: true,
        title: "¡Liquidación Exitosa!",
        description: `${res.mensaje}. Se generó la liquidación N° ${res.liquidacion.id} por un total de $${res.liquidacion.total.toFixed(2)}.`,
        isError: false,
      });
      await cargarDatos();
    } catch (err: unknown) {
      setConfirmOpen(false);
      setModalFeedback({
        open: true,
        title: "Error al Liquidar",
        description: err instanceof Error ? err.message : "Error al procesar la liquidación.",
        isError: true,
      });
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
                      <Button
                        size="sm"
                        onClick={() => handleOpenConfirm(d)}
                        disabled={liquidandoId === d.referidoId}
                        className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold text-xs"
                      >
                        {liquidandoId === d.referidoId ? (
                          <>
                            <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
                            Procesando...
                          </>
                        ) : (
                          "Liquidar Pago"
                        )}
                      </Button>
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

      {/* Modal Shadcn para Confirmar Liquidación */}
      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-indigo-100 text-indigo-600 mb-2">
              <DollarSign className="h-6 w-6" />
            </div>
            <DialogTitle className="text-center text-lg font-bold text-slate-900">
              Confirmar Liquidación de Comisiones
            </DialogTitle>
            <DialogDescription className="text-center text-slate-500 text-sm">
              Esta acción consolidará las comisiones pendientes y las marcará como pagadas.
            </DialogDescription>
          </DialogHeader>

          {selectedDeuda && (
            <div className="bg-slate-50 border border-slate-200 rounded-lg p-4 space-y-2.5 my-2">
              <div className="flex justify-between text-sm">
                <span className="text-slate-500">Distribuidor:</span>
                <span className="font-semibold text-slate-900">{selectedDeuda.nombre}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-500">ID Referido:</span>
                <span className="text-slate-700 font-mono text-xs">{selectedDeuda.referidoId}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-500">Comisiones acumuladas:</span>
                <span className="font-medium text-slate-700">{selectedDeuda.numeroComisiones} órdenes</span>
              </div>
              <div className="border-t border-slate-200 pt-2 flex justify-between items-center">
                <span className="font-semibold text-slate-900 text-sm">Monto total a liquidar:</span>
                <span className="font-bold text-green-600 text-lg">
                  ${selectedDeuda.totalPendiente.toFixed(2)}
                </span>
              </div>
            </div>
          )}

          <DialogFooter className="flex gap-2 sm:justify-end">
            <Button
              variant="outline"
              onClick={() => setConfirmOpen(false)}
              disabled={liquidandoId !== null}
            >
              Cancelar
            </Button>
            <Button
              className="bg-indigo-600 hover:bg-indigo-700 text-white"
              onClick={handleConfirmarLiquidacion}
              disabled={liquidandoId !== null}
            >
              {liquidandoId !== null ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Procesando pago...
                </>
              ) : (
                "Confirmar y Pagar"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Modal Shadcn para Resultado / Notificación */}
      <Dialog
        open={modalFeedback.open}
        onOpenChange={(open) => setModalFeedback((prev) => ({ ...prev, open }))}
      >
        <DialogContent className="sm:max-w-sm text-center">
          <DialogHeader>
            <div
              className={`mx-auto flex h-12 w-12 items-center justify-center rounded-full mb-2 ${
                modalFeedback.isError
                  ? "bg-red-100 text-red-600"
                  : "bg-emerald-100 text-emerald-600"
              }`}
            >
              {modalFeedback.isError ? (
                <AlertCircle className="h-6 w-6" />
              ) : (
                <CheckCircle2 className="h-6 w-6" />
              )}
            </div>
            <DialogTitle className="text-center text-lg font-bold text-slate-900">
              {modalFeedback.title}
            </DialogTitle>
            <DialogDescription className="text-center text-slate-600 text-sm mt-1">
              {modalFeedback.description}
            </DialogDescription>
          </DialogHeader>

          <DialogFooter className="sm:justify-center mt-2">
            <Button
              className="w-full sm:w-auto min-w-[120px]"
              variant={modalFeedback.isError ? "destructive" : "default"}
              onClick={() => setModalFeedback((prev) => ({ ...prev, open: false }))}
            >
              Entendido
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default AdminLiquidaciones;
