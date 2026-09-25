import { type ReactNode } from "react";
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
  Outlet,
} from "react-router-dom";
import Layout from "./Components/Layout/Layout";
import Dashboard from "./Components/Dashboard/Dashboard";
import Catalogo from "./Components/Catalogo/Catalogo";
import Storefront from "./Components/Storefront/Storefront";
import DetalleProducto from "./Components/Catalogo/DetalleProducto";
import MiRed from "./Components/MiRed";
import Carrito from "./Components/Carrito/Carrito";
import Checkout from "./Components/Carrito/Checkout";
import Confirmacion from "./Components/Carrito/Confirmacion";
import Login from "./Components/Auth/Login/Login";
import { CartProvider } from "./Context/CartContext";
import { AuthProvider, useAuth } from "./Context/AuthContext";

/**
 * Guarda de ruta que requiere que el usuario esté autenticado.
 */
const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
};

/**
 * Guarda de ruta exclusiva para administradores.
 * Redirige a los clientes hacia la vista de tienda.
 */
const AdminRoute = () => {
  const { isAuthenticated, user } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (user?.rol !== "admin") {
    return <Navigate to="/tienda" replace />;
  }

  return <Outlet />;
};

/**
 * CartBoundary remonta el CartProvider según el email del usuario para garantizar
 * el aislamiento e integridad de los datos del carrito por cuenta.
 */
const CartBoundary = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  return (
    <CartProvider key={user?.email ?? "anonimo"}>
      {children}
    </CartProvider>
  );
};

/**
 * Componente principal de la aplicación con la configuración del enrutador y los proveedores globales.
 */
function App() {
  return (
    <AuthProvider>
      <CartBoundary>
        <BrowserRouter>
          <Routes>
            {/* Ruta pública */}
            <Route path="/login" element={<Login />} />

            {/* Rutas protegidas */}
            <Route element={<ProtectedRoute />}>
              <Route path="/" element={<Layout />}>
                {/* Solo administrador */}
                <Route element={<AdminRoute />}>
                  <Route index element={<Dashboard />} />
                  <Route path="mi-red" element={<MiRed />} />
                </Route>

                {/* Ambos roles (admin y cliente) */}
                <Route path="tienda" element={<Storefront />} />
                <Route path="catalogo" element={<Catalogo />} />
                <Route path="producto/:id" element={<DetalleProducto />} />
                <Route path="carrito" element={<Carrito />} />
                <Route path="checkout" element={<Checkout />} />
                <Route path="confirmacion" element={<Confirmacion />} />
              </Route>
            </Route>

            {/* Redirección por defecto */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </CartBoundary>
    </AuthProvider>
  );
}

export default App;