// src/App.tsx

import { BrowserRouter, Navigate, Outlet, Route, Routes } from "react-router-dom";
import { AuthProvider, useAuth } from "./Context/AuthContext";
import { CartProvider } from "./Context/CartContext";
import Login from "./Components/Auth/Login/Login";
import Layout from "./Components/Layout/Layout";
import Dashboard from "./Components/Dashboard/Dashboard";
import Catalogo from "./Components/Catalogo/Catalogo";
import MiRed from "./Components/MiRed";
import Carrito from "./Components/Carrito/Carrito";

// Componente para proteger las rutas privadas
const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth();
  // Si no está autenticado, lo enviamos al login
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  // Si está autenticado, renderiza las rutas hijas (Outlet)
  return <Outlet />;
};
function App() {
  return (
    <AuthProvider> {/* Proveedor de Autenticación */}
      <CartProvider> {/* Proveedor del Carrito */}
        <BrowserRouter>
          <Routes>
            {/* Ruta pública */}
            <Route path="/login" element={<Login />} />
            {/* Rutas protegidas */}
            <Route element={<ProtectedRoute />}>
              <Route path="/" element={<Layout />}>
                <Route index element={<Dashboard />} />
                <Route path="catalogo" element={<Catalogo />} />
                <Route path="mi-red" element={<MiRed />} />
                <Route path="carrito" element={<Carrito />} />
              </Route>
            </Route>
            {/* Ruta comodín para capturar 404 y redirigir */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </CartProvider>
    </AuthProvider>
  );
}
export default App;