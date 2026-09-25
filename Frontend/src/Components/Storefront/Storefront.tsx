import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../../Context/AuthContext";
import { getProductos } from "../../services/productosService";
import { categorias, type Producto } from "../../data/productos";

/**
 * Vista de tienda a pantalla completa (Storefront) con Hero inmersivo, categorías y destacados.
 */
const Storefront = () => {
  const { user } = useAuth();
  const [destacados, setDestacados] = useState<Producto[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    let activo = true;
    getProductos()
      .then((data) => {
        if (activo) {
          setDestacados(data.slice(0, 4));
          setLoading(false);
        }
      })
      .catch(() => {
        if (activo) setLoading(false);
      });

    return () => {
      activo = false;
    };
  }, []);

  return (
    <div>
      {/* Hero Full-Screen */}
      <section className="relative min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-700 via-purple-700 to-slate-900 text-white overflow-hidden rounded-2xl mb-12">
        <div className="absolute -top-24 -left-24 w-96 h-96 bg-indigo-500/30 rounded-full blur-3xl" />
        <div className="absolute -bottom-32 -right-24 w-[28rem] h-[28rem] bg-purple-500/30 rounded-full blur-3xl" />

        <div className="relative z-10 text-center px-6 max-w-4xl mx-auto py-16">
          <p className="text-sm md:text-base uppercase tracking-widest text-indigo-200 mb-4 font-semibold">
            Catálogo de venta multinivel
          </p>
          <h1 className="text-4xl md:text-6xl font-extrabold leading-tight mb-6">
            Belleza que se comparte,{" "}
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-amber-300 to-pink-300">
              ganancias que crecen
            </span>
          </h1>
          <p className="text-lg md:text-xl text-indigo-100 mb-10 max-w-2xl mx-auto">
            Explora nuestra línea de cosméticos premium y descubre cómo tu red de
            referidos genera comisiones en cada nivel.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              to="/catalogo"
              className="bg-white text-indigo-700 font-bold px-8 py-4 rounded-lg hover:bg-indigo-50 transition shadow-lg text-center"
            >
              Ver Catálogo Completo
            </Link>
            {user?.rol === "admin" && (
              <Link
                to="/mi-red"
                className="border-2 border-white/60 text-white font-bold px-8 py-4 rounded-lg hover:bg-white/10 transition text-center"
              >
                Conocer el Plan Multinivel
              </Link>
            )}
          </div>
        </div>
      </section>

      {/* Categorías */}
      <section className="py-12 px-2 max-w-6xl mx-auto">
        <h2 className="text-2xl md:text-3xl font-bold text-slate-800 text-center mb-10">
          Explora por Categoría
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          {categorias.map((cat) => (
            <Link
              key={cat}
              to={`/catalogo?categoria=${encodeURIComponent(cat)}`}
              className="bg-white border border-slate-200 rounded-xl p-6 text-center shadow-sm hover:shadow-md hover:border-indigo-300 hover:-translate-y-1 transition-all"
            >
              <p className="font-semibold text-slate-700">{cat}</p>
            </Link>
          ))}
        </div>
      </section>

      {/* Productos Destacados */}
      <section className="py-12 px-6 bg-slate-50 rounded-2xl max-w-6xl mx-auto mb-12">
        <h2 className="text-2xl md:text-3xl font-bold text-slate-800 text-center mb-10">
          Productos Destacados
        </h2>

        {loading ? (
          <p className="text-center text-slate-500 py-8">Cargando destacados...</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {destacados.map((prod) => (
              <Link
                key={prod.id}
                to={`/producto/${prod.id}`}
                className="bg-white rounded-xl overflow-hidden border border-slate-200 shadow-sm hover:shadow-lg hover:-translate-y-1 transition-all group flex flex-col"
              >
                <img
                  src={prod.img}
                  alt={prod.nombre}
                  className="w-full h-48 object-cover group-hover:scale-105 transition-transform"
                />
                <div className="p-4 flex-1 flex flex-col justify-between">
                  <div>
                    <p className="text-xs uppercase tracking-wide text-indigo-500 font-semibold">
                      {prod.categoria}
                    </p>
                    <h3 className="font-semibold text-slate-800 mt-1">
                      {prod.nombre}
                    </h3>
                  </div>
                  <p className="text-indigo-600 font-bold mt-4">
                    ${prod.precio.toFixed(2)}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>

      {/* Cómo Funciona */}
      <section className="py-12 px-2 max-w-6xl mx-auto mb-12">
        <h2 className="text-2xl md:text-3xl font-bold text-slate-800 text-center mb-12">
          ¿Cómo funciona el plan multinivel?
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {[
            {
              paso: "1",
              titulo: "Explora y compra",
              texto:
                "Elige tus productos favoritos del catálogo y genera tu primera venta.",
            },
            {
              paso: "2",
              titulo: "Invita a tu red",
              texto:
                "Comparte tu enlace de referido. Cada invitado suma un nivel en tu red.",
            },
            {
              paso: "3",
              titulo: "Gana comisiones",
              texto:
                "Recibe 10 %, 5 % y 2 % de las ventas de tus niveles 1, 2 y 3.",
            },
          ].map((item) => (
            <div
              key={item.paso}
              className="bg-white border border-slate-200 rounded-xl p-8 text-center shadow-sm"
            >
              <div className="w-12 h-12 mx-auto rounded-full bg-indigo-600 text-white font-bold text-xl flex items-center justify-center mb-4">
                {item.paso}
              </div>
              <h3 className="font-bold text-slate-800 mb-2">{item.titulo}</h3>
              <p className="text-slate-500 text-sm">{item.texto}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Llamada Final */}
      <section className="py-16 px-6 bg-slate-900 text-center rounded-2xl">
        <h2 className="text-2xl md:text-4xl font-bold text-white mb-4">
          Comienza tu propia red hoy
        </h2>
        <p className="text-slate-300 mb-8 max-w-xl mx-auto">
          Únete a MultiCatálogo y transforma tu red de contactos en un negocio
          de cosméticos premium.
        </p>
        <Link
          to="/catalogo"
          className="inline-block bg-indigo-600 text-white font-bold px-10 py-4 rounded-lg hover:bg-indigo-500 transition shadow-lg"
        >
          Explorar el Catálogo
        </Link>
      </section>
    </div>
  );
};

export default Storefront;
