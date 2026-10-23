import { useState, type FormEvent } from "react";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { api, type RegisterRequest } from "../../../services/api";

/**
 * Vista de registro público de clientes con soporte para código de patrocinador/referido.
 */
const Register = () => {
  const search = useSearch({ from: "/registro" }) as { ref?: string };
  const navigate = useNavigate();

  const [nombre, setNombre] = useState<string>("");
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [refCode, setRefCode] = useState<string>(search?.ref || "");
  const [error, setError] = useState<string>("");
  const [exito, setExito] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    setExito("");
    setLoading(true);

    try {
      const payload: RegisterRequest = {
        nombre,
        email,
        password,
        ref: refCode.trim() || undefined,
      };

      await api.register(payload);
      setExito("¡Cuenta creada con éxito! Redirigiendo al login...");
      setTimeout(() => {
        navigate({ to: "/login" });
      }, 1500);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Ocurrió un error al procesar el registro");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 px-4 py-8">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md p-8 border border-slate-200">
        <div className="text-center mb-6">
          <h2 className="text-3xl font-bold text-slate-900">Crear Cuenta</h2>
          <p className="text-slate-500 mt-2 text-sm">
            Únete a YMRT y forma parte de nuestra red de belleza
          </p>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 text-center border border-red-200">
            {error}
          </div>
        )}

        {exito && (
          <div className="bg-green-50 text-green-700 p-3 rounded-lg text-sm mb-6 text-center border border-green-200">
            {exito}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Nombre Completo
            </label>
            <input
              type="text"
              value={nombre}
              onChange={(e) => setNombre(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition text-sm"
              placeholder="Ej. Juan Pérez"
              disabled={loading}
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Correo Electrónico
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition text-sm"
              placeholder="juan@ejemplo.com"
              disabled={loading}
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Contraseña
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition text-sm"
              placeholder="Al menos 6 caracteres"
              disabled={loading}
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Código de Invitación / Referido (Opcional)
            </label>
            <input
              type="text"
              value={refCode}
              onChange={(e) => setRefCode(e.target.value.toUpperCase())}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition text-sm uppercase font-mono"
              placeholder="Ej. LUIS"
              disabled={loading}
            />
            {refCode && (
              <p className="text-xs text-indigo-600 mt-1">
                Te unirás a la red del patrocinador: <span className="font-semibold">{refCode}</span>
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full min-h-[44px] bg-indigo-600 text-white font-bold py-3 rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition flex items-center justify-center gap-2 mt-2"
          >
            {loading ? "Creando cuenta..." : "Registrarme"}
          </button>
        </form>

        <div className="mt-6 text-center text-sm text-slate-600 border-t border-slate-200 pt-4">
          ¿Ya tienes una cuenta?{" "}
          <Link to="/login" className="text-indigo-600 font-semibold hover:underline">
            Inicia sesión aquí
          </Link>
        </div>
      </div>
    </div>
  );
};

export default Register;
