# GoForge

An e-commerce microservice project built on go-zero: 15 business services, an API gateway and an order message consumer, with separated frontend and backend, one-command Docker Compose startup and a local development workflow.

English | [中文](README.md)

## Features

| Module | Capabilities |
|---|---|
| User | Registration, login (JWT), profile, shipping addresses, points and check-in |
| Product | Categories, brands, SPU/SKU, banners, stock, on/off shelf |
| Trade | Shopping cart (Redis), order placement, payment (multiple channels and mock callback), cancellation, refund |
| Marketing | Coupons (coupon center, my coupons, discount preview, redemption on order, return on cancel/refund), flash sale |
| Supporting | Search (Elasticsearch), recommendation, reviews (MySQL with MongoDB details), logistics, in-app messages (broadcast and event-driven notifications) |
| Governance | Per-IP rate limiting at the gateway, CORS, Prometheus metrics, Kafka events, etcd service discovery |

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25, go-zero 1.10 (gRPC, Gateway), GORM |
| Storage | MySQL 8.0, Redis 7, MongoDB 7, Elasticsearch 8.15 |
| Middleware | etcd 3.5, Kafka 7.5, ZooKeeper 7.5 |
| Frontend | Vue 3, Vite 5, Element Plus, Pinia, Axios |
| Entry | go-zero Gateway, Nginx |
| Monitoring | Prometheus 2.54, Grafana |
| Deployment | Docker Compose, Kubernetes |

## Architecture

```
   admin(:80)      web(:8088)
        \             /
            Nginx
              |
      API Gateway :8080  ——  Swagger UI :8095
              |            (HTTP → gRPC routes in configs/*/gateway.yaml)
   ┌──────────┼──────────────────────────────────────────┐
 user      product      order      payment      inventory
 cart      promotion    review     logistics    message
 search    recommend    file       job          seckill
   └──────────┼──────────────────────────────────────────┘
              |
   MySQL · Redis · Kafka · etcd · Elasticsearch · MongoDB
```

Services register in etcd under `<service>.rpc` and are discovered by the gateway. `order-service-consumer` subscribes to flash-sale order and payment result events on Kafka.

## Project Layout

```
GoForge/
├── admin/                    Admin panel (Vue 3)
├── web/                      Customer site (Vue 3)
├── server/                   Go backend
│   ├── api/                  Protobuf definitions and generated code
│   ├── cmd/                  Service entrypoints and cmd/tools self-check tools
│   ├── internal/             handler, middleware, service, pkg
│   ├── database/             schema.sql, seed.sql, sharding.sql
│   └── docs/swagger/         Gateway API documentation
├── configs/                  dev, docker and k8s configurations
├── deploy/                   compose, docker, k8s, nginx, prometheus, grafana
├── scripts/                  Initialization and startup scripts
└── Makefile
```

## Services and Ports

| Service | Port | Description |
|---|---|---|
| api-gateway | 8080 | API gateway (8095 serves the API docs) |
| user-service | 8000 | Users, addresses, points |
| product-service | 8081 | Products, categories, SKUs, banners |
| order-service | 8082 | Orders |
| payment-service | 8083 | Payments and refunds |
| inventory-service | 8084 | Stock |
| cart-service | 8085 | Shopping cart |
| promotion-service | 8006 | Coupons and campaigns |
| review-service | 8007 | Reviews |
| logistics-service | 8008 | Logistics |
| message-service | 8009 | In-app messages |
| search-service | 8010 | Search and index synchronization |
| recommend-service | 8011 | Recommendations |
| file-service | 8012 | File upload |
| job-service | 8013 | Scheduled jobs |
| seckill-service | 8090 | Flash sale |
| order-service-consumer | — | Kafka consumer |

Host ports of the infrastructure: MySQL 3307, Redis 6379, etcd 12379, Kafka 9092, ZooKeeper 2181, Elasticsearch 9200, MongoDB 27017, Prometheus 9090, Grafana 3000.

## Quick Start

