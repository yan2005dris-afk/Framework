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
})
