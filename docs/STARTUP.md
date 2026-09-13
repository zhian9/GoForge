# 项目启动指南

> 两种启动方式：① Docker 一键部署（推荐，最简单）；② 本地开发（改代码调试用）。

---

## 一、环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Docker Desktop | 28+ | 部署 / 起基础设施必需 |
| Go | 1.21+ | 仅本地开发编译需要 |
| Node.js | 18+ | 仅本地开发前端需要 |

---

## 二、方式一：Docker 一键部署（推荐）

### 首次启动

```bash
cd deploy/compose
cp ../.env.example ../.env     # 复制环境变量（可按需改密码）
docker compose up -d           # 启动全部容器（首次会拉镜像/构建，约几分钟）
```

启动过程按「基础设施 → 微服务 → 网关」自动编排（靠 `depends_on` + `healthcheck`），MySQL 就绪前微服务不会启动。

### 访问地址

| 入口 | 地址 | 账号 |
|------|------|------|
| 管理后台 | http://localhost/ | admin / admin123 |
| 用户端 | http://localhost:8088/ | 注册即用 |
| API 网关 | http://localhost:8080/api/v1/... | — |
| API 文档 (Swagger) | http://localhost:8095/ | — |

### 停止

```bash
docker compose down          # 停止（保留数据卷）
docker compose down -v       # 停止并清空所有数据
```

---

## 三、方式二：本地开发

适合改代码、断点调试。

### 1. 起基础设施（Docker）

```bash
make start-infra   # 起 mysql / redis / etcd / mongodb / es / zookeeper / kafka
```

> ⚠️ 端口注意：Docker 的 MySQL 映射到宿主机 **3307**，但 `configs/dev/*.yaml` 里数据库端口默认是 **3306**。若本机没有自装 MySQL，需把各 dev 配置的 `Database.Port` 改成 `3307`（或本地自装 MySQL 在 3306）。

### 2. 编译 + 启动微服务

```bash
make build           # 编译所有服务到 server/bin/
make start-services  # 后台启动所有微服务 + API 网关
```

或一条命令完成（编译 + 基础设施 + 启动）：

```bash
make start-all
```

单个服务调试：

```bash
make run SVC=user    # 编译并前台运行某个服务
```

### 3. 启动前端

```bash
cd admin && npm install && npm run dev   # 管理后台 http://localhost:3000
cd web && npm install && npm run dev     # 用户端   http://localhost:3001
```

> 本地开发时前端 API 地址默认指向 `http://localhost:8080`（见 `admin/src/utils/request.ts` / web 的 `VITE_API_BASE_URL`）。

---

## 四、服务与端口一览

### 微服务

| 服务 | 端口 | 服务 | 端口 |
|------|------|------|------|
| api-gateway | 8080 (HTTP) / 8095 (Swagger) | review-service | 8007 |
| user-service | 8000 | logistics-service | 8008 |
| product-service | 8081 | message-service | 8009 |
| order-service | 8082 | search-service | 8010 |
| payment-service | 8083 | recommend-service | 8011 |
| inventory-service | 8084 | file-service | 8012 |
| cart-service | 8085 | job-service | 8013 |
| promotion-service | 8006 | seckill-service | 8090 |

### 基础设施（宿主机端口）

| 组件 | 端口 | 组件 | 端口 |
|------|------|------|------|
| MySQL | 3307 | Elasticsearch | 9200 |
| Redis | 6379 | Zookeeper | 2181 |
| etcd | 2379 | Kafka | 9092 |
| MongoDB | 27017 | Prometheus / Grafana | 9090 / 3000 |

### 前端（Nginx）

| 入口 | 端口 |
|------|------|
| 管理后台 admin | 80 |
| 用户端 web | 8088 |

---

## 五、改代码后如何重启

### Go 后端

```bash
# 先单次构建镜像（避免 docker compose up --build 导致 16 个服务重复构建）
docker build -t goforge-server:latest -f deploy/docker/server.Dockerfile .

# 重启受影响的容器
cd deploy/compose && docker compose up -d --force-recreate order-service
```

### 前端

```bash
cd deploy/compose
docker compose build admin-builder web-builder
docker compose up -d --force-recreate admin-builder web-builder
```

---

## 六、常见问题

1. **启动顺序**：必须基础设施先就绪，网关最后起，`depends_on` 已自动保证；手动逐个起时按「MySQL/Redis/etcd/Kafka → 微服务 → 网关」顺序。

2. **Docker 重启后服务 500「数据库连接未初始化」**：`restart: unless-stopped` 会让容器无序同时重启、绕过依赖。等 MySQL healthy 后执行 `docker compose restart <服务名>`。

3. **前端访问 000 / 白屏**：nginx 端口转发（wslrelay）失效，`docker compose restart nginx`。

4. **端口冲突**：确认宿主机 80/8080/8088/3307/6379/9092 等端口没被其他程序占用。

5. **`down` 后再 `up`，部分容器卡在 `Created`（gateway/nginx 不启动）**：Kafka 要等 Zookeeper 先 `healthy`，启动慢会反复失败几次，导致依赖它的 `order-consumer` 和依赖链末端的 gateway/nginx 停在 `Created`。**等 Kafka 变成 `healthy` 后再执行一次 `docker compose up -d`** 即可把剩余容器拉起。

---

*关联：[GATEWAY_DEPLOYMENT.md](GATEWAY_DEPLOYMENT.md)（网关设计 + 启动依赖原理）、[README.md](../README.md)*
