import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const port = parseInt(process.env.PORT || '3000', 10)
const apiPort = parseInt(process.env.YOLO_API_PORT || '9000', 10)

export default defineConfig({
  plugins: [tailwindcss(), react()],
  server: {
    port,
    proxy: {
      '/api': {
        target: `http://localhost:${apiPort}`,
        changeOrigin: true,
      },
      '/_schema': {
        target: `http://localhost:${apiPort}`,
        changeOrigin: true,
      },
    },
  },
})
