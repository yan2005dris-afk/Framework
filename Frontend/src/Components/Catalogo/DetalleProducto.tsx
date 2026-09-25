import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { getProductoById } from "../../services/productosService";
import type { Producto } from "../../data/productos";
import { useCart } from "../../Context/CartContext";

/**
 * Vista de detalle de producto con selector de galería y visualizador Lightbox interactivo.
 */
const DetalleProducto = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addToCart } = useCart();

  const [producto, setProducto] = useState<Producto | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [imagenActiva, setImagenActiva] = useState<number>(0);
  const [lightboxAbierto, setLightboxAbierto] = useState<boolean>(false);
  const [agregado, setAgregado] = useState<boolean>(false);

  useEffect(() => {
    let activo = true;
    getProductoById(Number(id))
      .then((data) => {
        if (!activo) return;
        setProducto(data ?? null);
        setLoading(false);
      })
      .catch(() => {
        if (activo) setLoading(false);
      });

    return () => {
      activo = false;
    };
  }, [id]);

  useEffect(() => {
    if (!lightboxAbierto) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setLightboxAbierto(false);
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [lightboxAbierto]);

  if (loading) {
    return <div className="text-center p-8 text-slate-500">Cargando producto...</div>;
  }

  if (!producto) {
    return (
      <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
        <p className="text-slate-600 text-lg mb-4">Producto no encontrado.</p>
        <Link
          to="/catalogo"
          className="text-indigo-600 font-semibold hover:underline"
        >
          ← Volver al catálogo
        </Link>
      </div>
    );
  }

  const imagenes = [producto.img, ...(producto.galeria || [])];

  const handleAddToCart = () => {
    addToCart(producto);
    setAgregado(true);
    setTimeout(() => setAgregado(false), 1500);
  };

  const handlePrev = () =>
    setImagenActiva((prev) => (prev === 0 ? imagenes.length - 1 : prev - 1));
  const handleNext = () =>
    setImagenActiva((prev) => (prev === imagenes.length - 1 ? 0 : prev + 1));

  return (
    <div>
      <button
        onClick={() => navigate(-1)}
        className="text-indigo-600 font-semibold hover:underline mb-6 inline-flex items-center gap-1"
      >
        ← Volver
      </button>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Galería de imágenes */}
        <div>
          <img
            src={imagenes[imagenActiva]}
            alt={producto.nombre}
            onClick={() => setLightboxAbierto(true)}
            className="w-full h-80 md:h-96 object-cover rounded-xl border border-slate-200 cursor-zoom-in shadow-sm"
          />
          <div className="flex gap-3 mt-4 overflow-x-auto pb-2">
            {imagenes.map((img, idx) => (
              <button
                key={img + idx}
                onClick={() => setImagenActiva(idx)}
                className={`rounded-lg overflow-hidden border-2 transition flex-shrink-0 ${
                  idx === imagenActiva
                    ? "border-indigo-600"
                    : "border-transparent opacity-70 hover:opacity-100"
                }`}
              >
                <img
                  src={img}
                  alt={`Vista ${idx + 1}`}
                  className="w-20 h-16 object-cover"
                />
              </button>
            ))}
          </div>
        </div>

        {/* Información del Producto */}
        <div>
          <p className="text-xs uppercase tracking-wide text-indigo-500 font-semibold">
            {producto.categoria}
          </p>
          <h1 className="text-3xl font-bold text-slate-800 mt-1 mb-3">
            {producto.nombre}
          </h1>
          <p className="text-3xl font-bold text-indigo-600 mb-6">
            ${producto.precio.toFixed(2)}
          </p>
          <p className="text-slate-600 leading-relaxed mb-8">
            {producto.descripcion}
          </p>

          <button
            onClick={handleAddToCart}
            className={`w-full md:w-auto px-10 py-4 rounded-lg font-bold transition shadow-md ${
              agregado
                ? "bg-green-600 text-white"
                : "bg-indigo-600 text-white hover:bg-indigo-700"
            }`}
          >
            {agregado ? "✓ Añadido al carrito" : "Añadir al Carrito"}
          </button>
        </div>
      </div>

      {/* Lightbox Modal */}
      {lightboxAbierto && (
        <div
          className="fixed inset-0 z-50 bg-slate-950/95 flex items-center justify-center p-4"
          onClick={() => setLightboxAbierto(false)}
        >
          <button
            className="absolute top-5 right-5 text-white text-4xl hover:text-slate-300 transition"
            onClick={() => setLightboxAbierto(false)}
            aria-label="Cerrar"
          >
            ×
          </button>

          <button
            onClick={(e) => {
              e.stopPropagation();
              handlePrev();
            }}
            className="absolute left-4 md:left-10 text-white text-4xl hover:text-indigo-400 transition"
            aria-label="Anterior"
          >
            ‹
          </button>

          <img
            src={imagenes[imagenActiva]}
            alt={producto.nombre}
            onClick={(e) => e.stopPropagation()}
            className="max-h-[85vh] max-w-[85vw] object-contain rounded-lg"
          />

          <button
            onClick={(e) => {
              e.stopPropagation();
              handleNext();
            }}
            className="absolute right-4 md:right-10 text-white text-4xl hover:text-indigo-400 transition"
            aria-label="Siguiente"
          >
            ›
          </button>

          <p className="absolute bottom-6 text-white/80 text-sm">
            {imagenActiva + 1} / {imagenes.length}
          </p>
        </div>
      )}
    </div>
  );
};

export default DetalleProducto;
