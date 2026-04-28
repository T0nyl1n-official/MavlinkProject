import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/users': {
        target: 'https://api.deeppluse.dpdns.org',
        changeOrigin: true,
        secure: false
      },
      '/api': {
        target: 'https://api.deeppluse.dpdns.org',
        changeOrigin: true,
        secure: false
      },
      '/terminal': {
        target: 'https://api.deeppluse.dpdns.org',
        changeOrigin: true,
        secure: false
      },
      '/mavlink': {
        target: 'https://api.deeppluse.dpdns.org',
        changeOrigin: true,
        secure: false
      },
      '/admin': {
        target: 'https://api.deeppluse.dpdns.org',
        changeOrigin: true,
        secure: false
      }
    }
  }
})