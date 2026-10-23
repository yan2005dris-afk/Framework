import { useEffect, useState } from "react";
import { getRed } from "../services/redService";
import { api, type ResumenComisionesDTO } from "../services/api";
import {
  contarRed,
  sumarVentasRed,
  sumarComisiones,
  nivelAlcanzado,
  comisionDeReferido,
  TASA_COMISION,
  type Referido,
} from "../data/red";

const NIVEL_ESTILO: Record<
  number,
  { badge: string; border: string; tasa: string }
> = {
  1: {
    badge: "bg-indigo-100 text-indigo-700",
    border: "border-indigo-200",
    tasa: "10 %",
  },
  2: {
    badge: "bg-purple-100 text-purple-700",
    border: "border-purple-200",
    tasa: "5 %",
  },
  3: {
    badge: "bg-pink-100 text-pink-700",
    border: "border-pink-200",
    tasa: "2 %",
  },
};

const NodoReferido = ({ nodo }: { nodo: Referido }) => {
  const estilo = NIVEL_ESTILO[nodo.nivel] ?? NIVEL_ESTILO[1];
  const comision = comisionDeReferido(nodo);

  return (
    <li>
      <div
        className={`bg-white rounded-lg border ${estilo.border} shadow-sm p-4 flex flex-wrap items-center justify-between gap-3`}
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center font-bold text-slate-600">
            {nodo.nombre.charAt(0)}
          </div>
          <div>
            <p className="font-semibold text-slate-800">{nodo.nombre}</p>
            <span
              className={`inline-block mt-1 text-xs font-semibold px-2 py-0.5 rounded-full ${estilo.badge}`}
            >
              Nivel {nodo.nivel} · {estilo.tasa} comisión
            </span>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm text-slate-500">Ventas</p>
          <p className="font-bold text-slate-800">
            ${nodo.ventas.toLocaleString()}
          </p>
          <p className="text-xs text-indigo-600 font-semibold mt-1">
            +${comision.toFixed(2)} para ti
          </p>
        </div>
      </div>

      {nodo.hijos && nodo.hijos.length > 0 && (
        <ul className="mt-3 ml-6 space-y-3 border-l-2 border-slate-200 pl-4">
          {nodo.hijos.map((hijo) => (
            <NodoReferido key={hijo.id} nodo={hijo} />
          ))}
        </ul>
      )}
    </li>
  );
};

