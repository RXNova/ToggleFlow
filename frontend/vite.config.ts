import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      // @ maps to src/ — like Angular's path aliases in tsconfig
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // Forward /api and /sdk requests to the Go backend in dev
      // Angular equivalent: proxy.conf.json
      '/api': 'http://localhost:8080',
      '/sdk': 'http://localhost:8080',
    },
  },
})
