import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        // 把不常变的依赖拆成独立 chunk：改业务代码时用户只需重新下载 app chunk，
        // vendor（尤其是 element-plus）可以从浏览器缓存里直接命中。
        manualChunks: {
          vue: ['vue', 'vue-router', 'pinia'],
          'element-plus': ['element-plus'],
        },
      },
    },
  },
  server: {
    port: 3001, // 用户端使用 3001 端口，避免与管理后台冲突
    host: '0.0.0.0',
    proxy: {
      '^/api/.*': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        ws: true,
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            console.log('❌ Proxy error:', err);
          });
          proxy.on('proxyReq', (proxyReq, req, _res) => {
            console.log('➡️  Proxying:', req.method, req.url, '->', proxyReq.path);
          });
          proxy.on('proxyRes', (proxyRes, req, _res) => {
            console.log('⬅️  Response:', proxyRes.statusCode, req.url);
          });
        },
      },
      // 静态资源（上传的图片）与 /api 同源：前端代码里不再写死 http://localhost:8080，
      // 因此开发环境需要把这两个前缀也代理到网关。
      '^/(uploads|images)/.*': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})

