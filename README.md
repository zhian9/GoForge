# GoForge

基于 **go-zero** 的 Go 微服务电商系统，15 个业务微服务 + API 网关 + 消息消费者，前后端分离，Docker Compose 一键部署。

## 技术栈

| 层 | 技术 |
|---|------|
| 框架 | go-zero v1.10 (gRPC + Gateway) |
| 数据库 | MySQL 8.0 (GORM) |
| 缓存 | Redis 7 |
| 注册中心 | etcd |
| 消息队列 | Kafka |
| 搜索引擎 | Elasticsearch 8 |
| 文档存储 | MongoDB 7 |
| 监控 | Prometheus + Grafana |
| 前端 | Vue 3 + Element Plus + Vite |
| 网关 | go-zero Gateway + Nginx |
| 部署 | Docker Compose (28 容器) |

## 项目结构

```
GoForge/
├── server/                     # Go 后端
│   ├── cmd/                    # 17 个服务入口
│   │   ├── api-gateway/        # API 网关 (:8080)
│   │   ├── user-service/       # 用户服务 (:8000)
│   │   ├── product-service/    # 商品服务 (:8081)
│   │   ├── order-service/      # 订单服务 (:8082)
│   │   ├── payment-service/    # 支付服务 (:8083)
│   │   ├── inventory-service/  # 库存服务 (:8084)
│   │   ├── cart-service/       # 购物车 (:8085)
│   │   ├── promotion-service/  # 营销服务 (:8006)
│   │   ├── review-service/     # 评价服务 (:8007)
│   │   ├── logistics-service/  # 物流服务 (:8008)
│   │   ├── message-service/    # 消息服务 (:8009)
│   │   ├── search-service/     # 搜索服务 (:8010)
│   │   ├── recommend-service/  # 推荐服务 (:8011)
│   │   ├── file-service/       # 文件服务 (:8012)
│   │   ├── job-service/        # 任务服务 (:8013)
│   │   └── seckill-service/    # 秒杀服务 (:8090)
│   ├── api/                    # Protobuf 接口定义
│   ├── internal/               # 内部包
│   │   ├── service/            # 业务逻辑 (事务脚本模式)
│   │   ├── handler/            # HTTP 处理器
│   │   ├── middleware/         # gRPC/HTTP 中间件
│   │   └── pkg/               # 共享库 (cache/database/mq/...)
│   └── database/               # SQL Schema + 种子数据
├── configs/                    # 配置文件
│   ├── dev/                    # 开发环境 (硬编码)
│   ├── docker/                 # Docker 环境 (${ENV} 占位符)
│   ├── test/                   # 测试环境
│   └── prod/                   # 生产环境
├── admin/                      # Vue3 管理后台 (:80)
├── web/                        # Vue3 用户端 (:8088)
├── deploy/                     # Docker 部署
│   ├── compose/                # docker-compose.yml
│   ├── docker/                 # Dockerfiles
│   ├── nginx/                  # Nginx 反向代理
│   ├── prometheus/             # Prometheus 配置
│   └── grafana/                # Grafana 面板
├── scripts/                    # 启动/停止/初始化脚本
└── Makefile                    # 构建管理
```

## 架构

```
                    ┌─────────────┐
                    │   Nginx :80 │
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │  Admin   │    │   Web    │    │   API    │
    │  :3000   │    │  :3001   │    │  :8080   │
    └──────────┘    └──────────┘    └────┬─────┘
                                        │
    ┌───────────┬───────────┬───────────┼───────────┬───────────┐
    ▼           ▼           ▼           ▼           ▼           ▼
┌──────┐  ┌──────┐  ┌────────┐  ┌──────┐  ┌────────┐  ┌──────────┐
│ User │  │Product│  │ Order  │  │Payment│  │Inventory│  │ Seckill  │
│:8000 │  │:8081 │  │ :8082  │  │:8083 │  │ :8084  │  │  :8090   │
└──┬───┘  └──┬───┘  └───┬────┘  └──┬───┘  └───┬────┘  └────┬─────┘
   │         │          │         │          │            │
   └─────────┴──────────┴─────────┴──────────┴────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
   ┌─────────┐          ┌─────────┐          ┌──────────┐
   │  MySQL  │          │  Redis  │          │  Kafka   │
   │  :3306  │          │  :6379  │          │  :9092   │
   └─────────┘          └─────────┘          └──────────┘
```

