import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// 默认输出到 frontend/dist，适合 Vercel 托管前端。
// 本地如需继续把静态资源嵌入 Go 后端，可执行：BUILD_TARGET=backend npm run build
const outDir = process.env.BUILD_TARGET === 'backend'
  ? '../backend/frontend/dist'
  : 'dist'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
  },
  build: {
    outDir,
    emptyOutDir: true,
    assetsDir: 'assets',
  },
})
