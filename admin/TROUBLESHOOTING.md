# 故障排查指南

## 404 错误：请求的资源不存在

如果遇到 `POST http://localhost:3000/api/v1/user/login 404 (Not Found)` 错误：

### 1. 检查 Vite 开发服务器是否正在运行

```bash
cd admin
npm run dev
```

确保看到类似输出：
```
  VITE v5.x.x  ready in xxx ms

  ➜  Local:   http://localhost:3000/
  ➜  Network: use --host to expose
```

### 2. 检查后端 API Gateway 是否运行

```bash
# 检查 Gateway 是否在 8080 端口运行
curl http://localhost:8080/api/v1/user/login
```

应该返回错误（不是 404），说明 Gateway 正在运行。

### 3. 重启 Vite 开发服务器

**重要**：修改 `vite.config.ts` 后必须重启开发服务器！

```bash
# 1. 停止当前服务器（Ctrl+C）
# 2. 重新启动
npm run dev
```

### 4. 检查浏览器控制台

打开浏览器开发者工具（F12），查看：
- **Network 标签**：检查请求 URL 是否正确
- **Console 标签**：查看是否有代理日志（如果配置了调试日志）

### 5. 验证代理配置

在 Vite 控制台应该看到代理日志：
```
➡️  Proxying: POST /api/v1/user/login -> /api/v1/user/login
⬅️  Response: 200 /api/v1/user/login
```

如果没有看到这些日志，说明代理没有生效。

### 6. 手动测试代理

在浏览器控制台运行：
```javascript
fetch('/api/v1/user/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'test', password: '123456', login_type: 1 })
}).then(r => r.json()).then(console.log).catch(console.error)
```

### 7. 常见问题

#### 问题：代理不工作
- **解决**：确保 `vite.config.ts` 中的 `proxy` 配置正确
- **解决**：重启 Vite 开发服务器

#### 问题：后端返回 404
- **解决**：检查后端 API Gateway 是否在 `http://localhost:8080` 运行
- **解决**：检查 Gateway 配置中的路由映射是否正确

#### 问题：CORS 错误
- **解决**：后端已配置 CORS，如果仍有问题，检查 Gateway 日志

### 8. 临时解决方案（如果代理不工作）

如果代理仍然不工作，可以临时修改 `src/utils/request.ts`：

```typescript
const service: AxiosInstance = axios.create({
  baseURL: 'http://localhost:8080/api', // 直接使用后端地址
  timeout: 30000,
  // ...
})
```

**注意**：这会导致 CORS 问题，需要确保后端 CORS 配置正确。

