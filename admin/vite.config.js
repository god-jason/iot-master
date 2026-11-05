import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      // 安装模式接口只存在于安装模式后端（8080），单独代理，
      // 其余接口仍走正常模式后端
      '/api/setup': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
        upgrade: true
      },
      '/api': {
        target: 'http://localhost:8888',
        changeOrigin: true,
        ws: true,
        upgrade: true
      },
      '/images': {
        target: 'http://localhost:8888',
        changeOrigin: true
      },
      '/attachments': {
        target: 'http://localhost:8888',
        changeOrigin: true
      }
    }
  }
})
