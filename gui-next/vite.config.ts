import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

// CursedChrome GUI build config.
//
// Output goes into ../gui/dist so the Go server can keep serving from
// the same path (GUI_DIST_PATH = /work/gui/dist) without needing any
// backend change. The /api/v1/* and /favicon.ico paths are proxied to
// the live API during dev so cookie-based session works against the
// running Go server on :8118.
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  build: {
    outDir: path.resolve(__dirname, '../gui/dist'),
    emptyOutDir: true,
    target: 'es2022',
    sourcemap: true,
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://192.168.13.202:8118',
        changeOrigin: true,
        secure: false,
      },
      '/health': 'http://192.168.13.202:8118',
      '/favicon.ico': 'http://192.168.13.202:8118',
    },
  },
})
