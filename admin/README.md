# GoForge Admin Panel

管理后台前端项目

## 技术栈

- Vue 3 + TypeScript
- Element Plus
- Vue Router
- Pinia
- Axios
- Vite

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器（确保后端 API Gateway 在 http://localhost:8080 运行）
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview
```

## 项目结构

```
goforge-admin/
├── src/
│   ├── api/          # API 接口
│   ├── assets/       # 静态资源
│   ├── components/   # 公共组件
│   ├── layouts/      # 布局组件
│   ├── router/       # 路由配置
│   ├── stores/       # Pinia 状态管理
│   ├── views/        # 页面组件
│   ├── utils/        # 工具函数
│   └── main.ts       # 入口文件
├── public/           # 公共静态文件
└── package.json
```

## API 配置

后端 API Gateway 地址：`http://localhost:8080`

开发环境已配置 Vite 代理，所有 `/api` 请求会自动转发到后端。

**重要提示**：
- 确保后端 API Gateway 在 `http://localhost:8080` 运行
- 如果修改了 Vite 配置，需要重启开发服务器（`npm run dev`）
- 如果遇到 404 错误，检查：
  1. 后端 API Gateway 是否正常运行
  2. Vite 开发服务器是否已重启
  3. 浏览器控制台的网络请求 URL 是否正确

## 常见问题

### 1. 404 错误
- 检查后端 API Gateway 是否在 `http://localhost:8080` 运行
- 重启 Vite 开发服务器：`Ctrl+C` 停止，然后 `npm run dev` 重新启动

### 2. CORS 错误
- 后端已配置 CORS，如果仍有问题，检查 Gateway 配置

### 3. 登录失败
- 确保已创建测试用户
- 检查用户名和密码是否正确
