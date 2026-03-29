import { defineConfig } from '@tanstack/start/config'

export default defineConfig({
  server: {
    preset: 'node-server',
  },
  vite: {
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:9000',
          changeOrigin: true,
        },
        '/_schema': {
          target: 'http://localhost:9000',
          changeOrigin: true,
        },
      },
    },
  },
})
