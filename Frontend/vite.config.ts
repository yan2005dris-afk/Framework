import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

// Destino por defecto del proxy de dev cuando no hay VITE_DEV_API_TARGET.
// Es el caso de `pnpm dev` en el host: el gateway se publica en :8080.
const DEFAULT_DEV_API_TARGET = 'http://localhost:8080'

// https://vite.dev/config/
// Forma de funcion (en vez del objeto literal) para poder leer el entorno.
export default defineConfig(({ mode }) => {
  // `loadEnv` lee los .env del proyecto (.env, .env.local) y ademas mete lo que
  // venga en `process.env` con prefijo VITE_ (gana process.env sobre el
  // archivo). Ese detalle es lo que hace que el mismo archivo sirva para las
  // dos formas de correr el frontend:
  //   - en el host (`pnpm dev`): sin la variable, cae en localhost:8080.
  //   - en el servicio `frontend` del compose (profile `dev`): el contenedor
  //     inyecta VITE_DEV_API_TARGET=http://gateway:8080, porque dentro de la
  //     red de compose el gateway no se llama localhost.
  const env = loadEnv(mode, __dirname, 'VITE_')
  const devApiTarget = env.VITE_DEV_API_TARGET || DEFAULT_DEV_API_TARGET

  return {
    plugins: [
      // El plugin de TanStack Router genera routeTree.gen.ts a partir de
      // src/routes/. Tiene que ir ANTES de react().
      tanstackRouter({ target: 'react', autoCodeSplitting: true }),
      react(),
      tailwindcss(),
    ],
    resolve: {
      // Alias @/ -> src/. Lo exige shadcn para resolver sus imports.
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      // Escucha en 0.0.0.0 y no solo en loopback: sin esto el dev server corre
      // dentro del servicio `frontend` del compose pero el host no llega.
      // Es el equivalente en config de `vite --host`, y evita depender de que
      // cada quien pase el flag (con pnpm el `pnpm run dev -- --host` termina
      // ejecutando `vite -- --host`, que NO aplica el flag).
      host: true,
      // En dev el frontend habla a `/api` (mismo origen, ver VITE_API_URL) y este
      // proxy reenvia al gateway, que es el unico punto de entrada del stack.
      // Asi no hay CORS y la cookie httpOnly `auth_token` (host-only, SameSite=lax)
      // se emite y viaja en el mismo host que el SPA.
      // El target sale de VITE_DEV_API_TARGET; ver la nota de arriba.
      // Si se corre el Backend nativo en :3000 sin gateway:
      // `VITE_DEV_API_TARGET=http://localhost:3000 pnpm dev`.
      proxy: {
        '/api': {
          target: devApiTarget,
          changeOrigin: false,
        },
      },
    },
  }
})
