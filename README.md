# GoForge

基于 go-zero 的电商微服务项目：15 个业务服务 + API 网关 + 订单消息消费者，前后端分离，支持 Docker Compose 一键启动与本地开发。

[English](README.en.md) | 中文

## 功能

| 模块 | 能力 |
|---|---|
| 用户 | 注册、登录（JWT）、资料维护、收货地址、积分签到 |
| 商品 | 分类、品牌、SPU/SKU、Banner、库存、上下架 |
| 交易 | 购物车（Redis）、下单、支付（多渠道与模拟回调）、取消、退款 |
| 营销 | 优惠券（领券中心、我的券、试算、下单核销、取消与退款退券）、秒杀 |
| 支撑 | 搜索（Elasticsearch）、推荐、评价（MySQL + MongoDB 详情）、物流、站内消息（后台群发与事件驱动通知） |
| 治理 | 网关按 IP 限流、CORS、Prometheus 指标、Kafka 事件、etcd 服务发现 |

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25、go-zero 1.10（gRPC、Gateway）、GORM |
| 存储 | MySQL 8.0、Redis 7、MongoDB 7、Elasticsearch 8.15 |
| 中间件 | etcd 3.5、Kafka 7.5、ZooKeeper 7.5 |
| 前端 | Vue 3、Vite 5、Element Plus、Pinia、Axios |
| 入口 | go-zero Gateway、Nginx |
| 监控 | Prometheus 2.54、Grafana |
| 部署 | Docker Compose、Kubernetes |

## 架构

```
   admin(:80)      web(:8088)
        \             /
            Nginx
              |
      API Gateway :8080  ——  Swagger UI :8095
              |            (configs/*/gateway.yaml 维护 HTTP → gRPC 映射)
   ┌──────────┼──────────────────────────────────────────┐
 user      product      order      payment      inventory
 cart      promotion    review     logistics    message
 search    recommend    file       job          seckill
   └──────────┼──────────────────────────────────────────┘
              |
   MySQL · Redis · Kafka · etcd · Elasticsearch · MongoDB
```

服务通过 etcd 注册与发现（key 为 `<service>.rpc`）；`order-service-consumer` 订阅 Kafka 中的秒杀订单与支付结果事件，负责落单与订单状态更新。

## 目录结构

```
GoForge/
├── admin/                    管理后台（Vue 3）
├── web/                      用户端（Vue 3）
├── server/                   Go 后端
│   ├── api/                  Protobuf 定义与生成代码
│   ├── cmd/                  服务入口、cmd/tools 自检工具
│   ├── internal/             handler、middleware、service、pkg
│   ├── database/             schema.sql、seed.sql、sharding.sql
│   └── docs/swagger/         网关接口文档
├── configs/                  dev、docker、k8s 三套配置
├── deploy/                   compose、docker、k8s、nginx、prometheus、grafana
├── scripts/                  初始化与启停脚本
└── Makefile
```

## 服务与端口

| 服务 | 端口 | 说明 |
|---|---|---|
| api-gateway | 8080 | API 网关（8095 为接口文档） |
| user-service | 8000 | 用户、地址、积分 |
| product-service | 8081 | 商品、分类、SKU、Banner |
| order-service | 8082 | 订单 |
| payment-service | 8083 | 支付与退款 |
| inventory-service | 8084 | 库存 |
| cart-service | 8085 | 购物车 |
| promotion-service | 8006 | 优惠券、活动 |
| review-service | 8007 | 评价 |
| logistics-service | 8008 | 物流 |
| message-service | 8009 | 站内消息 |
| search-service | 8010 | 搜索与索引同步 |
| recommend-service | 8011 | 推荐 |
| file-service | 8012 | 文件上传 |
| job-service | 8013 | 定时任务 |
| seckill-service | 8090 | 秒杀 |
| order-service-consumer | — | Kafka 消费者 |

基础设施在宿主机的映射端口：MySQL 3307、Redis 6379、etcd 12379、Kafka 9092、ZooKeeper 2181、Elasticsearch 9200、MongoDB 27017、Prometheus 9090、Grafana 3000。

## 快速开始

环境要求：Go 1.25+、Node.js 22+、Docker Desktop（Compose v2）。