## 快速开始

### 环境要求

- Go 1.25+
- Docker Desktop 28+
- MySQL 8.0+
- Node.js 22+

### 本地开发

```bash
# 1. 克隆
git clone git@github.com:zhian9/GoForge.git
cd GoForge

# 2. 初始化数据库
mysql -uroot -p < server/database/schema.sql
# 创建 goforge 用户
mysql -uroot -p -e "CREATE USER 'goforge'@'localhost' IDENTIFIED BY '123456';
                     GRANT ALL ON go_forge.* TO 'goforge'@'localhost';"

# 3. 启动基础设施
docker compose -f server/docker-compose-infra.yml up -d

# 4. 编译 + 启动
make build
make start-services

# 5. 启动前端 (可选)
cd admin && npm install && npm run dev    # 管理后台 :3000
cd web && npm install && npm run dev      # 用户端 :3001
```

### Docker 部署（推荐）

```bash
cd deploy/compose
cp ../.env.example ../.env      # 编辑密码
docker compose up -d            # 启动全部容器
```

> 改 Go 代码后重新部署：先单次构建镜像再拉起（避免 `--build` 导致 16 个服务重复构建）：
> ```bash
> docker build -t goforge-server:latest -f deploy/docker/server.Dockerfile .
> docker compose up -d --force-recreate <service>
> ```
> 前端改动：`docker compose build admin-builder web-builder && docker compose up -d --force-recreate admin-builder web-builder`

## 测试账号

| 角色 | 地址 | 账号 / 密码 |
|------|------|-------------|
| 管理后台 | http://localhost/ | admin / admin123 |
| 用户端 | http://localhost:8088/ | 注册即用 |

## 核心业务闭环

- 商品浏览 / 分类 / 搜索 / 秒杀
- 购物车 → 下单（原子扣减库存，防超卖）
- 支付（Mock 渠道 + 幂等乐观锁）→ 订单待发货
- 管理员发货 → 物流轨迹
- 确认收货 → 评价
- 取消 / 退款（回增库存）
- 签到积分（下单支付 1 元 = 1 积分）

## API 端点

| 服务 | 路径前缀 |
|------|---------|
| 用户 | `/api/v1/user/*` |
| 商品 | `/api/v1/products/*` |
| 订单 | `/api/v1/orders/*` |
| 支付 | `/api/v1/payments/*` |
| 购物车 | `/api/v1/cart/*` |
| 库存 | `/api/v1/inventory/*` |
| 秒杀 | `/api/v1/seckill/*` |
| 营销 | `/api/v1/promotion/*` |
| 评价 | `/api/v1/reviews/*` |
| 搜索 | `/api/v1/search/*` |
| 文件 | `/api/v1/files/*` |

完整接口文档: http://localhost:8095 (Swagger UI)

### 快速测试

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"hello","password":"Hello123","login_type":1}'

# 登录
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"hello","password":"Hello123","login_type":1}'

# 商品列表
curl http://localhost:8080/api/v1/products
```

## 监控

| 服务 | 地址 | 账号 |
|------|------|------|
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin |
| Swagger | http://localhost:8095 | - |
| etcd Keeper | http://localhost:8089 | - |
| Kafka UI | http://localhost:18090 | - |

## Make 命令

```bash
make help          # 查看所有命令
make build         # 编译全部服务
make start-infra   # 启动基础设施
make start-all     # 完整启动
make stop-all      # 停止所有服务
make status        # 查看端口状态
make run SVC=user  # 编译运行单个服务
make clean         # 清理编译产物
make test          # 运行测试
```