Requirements: Go 1.25+, Node.js 22+, Docker Desktop (Compose v2).

```bash
git clone git@github.com:zhian9/GoForge.git
cd GoForge/deploy/compose
docker compose up -d
docker compose ps
```

On first start the MySQL container runs `server/database/schema.sql` and `seed.sql` automatically.

| Entry | URL |
|---|---|
| Customer site | http://localhost:8088 |
| Admin panel | http://localhost |
| API gateway | http://localhost:8080 |
| API docs | http://localhost:8095 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 |

Demo accounts from the seed data: admin panel `admin / admin123`, customer site `demo1 / 123456`. Change the database passwords and `JWT_SECRET` in `deploy/.env` before exposing the stack.

## Local Development

Option 1: infrastructure in Docker, services on the host

```bash
make start-infra      # MySQL / Redis / etcd / Kafka / Elasticsearch / MongoDB
make build            # build the 15 services and the gateway into server/bin/
make start-services
```

Option 2: dev bind-mount mode (no image rebuild when code changes)

```bash
make dev-build                     # cross-compile (linux/amd64) into server/dev-bin/
make dev-up                        # start with docker-compose.dev.yml
make dev-restart SVC=order         # restart a single service
```

| Change | Action |
|---|---|
| Configuration | Edit `configs/docker/*.yaml`, then `docker compose -f docker-compose.yml -f docker-compose.dev.yml restart <service>` |
| Go code | `make dev-build`, then `make dev-restart SVC=<service>` |
| Frontend | `npm run build` in `web/` or `admin/`, then refresh the browser |

Standalone frontend development: `cd web && npm install && npm run dev` (:3001) and `cd admin && npm install && npm run dev` (:3000). Both proxy `/api`, `/uploads` and `/images` to the gateway.

## Configuration

| Directory | Usage |
|---|---|
| `configs/dev` | Direct host access (`127.0.0.1`) |
| `configs/docker` | Container network (`mysql:3306`, `redis:6379`, `etcd:2379`, `kafka:29092`) |
| `configs/k8s` | Kubernetes hybrid mode (`172.18.0.1`) |

All HTTP routes (HTTP → gRPC mappings) live in `Upstreams.Mappings` of `configs/*/gateway.yaml`. Passwords and secrets are injected through environment variables (`deploy/.env`, `deploy/k8s/secret.yaml`).

## Database

| File | Description |
|---|---|
| `server/database/schema.sql` | Table definitions |
| `server/database/seed.sql` | Demo data |
| `server/database/sharding.sql` | Optional sharding example |
| `scripts/init.sh`, `scripts/init.bat` | Local initialization scripts |

## API Documentation

```bash
make swagger
```

Generates `server/docs/swagger/api/<service>/v1/<service>.swagger.json` plus `index.json`; the Swagger UI on :8095 loads the documents listed in `index.json`. Documents are produced by `cmd/generate-swagger` from `configs/dev/gateway.yaml`.

## Build and Test

```bash
make build          # build the backend
make test           # Go tests
make lint           # golangci-lint (install separately)
cd web && npm run build
cd admin && npm run build
```

The repository does not contain `*.pb.go` files; run `make proto` after cloning (requires `protoc`, `protoc-gen-go` and `protoc-gen-go-grpc`).

## Deployment

| Resource | Location |
|---|---|
| Docker Compose | `deploy/compose/docker-compose.yml`; override tag and passwords with `deploy/.env` |
| Kubernetes | `deploy/k8s/` (namespace, deployments, services, configmap, secret.example.yaml) |
| Nginx | `deploy/nginx/default.conf` (admin panel on 80, customer site on 8088, `/api` and `/uploads` proxied to the gateway) |

## Tools

| Command | Description |
|---|---|
| `make auth-check` | Authorization self-check (privilege escalation, forged tokens) |
| `make seckill-check` | Flash-sale self-check (overselling, duplicate orders, consistency) |
| `make status` | Check service ports |
| `make redis-cli` | Open a Redis shell in the container |

## License

MIT
