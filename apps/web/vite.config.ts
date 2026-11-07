import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => ({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api/v1': { target: loadEnv(mode, process.cwd(), 'VITE_').VITE_API_PROXY ?? 'http://127.0.0.1:8080', changeOrigin: false },
    },
  },
}))
