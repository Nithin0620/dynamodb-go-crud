import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      '/auth': 'http://backend:8080',
      '/products': 'http://backend:8080',
      '/orders': 'http://backend:8080',
      '/admin': 'http://backend:8080',
      '/health': 'http://backend:8080',
    }
  }
})
