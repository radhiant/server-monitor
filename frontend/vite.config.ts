import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// Where each node's agent listens in development. Override any of them with
// SRV_<ID>=host:port, e.g. SRV_WEB01=192.0.2.50:9191 npm run dev
const AGENTS: Record<string, string> = {
  web01: '192.0.2.11:9191',
  app01: '192.0.2.12:9191',
  db01: '192.0.2.13:9191',
  monitor01: '192.0.2.14:9191',
  proxy01: '192.0.2.15:9191',
  edge01: '198.51.100.20:9191',
}

const proxy = Object.fromEntries(
  Object.entries(AGENTS).flatMap(([id, addr]) => {
    const target = process.env[`SRV_${id.toUpperCase()}`] || addr
    return [
      [`/servers/${id}/api`, { target: `http://${target}`, changeOrigin: true, rewrite: (p: string) => p.replace(`/servers/${id}/api`, '/api') }],
      [`/servers/${id}/ws`, { target: `ws://${target}`, ws: true, rewrite: (p: string) => p.replace(`/servers/${id}/ws`, '/ws') }],
    ]
  }),
)

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  build: {
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: {
          'echarts': ['echarts'],
          'vue-vendor': ['vue', 'pinia', 'lucide-vue-next']
        }
      }
    }
  },
  server: {
    port: 3000,
    host: true,
    proxy
  }
})