const MiRed = () => {
  const [red, setRed] = useState<Referido | null>(null);
  const [comisionesResumen, setComisionesResumen] = useState<ResumenComisionesDTO | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [copiado, setCopiado] = useState<boolean>(false);

  const cargarDatos = () => {
    setLoading(true);
    setError(null);
    Promise.all([getRed(), api.getComisionesMias().catch(() => null)])
      .then(([dataRed, dataComisiones]) => {
        setRed(dataRed);
        setComisionesResumen(dataComisiones);
        setLoading(false);
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Error al cargar la red multinivel");
        setLoading(false);
      });
  };

  useEffect(() => {
    cargarDatos();
  }, []);

  const copiarEnlace = () => {
    const enlace = `${window.location.origin}/registro?ref=TU`;
    navigator.clipboard.writeText(enlace);
    setCopiado(true);
    setTimeout(() => setCopiado(false), 2000);
  };

  if (loading) {
    return <div className="text-center p-8 text-slate-500">Cargando red multinivel...</div>;
  }

  if (error || !red) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 p-6 rounded-xl text-center space-y-3">
        <p className="font-semibold">{error || "No se pudo cargar la red"}</p>
        <button
          onClick={cargarDatos}
          className="px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 transition"
        >
          Reintentar
        </button>
      </div>
    );
  }

  const referidosActivos = contarRed(red);
  const ventasRed = sumarVentasRed(red);
  const comisionesMes = sumarComisiones(red);
  const nivel = nivelAlcanzado(red.hijos?.length ?? 0);

  const resumen = [
    {
      etiqueta: "Referidos Activos",
      valor: referidosActivos.toString(),
      color: "text-indigo-600",
    },
    {
      etiqueta: "Ventas de la Red",
      valor: `$${ventasRed.toLocaleString()}`,
      color: "text-indigo-600",
    },
    {
      etiqueta: "Comisiones del Mes",
      valor: `$${comisionesMes.toFixed(2)}`,
      color: "text-green-600",
    },
    {
      etiqueta: "Nivel Alcanzado",
      valor: nivel,
      color: "text-amber-600",
    },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-800 mb-2">
          Mi Red de Referidos
        </h1>
        <p className="text-slate-500 text-sm">
          Estructura multinivel: gana el {TASA_COMISION[1] * 100} % de nivel 1,{" "}
          {TASA_COMISION[2] * 100} % de nivel 2 y {TASA_COMISION[3] * 100} % de nivel 3.
        </p>
      </div>

      {/* Banner de Invitación de Referido */}
      <div className="bg-gradient-to-r from-indigo-600 to-indigo-800 text-white rounded-xl p-6 shadow-md flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h3 className="text-lg font-bold">Tu Enlace de Patrocinador</h3>
          <p className="text-indigo-100 text-xs mt-1">
            Comparte este enlace para que nuevos distribuidores o clientes se registren en tu red
          </p>
        </div>

        <div className="flex items-center gap-2 bg-white/10 p-1.5 rounded-lg border border-white/20">
          <input
            type="text"
            readOnly
            value={`${typeof window !== "undefined" ? window.location.origin : ""}/registro?ref=TU`}
            className="bg-transparent text-xs text-white px-2 py-1 outline-none w-48 sm:w-64 select-all font-mono"
          />
          <button
            onClick={copiarEnlace}
            className="bg-white text-indigo-700 hover:bg-indigo-50 font-bold text-xs px-3.5 py-1.5 rounded transition shadow-sm"
          >
            {copiado ? "¡Copiado!" : "Copiar Enlace"}
          </button>
        </div>
      </div>

      {/* Tarjetas de resumen */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {resumen.map((item) => (
          <div
            key={item.etiqueta}
            className="bg-white p-6 rounded-xl shadow-sm border border-slate-200"
          >
            <p className="text-xs text-slate-500 uppercase font-semibold">
              {item.etiqueta}
            </p>
            <p className={`text-2xl font-bold mt-2 ${item.color}`}>
              {item.valor}
            </p>
          </div>
        ))}
      </div>

      {/* Ledger de Comisiones Devengadas */}
      {comisionesResumen && (
        <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
          <div className="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <h2 className="font-bold text-slate-900 text-base">Ledger de Comisiones</h2>
              <p className="text-xs text-slate-500 mt-0.5">Auditoría detallada de cada comisión generada por tus referidos</p>
            </div>

            <div className="flex items-center gap-3">
              <span className="text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200 px-2.5 py-1 rounded-lg">
                Pendiente: ${comisionesResumen.totalPendiente.toFixed(2)}
              </span>
              <span className="text-xs font-semibold bg-green-50 text-green-700 border border-green-200 px-2.5 py-1 rounded-lg">
                Pagado: ${comisionesResumen.totalPagado.toFixed(2)}
              </span>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-sm">
              <thead>
                <tr className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold text-xs uppercase tracking-wider">
                  <th className="py-3 px-4">Pedido ID</th>
                  <th className="py-3 px-4">Nivel MLM</th>
                  <th className="py-3 px-4">Tasa</th>
                  <th className="py-3 px-4">Monto</th>
                  <th className="py-3 px-4">Fecha</th>
                  <th className="py-3 px-4 text-right">Estado</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-slate-800">
                {comisionesResumen.comisiones.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="py-6 text-center text-slate-400 text-xs">
                      No hay registros en el ledger todavía.
                    </td>
                  </tr>
                ) : (
                  comisionesResumen.comisiones.map((c) => (
                    <tr key={c.id} className="hover:bg-slate-50/75 transition text-xs">
                      <td className="py-3 px-4 font-mono font-semibold text-slate-900">
                        #{c.pedidoId}
                      </td>
                      <td className="py-3 px-4">Nivel {c.nivel}</td>
                      <td className="py-3 px-4">{c.porcentaje}%</td>
                      <td className="py-3 px-4 font-bold text-green-600">${c.monto.toFixed(2)}</td>
                      <td className="py-3 px-4 text-slate-500">
                        {new Date(c.creadoEn).toLocaleDateString("es-ES")}
                      </td>
                      <td className="py-3 px-4 text-right">
                        <span
                          className={`px-2 py-0.5 rounded-full font-semibold capitalize ${
                            c.estado === "pagada"
                              ? "bg-green-100 text-green-700"
                              : "bg-amber-100 text-amber-700"
                          }`}
                        >
                          {c.estado}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Árbol de referidos */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-4">
          Jerarquía de tu Red
        </h2>

        {/* Raíz: Usuario autenticado */}
        <div className="bg-slate-900 text-white rounded-lg p-4 mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-indigo-500 flex items-center justify-center font-bold">
              {red.nombre.charAt(0)}
            </div>
            <div>
              <p className="font-semibold">{red.nombre}</p>
              <p className="text-xs text-slate-300">Nivel alcanzado: {nivel}</p>
            </div>
          </div>
          <p className="font-bold text-green-400">
            Comisiones del mes: ${comisionesMes.toFixed(2)}
          </p>
        </div>

        <ul className="space-y-3">
          {red.hijos?.map((hijo) => (
            <NodoReferido key={hijo.id} nodo={hijo} />
          ))}
        </ul>
      </div>
    </div>
  );
};

export default MiRed;
