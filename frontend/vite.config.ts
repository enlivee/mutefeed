import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: path.resolve(__dirname, '../internal/web/static'),
    emptyOutDir: true,
  },
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8899',
      '/feed': 'http://localhost:8899',
      '/badge': 'http://localhost:8899',
      '/healthz': 'http://localhost:8899',
    },
  },
})
