import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'
// https://vite.dev/config/
export default defineConfig({
  base: '/lion/',
  plugins: [
    vue(),
    tailwindcss(),
    vueJsx(),
  ],
  resolve: {
    extensions: ['.js', '.ts', '.tsx', '.vue'],
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    port: 9529,
    proxy: {
      '^/lion/ws': {
        target: 'http://localhost:8081',
        ws: true,
        changeOrigin: true,
      },
      '^/lion/api': {
        target: 'http://localhost:8081',
        ws: true,
        changeOrigin: true,
      },
      '^/lion/token': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        ws: true,
      },
    }
  },
})
