import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

const apiPort = parseInt(process.env.YOLO_API_PORT || '9000', 10)
const viewSrc = resolve(__dirname, '../../../../view/src')

export default defineConfig({
  plugins: [tailwindcss(), react()],
  resolve: {
    alias: {
      '@yolo-hq/view/styles': viewSrc + '/styles/base.css',
      '@yolo-hq/view': viewSrc + '/index.ts',
    },
  },
  server: {
    port: 3002,
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
