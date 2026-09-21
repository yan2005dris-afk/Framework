import { Icon } from "./Share/Icon";

const MiRed = () => {
    const referidos = [
        { id: 1, nombre: "Ana García", nivel: "Nivel 1", ventas: "$1,200" },
        { id: 2, nombre: "Luis Poveda", nivel: "Nivel 1", ventas: "$850" },
        { id: 3, nombre: "Marta Sánchez", nivel: "Nivel 2", ventas: "$430" },
    ];

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-bold text-slate-800">
                    Mi Red de Referidos
                </h1>
                <p className="text-sm text-slate-500 mt-1">
                    Visualiza el desempeño y ventas de tus miembros afiliados.
                </p>
            </div>

            {/* Vista en formato Cards para móviles (< md) */}
            <div className="grid grid-cols-1 gap-4 md:hidden">
                {referidos.map((ref) => (
                    <div
                        key={ref.id}
                        className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm flex flex-col gap-3"
                    >
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <div className="w-10 h-10 rounded-full bg-indigo-50 text-indigo-600 flex items-center justify-center font-semibold text-sm">
                                    <Icon name="user" size={20} />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-slate-800 text-base">
                                        {ref.nombre}
                                    </h3>
                                    <span className="inline-block bg-slate-100 text-slate-600 text-xs px-2 py-0.5 rounded-full font-medium mt-0.5">
                                        {ref.nivel}
                                    </span>
                                </div>
                            </div>
                        </div>

                        <div className="flex items-center justify-between pt-2 border-t border-slate-100 text-sm">
                            <span className="text-slate-500 font-medium">Ventas Mensuales:</span>
                            <span className="text-indigo-600 font-bold text-base">
                                {ref.ventas}
                            </span>
                        </div>
                    </div>
                ))}
            </div>

            {/* Vista en formato Tabla para pantallas medianas y grandes (>= md) */}
            <div className="hidden md:block bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                <table className="w-full text-left">
                    <thead className="bg-slate-50 border-b border-slate-200">
                        <tr>
                            <th className="p-4 font-semibold text-slate-600">Nombre</th>
                            <th className="p-4 font-semibold text-slate-600">Jerarquía</th>
                            <th className="p-4 font-semibold text-slate-600">
                                Ventas Mensuales
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {referidos.map((ref) => (
                            <tr
                                key={ref.id}
                                className="border-b border-slate-100 hover:bg-slate-50 transition-colors"
                            >
                                <td className="p-4 text-slate-800 font-medium">{ref.nombre}</td>
                                <td className="p-4 text-slate-500">
                                    <span className="bg-slate-100 text-slate-600 text-xs px-2.5 py-1 rounded-full font-medium">
                                        {ref.nivel}
                                    </span>
                                </td>
                                <td className="p-4 text-indigo-600 font-bold">
                                    {ref.ventas}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default MiRed;