```bash
git clone git@github.com:zhian9/GoForge.git
cd GoForge/deploy/compose
docker compose up -d
docker compose ps
```

MySQL 容器首次启动时会自动执行 `server/database/schema.sql` 与 `seed.sql`。

| 入口 | 地址 |
|---|---|
| 用户端 | http://localhost:8088 |
| 管理后台 | http://localhost |
| API 网关 | http://localhost:8080 |
| 接口文档 | http://localhost:8095 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 |

演示账号（来自 seed 数据）：管理后台 `admin / admin123`，用户端 `demo1 / 123456`。对外部署前请修改 `deploy/.env` 中的数据库密码与 `JWT_SECRET`。

## 本地开发

方式一：基础设施运行在 Docker，服务运行在宿主机

```bash
make start-infra      # MySQL / Redis / etcd / Kafka / Elasticsearch / MongoDB
make build            # 编译 15 个业务服务与网关到 server/bin/
make start-services   # 启动全部服务
```

方式二：dev 挂载模式（修改代码无需重建镜像）

```bash
make dev-build                     # 交叉编译（linux/amd64）到 server/dev-bin/
make dev-up                        # 使用 docker-compose.dev.yml 启动
make dev-restart SVC=order         # 重启单个服务
```

| 改动内容 | 操作 |
|---|---|
| 配置 | 修改 `configs/docker/*.yaml`，然后 `docker compose -f docker-compose.yml -f docker-compose.dev.yml restart <service>` |
| Go 代码 | `make dev-build`，然后 `make dev-restart SVC=<service>` |
| 前端 | 在 `web/` 或 `admin/` 执行 `npm run build`，刷新浏览器 |

前端单独开发：`cd web && npm install && npm run dev`（:3001）、`cd admin && npm install && npm run dev`（:3000），两者都已配置 `/api`、`/uploads`、`/images` 代理到网关。

## 配置

| 目录 | 用途 |
|---|---|
| `configs/dev` | 宿主机直连（`127.0.0.1`） |
| `configs/docker` | 容器网络（`mysql:3306`、`redis:6379`、`etcd:2379`、`kafka:29092`） |
| `configs/k8s` | Kubernetes 混合模式（`172.18.0.1`） |

网关的全部 HTTP 路由（HTTP → gRPC 映射）维护在 `configs/*/gateway.yaml` 的 `Upstreams.Mappings` 中；密码与密钥通过环境变量注入（`deploy/.env`、`deploy/k8s/secret.yaml`）。

## 数据库

| 文件 | 说明 |
|---|---|
| `server/database/schema.sql` | 表结构 |
| `server/database/seed.sql` | 演示数据 |
| `server/database/sharding.sql` | 可选分表示例 |
| `scripts/init.sh`、`scripts/init.bat` | 本地初始化脚本 |

## 接口文档

```bash
make swagger
```

输出 `server/docs/swagger/api/<service>/v1/<service>.swagger.json` 与 `index.json`，网关 :8095 的 Swagger UI 按 `index.json` 加载。文档由 `cmd/generate-swagger` 依据 `configs/dev/gateway.yaml` 生成。

## 构建与测试

```bash
make build          # 编译后端
make test           # Go 测试
make lint           # golangci-lint（需自行安装）
cd web && npm run build
cd admin && npm run build
```

仓库不包含 `*.pb.go`，新克隆后先执行 `make proto`（需要 `protoc`、`protoc-gen-go`、`protoc-gen-go-grpc`）。

## 部署

| 资源 | 位置 |
|---|---|
| Docker Compose | `deploy/compose/docker-compose.yml`，可通过 `deploy/.env` 覆盖镜像 tag 与密码 |
| Kubernetes | `deploy/k8s/`（namespace、deployments、services、configmap、secret.example.yaml） |
| Nginx | `deploy/nginx/default.conf`（80 管理后台、8088 用户端，`/api`、`/uploads` 反代到网关） |

## 辅助工具

| 命令 | 说明 |
|---|---|
| `make auth-check` | 鉴权自检（越权访问、伪造 token） |
| `make seckill-check` | 秒杀链路自检（超卖、重复下单、账目一致性） |
| `make status` | 检查服务端口 |
| `make redis-cli` | 进入 Redis 容器 |

## License

MIT
