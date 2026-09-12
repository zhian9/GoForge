.PHONY: help build test lint clean proto deps start-infra stop-infra start-all stop-all status redis-cli run

# ============================================
# GoForge - Makefile
# 在项目根目录运行: make <target>
# ============================================

SERVER_DIR  := server
CONFIG_DIR  := configs/dev
SCRIPTS_DIR := scripts
DEPLOY_DIR  := deploy/compose

GOCMD   := go
GOBUILD := $(GOCMD) build
GOTEST  := $(GOCMD) test
GOMOD   := $(GOCMD) mod

SERVICES := user product order payment inventory cart promotion review logistics message search recommend file job seckill
GATEWAY  := api-gateway

help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================
# 构建
# ============================================

build: ## 编译所有服务到 server/bin/
	@echo "Building all services..."
	@mkdir -p $(SERVER_DIR)/bin
	@cd $(SERVER_DIR) && for svc in $(SERVICES); do \
		echo "  $$svc-service"; \
		$(GOBUILD) -o "bin/$$svc-service.exe" "./cmd/$$svc-service"; \
	done
	@cd $(SERVER_DIR) && echo "  $(GATEWAY)" && $(GOBUILD) -o "bin/$(GATEWAY).exe" "./cmd/$(GATEWAY)"
	@echo "Done. Binaries in server/bin/"

build-service: ## 编译单个服务 (usage: make build-service SVC=user)
	@cd $(SERVER_DIR) && $(GOBUILD) -o "bin/$(SVC)-service.exe" "./cmd/$(SVC)-service"

clean: ## 清理编译产物
	@rm -rf $(SERVER_DIR)/bin/
	@echo "Cleaned."

test: ## 运行测试
	@cd $(SERVER_DIR) && $(GOTEST) -v ./... -cover

lint: ## 运行 lint
	@cd $(SERVER_DIR) && golangci-lint run 2>/dev/null || echo "golangci-lint not installed, skip"

deps: ## 下载依赖
	@cd $(SERVER_DIR) && $(GOMOD) download && $(GOMOD) tidy

proto: ## 生成 Protobuf 代码
	@cd $(SERVER_DIR) && find api -name "*.proto" -exec protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;

# ============================================
# Docker 基础设施
# ============================================

start-infra: ## 启动基础设施 (MySQL/Redis/etcd/Mongo/ES/Kafka)
	@echo "Starting infrastructure..."
	@cd $(DEPLOY_DIR) && docker compose --progress plain up -d mysql redis etcd mongodb elasticsearch zookeeper kafka
	@echo "Infrastructure started."

stop-infra: ## 停止基础设施
	@cd $(DEPLOY_DIR) && docker compose stop mysql redis etcd mongodb elasticsearch zookeeper kafka
	@echo "Infrastructure stopped."

# ============================================
# 服务管理
# ============================================

start-services: ## 启动所有微服务（仅服务，不含基础设施）
	@echo "Starting all microservices..."
	@cd $(SERVER_DIR) && for svc in $(SERVICES); do \
		nohup "bin/$$svc-service.exe" -f "../$(CONFIG_DIR)/$$svc-config.yaml" > "logs/$$svc-service.log" 2>&1 & \
		echo "  $$svc-service started"; \
		sleep 0.3; \
	done
	@sleep 2
	@cd $(SERVER_DIR) && nohup "bin/$(GATEWAY).exe" -f "../$(CONFIG_DIR)/gateway.yaml" > "logs/$(GATEWAY).log" 2>&1 &
	@echo "Gateway started. All services launched."

start-all: start-infra ## 完整启动（基础设施 + 编译 + 所有服务）
	@$(MAKE) build
	@echo "Waiting for infrastructure..."
	@sleep 10
	@$(MAKE) start-services
	@echo "GoForge is running!"

stop-all: ## 停止所有微服务
	@powershell -Command "Get-Process -Name 'user-service','product-service','order-service','payment-service','inventory-service','cart-service','promotion-service','review-service','logistics-service','message-service','search-service','recommend-service','file-service','job-service','seckill-service','api-gateway' -ErrorAction SilentlyContinue | Stop-Process -Force" 2>/dev/null
	@echo "All services stopped."

status: ## 检查服务状态
	@echo "=== Port Status ==="
	@for port in 8000 8081 8082 8083 8084 8085 8006 8007 8008 8009 8010 8011 8012 8013 8090 8080; do \
		if netstat -ano 2>/dev/null | grep -q ":$$port .*LISTENING"; then echo "  :$$port OK"; else echo "  :$$port MISSING"; fi; \
	done

run: ## 运行单个服务 (usage: make run SVC=user)
	@cd $(SERVER_DIR) && $(GOBUILD) -o "bin/$(SVC)-service.exe" "./cmd/$(SVC)-service" && "./bin/$(SVC)-service.exe" -f "../$(CONFIG_DIR)/$(SVC)-config.yaml"

# ============================================
# 工具
# ============================================

redis-cli: ## 连接 Redis CLI
	@docker exec -it goforge-redis redis-cli -a $${REDIS_PASSWORD:-508065}

.DEFAULT_GOAL := help
