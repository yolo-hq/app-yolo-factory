import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

const port = parseInt(process.env.PORT || '3100', 10)
const apiPort = parseInt(process.env.YOLO_API_PORT || '9000', 10)

// Resolve @yolo-hq/view to source for HMR during dev
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
