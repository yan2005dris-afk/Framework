import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
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
    // En dev el frontend habla a `/api` (mismo origen, ver VITE_API_URL) y este
    // proxy reenvia al gateway, que es el unico punto de entrada del stack.
    // Asi no hay CORS y la cookie httpOnly `auth_token` (host-only, SameSite=lax)
    // se emite y viaja en el mismo host que el SPA.
    // Si se corre el Backend nativo en :3000 sin gateway, cambiar el target.
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: false,
      },
    },
  },
})
