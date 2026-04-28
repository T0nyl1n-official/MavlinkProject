import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { startTunnel, stopTunnel } from './src/utils/startTunnel'

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  if (command === 'serve' && env.VITE_AUTO_START_TUNNEL === 'true') {
    startTunnel()
    process.once('exit', stopTunnel)
    process.once('SIGINT', () => {
      stopTunnel()
      process.exit(0)
    })
    process.once('SIGTERM', () => {
      stopTunnel()
      process.exit(0)
    })
  }

  return {
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
  }
})