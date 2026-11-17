import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

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

      // U4-T3: convierte el SPA en una PWA instalable.
      VitePWA({
        // `autoUpdate`: el service worker toma el control inmediatamente
        // y precachea todos los recursos en segundo plano sin esperar a que
        // el usuario cierre todas las pestañas.
        registerType: 'autoUpdate',
        // Inyecta el registro automáticamente en el HTML
        injectRegister: 'auto',
        manifest: false,
        workbox: {
          // Los .woff2 van aparte porque @fontsource los carga en runtime y no
          // aparecen en el grafo de imports del build.
          globPatterns: ['**/*.{js,css,html,svg,png,woff,woff2}'],
          // Toda navegacion offline entrega el index.html del propio build, no
          // una copia vieja de un cache separado.
          navigateFallback: '/index.html',
          // Explicito aunque navigateFallback solo aplica a navegaciones: si
          // alguien enruta /api al fallback, un 404 de API se vuelve un 200 de
          // HTML y el error se deforma en el cliente.
          navigateFallbackDenylist: [/^\/api\//],
          // Reclamar inmediatamente a todos los clientes abiertos en cuanto se active
          clientsClaim: true,
          skipWaiting: true,
          runtimeCaching: [
            {
              // Detalle de un producto: cambia poco y se mira mucho. Se sirve
              // al instante desde el cache y se refresca por detras.
              urlPattern: /^\/api\/productos\/\d+$/,
              handler: 'StaleWhileRevalidate',
              options: {
                cacheName: 'api-producto-detalle',
                expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 * 24 },
              },
            },
            {
              // Listado del catalogo: es lo unico que es igual para todos los
              // usuarios, asi que cachearlo no filtra nada. NetworkFirst para
              // que un deploy con productos nuevos se vea al instante.
              urlPattern: /^\/api\/productos\/?$/,
              handler: 'NetworkFirst',
              options: {
                cacheName: 'api-catalogo',
                networkTimeoutSeconds: 4,
                expiration: { maxEntries: 5, maxAgeSeconds: 60 * 60 * 24 },
              },
            },
            {
              // TODO lo demas de /api va SOLO por red, sin regla de cache.
              //
              // El checklist de U4-T3 pedia NetworkFirst para todo /api/*, pero
              // /api/me y /api/red devuelven datos del usuario autenticado por
              // la cookie httpOnly: cacheados, pueden seguir mostrandose
              // despues de que el mismo dispositivo haya cambiado de sesion.
              // `/api/checkout` y `/api/login` van por la misma razon y por una
              // mas fuerte: nunca se reintenta sola una operacion transaccional
              // desde el service worker.
              urlPattern: /^\/api\//,
              handler: 'NetworkOnly',
            },
          ],
        },
        devOptions: {
          // El service worker en `pnpm dev` cachea el HMR y las rutas del
          // router y termina con devtools mostrando un SW que no existe. Se
          // verifica sobre `pnpm build && pnpm preview`, que es como corre la
          // app de verdad (dentro de la imagen del gateway).
          enabled: false,
        },
      }),
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
