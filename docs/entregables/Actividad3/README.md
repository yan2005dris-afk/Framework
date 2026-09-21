# Conexión Frontend - Backend (API RESTful en Go)

**Asignatura:** Frameworks - 7mo Semestre  
**Actividad:** Actividad 3 - Integración Fullstack y Ejercicio de Extensión  
**Repositorio GitHub:** [https://github.com/yan2005dris-afk/Framework](https://github.com/yan2005dris-afk/Framework)

---

## 📌 Resumen de lo Implementado

Se conectó la aplicación cliente (**React 19 + TypeScript + Vite + Tailwind CSS**) con la API RESTful (**Go + Fiber**):

1. **Configuración del Servidor Backend (Go + Fiber):**
   - Puerto de escucha: `8080`.
   - Middleware de **CORS** (`AllowOrigins: "*"`, `AllowHeaders`, `AllowMethods`) para permitir peticiones cruzadas desde el frontend.
   - Middleware de **Logger** para visualizar el método, ruta y código de respuesta HTTP en la terminal.

2. **Endpoints RESTful:**
   - `POST /api/login`: Recibe credenciales (`email`, `password`), valida al usuario y devuelve el token de sesión.
   - `GET /api/productos`: Retorna el listado de productos en formato JSON.

3. **Cliente API en Frontend (`Frontend/src/services/api.ts`):**
   - Centraliza las llamadas HTTP con `fetch` y manejo de errores.

4. **Autenticación y Persistencia (`Login.tsx` y `AuthContext.tsx`):**
   - Envío de credenciales a la API de Go con estados de carga (`loading`).
   - Persistencia de sesión mediante `localStorage` para evitar que el usuario se desconecte al recargar la página.

5. **Catálogo Dinámico (`Catalogo.tsx`):**
   - Carga reactiva de productos mediante `useEffect`.
   - Skeletons de carga animados y botón de actualización interactivo.

6. **Ejercicio de Extensión:**
   - `Sidebar.tsx`: Soporta la prop `isCollapsed` (reducción a 80px y solo iconos).
   - `Navbar.tsx`: Botón `Toggle` para colapsar/expandir y abrir el menú lateral en móviles.
   - Adaptación responsiva: Drawer deslizable en celulares y vista de **Cards** en `MiRed.tsx`.

---

## 🚀 Instrucciones de Ejecución

### 1. Iniciar el Backend (Go)
```bash
cd Backend
go run main.go
```
*Servidor en:* `http://localhost:8080`

### 2. Iniciar el Frontend (React + Vite)
```bash
cd Frontend
pnpm install
pnpm dev
```
*Aplicación en:* `http://localhost:5173`

### 🔑 Credenciales de Prueba
- **Usuario:** `admin@upse.edu.ec`
- **Contraseña:** `123456`

---

## 📄 Documentos Entregables

- [Conexion_API_Backend_Reporte.pdf](./Conexion_API_Backend_Reporte.pdf)
- [Conexion_API_Backend_Reporte.docx](./Conexion_API_Backend_Reporte.docx)
- [Actividad_3_Reporte_Final.pdf](./Actividad_3_Reporte_Final.pdf)
- [Actividad_3_Reporte_Final.docx](./Actividad_3_Reporte_Final.docx)
