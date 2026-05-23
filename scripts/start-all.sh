#!/bin/bash

# 启动所有服务脚本 (Linux/Mac)
# 包括基础设施服务（Docker Compose）和所有微服务
# 使用方法: ./scripts/start-all.sh [--build] [--gateway] [--skip-infra]

set -e

BUILD=false
GATEWAY=false
SKIP_INFRA=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --build)
            BUILD=true
            shift
            ;;
        --gateway)
            GATEWAY=true
            shift
            ;;
        --skip-infra)
            SKIP_INFRA=true
            shift
            ;;
        *)
            echo "未知参数: $1"
            echo "用法: $0 [--build] [--gateway] [--skip-infra]"
            exit 1
            ;;
    esac
done

# 颜色输出函数
print_section() {
    echo ""
    echo "============================================"
    echo "$1"
    echo "============================================"
    echo ""
}

print_info() {
    echo -e "\\033[36m$1\\033[0m"
}

print_success() {
    echo -e "\\033[32m✓ $1\\033[0m"
}

print_error() {
    echo -e "\\033[31m✗ $1\\033[0m"
}

print_warning() {
    echo -e "\\033[33m⚠ $1\\033[0m"
}

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        print_error "$1 未安装或未在 PATH 中"
        exit 1
    fi
}

# 等待服务就绪
wait_for_service() {
    local service_name=$1
    local url=$2
    local max_retries=${3:-30}
    local delay=${4:-2}

    print_info "等待 $service_name 就绪..."
    local retries=0

    while [ $retries -lt $max_retries ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            print_success "$service_name 已就绪"
            return 0
        fi
        retries=$((retries + 1))
        sleep $delay
    done

    print_warning "$service_name 在 $((max_retries * delay)) 秒内未就绪"
    return 1
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 || netstat -an 2>/dev/null | grep -q ":$port.*LISTEN"; then
        return 0
    else
        return 1
    fi
}

# 释放端口（杀掉占用端口的进程）
free_port() {
    local port=$1
    local pids=$(lsof -ti:$port 2>/dev/null)
    if [ -n "$pids" ]; then
        for pid in $pids; do
            print_warning "端口 $port 被进程 $pid 占用，正在释放..."
            kill -9 "$pid" 2>/dev/null && print_success "已释放端口 $port (进程 $pid)" || print_error "无法释放端口 $port"
        done
        sleep 1  # 等待端口释放
    fi
}

# ============================================
# 1. 启动基础设施服务（Docker Compose）
# ============================================
if [ "$SKIP_INFRA" = false ]; then
    print_section "启动基础设施服务"

    check_command docker

    COMPOSE_FILE="docker-compose-infra.yml"
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_error "Docker Compose 文件不存在: $COMPOSE_FILE"
        exit 1
    fi

    # 检查 Docker 是否运行
    if ! docker ps > /dev/null 2>&1; then
        print_error "Docker 守护进程未运行，请先启动 Docker"
        exit 1
    fi

    # 先停止并清理可能存在的旧容器（避免名称冲突）
    print_info "清理可能存在的旧容器..."
    docker compose -f "$COMPOSE_FILE" down --remove-orphans 2>/dev/null || true

    # 强制删除可能存在的容器（防止名称冲突）
    for container in infra-redis infra-etcd infra-etcd-keeper infra-mongodb infra-elasticsearch infra-prometheus; do
        if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            print_info "删除容器: $container"
            docker rm -f "$container" 2>/dev/null || true
        fi
    done

    print_info "启动 Docker Compose 服务..."
    docker compose -f "$COMPOSE_FILE" up -d

    if [ $? -ne 0 ]; then
        print_error "启动基础设施服务失败"
        exit 1
    fi

    print_info "基础设施服务启动中，等待服务就绪..."
    sleep 10  # 增加等待时间，让容器完全启动

    # 检查基础设施服务状态
    print_info "检查基础设施服务状态..."
    docker compose -f "$COMPOSE_FILE" ps

    # 验证 Redis 是否就绪
    print_info "验证 Redis 连接..."
    max_retries=10
    retries=0
    while [ $retries -lt $max_retries ]; do
        if docker exec infra-redis redis-cli PING > /dev/null 2>&1; then
            print_success "Redis 已就绪"
            break
        fi
        retries=$((retries + 1))
        if [ $retries -lt $max_retries ]; then
            print_info "等待 Redis 就绪... ($retries/$max_retries)"
            sleep 2
        fi
    done

    if [ $retries -eq $max_retries ]; then
        print_warning "Redis 在 $((max_retries * 2)) 秒内未就绪，但继续启动..."
    fi

    # 验证 Kafka 是否就绪（通过检查容器状态）
    print_info "验证 Kafka 连接..."
    max_retries=15
    retries=0
    while [ $retries -lt $max_retries ]; do
        if docker exec infra-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1 || \\
           docker ps --format '{{.Names}}' | grep -q "^infra-kafka$"; then
            # 检查 Kafka 容器是否健康
            kafka_status=$(docker inspect infra-kafka --format='{{.State.Status}}' 2>/dev/null || echo "not-found")
            if [ "$kafka_status" = "running" ]; then
                print_success "Kafka 已就绪"
                break
            fi
        fi
        retries=$((retries + 1))
        if [ $retries -lt $max_retries ]; then
            print_info "等待 Kafka 就绪... ($retries/$max_retries)"
            sleep 2
        fi
    done

    if [ $retries -eq $max_retries ]; then
        print_warning "Kafka 在 $((max_retries * 2)) 秒内未就绪，但继续启动..."
    fi

    print_success "基础设施服务启动完成"
    echo ""
fi

# ============================================
# 2. 初始化秒杀服务（生成 proto 文件）
# ============================================
print_section "初始化秒杀服务"

if command -v protoc > /dev/null; then
    if [ ! -f "api/seckill/v1/seckill.pb.go" ]; then
        print_info "生成秒杀服务 Proto 文件..."
        protoc --go_out=. --go-grpc_out=. api/seckill/v1/seckill.proto 2>/dev/null && \\
            print_success "秒杀服务 Proto 文件生成完成" || \\
            print_warning "秒杀服务 Proto 文件生成失败（可能已存在）"
    else
        print_success "秒杀服务 Proto 文件已存在"
    fi
else
    print_warning "protoc 未安装，跳过 Proto 文件生成"
fi

# ============================================
# 3. 检查配置文件
# ============================================
print_section "检查配置文件"

# 核心服务配置列表（使用普通数组，兼容 macOS bash 3.2）
# 格式: service-name:config-file:port
CORE_SERVICES=(
    "user-service:../configs/dev/user-config.yaml:8000"
    "product-service:../configs/dev/product-config.yaml:8081"
    "order-service:../configs/dev/order-config.yaml:8082"
    "payment-service:../configs/dev/payment-config.yaml:8083"
    "inventory-service:../configs/dev/inventory-config.yaml:8084"
    "cart-service:../configs/dev/cart-config.yaml:8085"
    "seckill-service:../configs/dev/seckill-config.yaml:8090"
)

# 扩展服务配置列表
EXTENDED_SERVICES=(
    "promotion-service:../configs/dev/promotion-config.yaml:8006"
    "review-service:../configs/dev/review-config.yaml:8007"
    "logistics-service:../configs/dev/logistics-config.yaml:8008"
    "message-service:../configs/dev/message-config.yaml:8009"
    "search-service:../configs/dev/search-config.yaml:8010"
    "recommend-service:../configs/dev/recommend-config.yaml:8011"
    "file-service:../configs/dev/file-config.yaml:8012"
    "job-service:../configs/dev/job-config.yaml:8013"
)

# 后台服务（不需要 gRPC 端口，后台运行）
BACKGROUND_SERVICES=(
    "order-service-consumer:../configs/dev/order-config.yaml"
)

# 检查配置文件
MISSING_CONFIGS=()
for service_config in "${CORE_SERVICES[@]}" "${EXTENDED_SERVICES[@]}" "${BACKGROUND_SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    config=$(echo "$service_config" | cut -d: -f2)

    if [ ! -f "$config" ]; then
        MISSING_CONFIGS+=("$service:$config")
        print_error "$service: $config 不存在"
    else
        print_success "$service: $config"
    fi
done

if [ ${#MISSING_CONFIGS[@]} -gt 0 ]; then
    echo ""
    print_error "以下服务的配置文件不存在:"
    for item in "${MISSING_CONFIGS[@]}"; do
        service=$(echo "$item" | cut -d: -f1)
        config=$(echo "$item" | cut -d: -f2)
        echo "  - $service: $config"
        example_file="${config}.example"
        if [ -f "$example_file" ]; then
            echo "    请复制: $example_file -> $config"
        fi
    done
    echo ""
    print_info "提示: 运行以下命令复制示例配置文件:"
    echo "  cd configs/dev && cp *.yaml.example *.yaml"
    exit 1
fi

print_success "配置文件检查通过"
echo ""

# ============================================
# 4. 编译服务（如果需要）
# ============================================
if [ "$BUILD" = true ]; then
    print_section "编译服务"

    for service_config in "${CORE_SERVICES[@]}" "${EXTENDED_SERVICES[@]}" "${BACKGROUND_SERVICES[@]}"; do
        service=$(echo "$service_config" | cut -d: -f1)
        if [ ! -f "bin/$service" ]; then
            print_info "编译 $service..."
            go build -o "bin/$service" "./cmd/$service"
            if [ $? -ne 0 ]; then
                print_error "编译失败: $service"
                exit 1
            fi
            print_success "$service 编译完成"
        else
            echo "  - $service 已存在，跳过编译"
        fi
    done

    print_success "编译完成"
    echo ""
fi

# ============================================
# 5. 检查端口占用
# ============================================
print_section "检查端口占用"

PORT_CONFLICTS=()
for service_config in "${CORE_SERVICES[@]}" "${EXTENDED_SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    port=$(echo "$service_config" | cut -d: -f3)

    if check_port $port; then
        PORT_CONFLICTS+=("$service:$port")
        print_warning "端口 $port 已被占用 ($service)，正在自动释放..."
        free_port $port
        # 再次检查端口是否已释放
        if check_port $port; then
            print_error "端口 $port 释放失败 ($service)"
        else
            print_success "端口 $port 已释放 ($service)"
        fi
    else
        print_success "端口 $port 可用 ($service)"
    fi
done

if [ "$GATEWAY" = true ]; then
    GATEWAY_PORT=8080
    if check_port $GATEWAY_PORT; then
        print_warning "端口 $GATEWAY_PORT 已被占用 (api-gateway)，正在自动释放..."
        free_port $GATEWAY_PORT
        # 再次检查端口是否已释放
        if check_port $GATEWAY_PORT; then
            print_error "端口 $GATEWAY_PORT 释放失败 (api-gateway)"
        else
            print_success "端口 $GATEWAY_PORT 已释放 (api-gateway)"
        fi
    else
        print_success "端口 $GATEWAY_PORT 可用 (api-gateway)"
    fi
fi

if [ ${#PORT_CONFLICTS[@]} -gt 0 ]; then
    echo ""
    print_info "已尝试自动释放被占用的端口"
    echo ""
fi

# ============================================
# 6. 创建日志目录
# ============================================
mkdir -p logs
print_success "日志目录: logs"

# ============================================
# 7. 启动所有微服务
# ============================================
print_section "启动微服务"

PIDS=()
STARTED_SERVICES=()

# 启动核心服务
print_info "启动核心服务..."
for service_config in "${CORE_SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    config=$(echo "$service_config" | cut -d: -f2)
    port=$(echo "$service_config" | cut -d: -f3)

    print_info "启动 $service (端口: $port)..."

    # 启动前再次检查并释放端口（防止在检查后端口被占用）
    if check_port $port; then
        print_warning "端口 $port 被占用，正在释放..."
        free_port $port
    fi

    log_file="logs/${service}.log"

    # 检查二进制文件是否存在且可执行（排除 Windows .exe 文件）
    if [ -f "bin/$service" ] && [ ! -f "bin/$service.exe" ] && file "bin/$service" 2>/dev/null | grep -q "executable" && ! file "bin/$service" 2>/dev/null | grep -q "PE32"; then
        ./bin/$service -f "$config" > "$log_file" 2>&1 &
    else
        # 使用 go run 作为回退方案
        go run "cmd/$service/main.go" -f "$config" > "$log_file" 2>&1 &
    fi

    PIDS+=($!)
    STARTED_SERVICES+=("$service:$port")
    sleep 3  # 增加等待时间，让服务完全启动
done

# 等待核心服务就绪（特别是 order-service，API Gateway 依赖它）
print_info "等待核心服务就绪..."
sleep 5

# 验证 order-service 是否就绪
if echo "${CORE_SERVICES[@]}" | grep -q "order-service"; then
    print_info "验证 order-service 连接..."
    max_retries=10
    retries=0
    while [ $retries -lt $max_retries ]; do
        # 检查端口是否在监听
        if check_port 8082; then
            print_success "order-service 已就绪 (端口 8082)"
            break
        fi
        retries=$((retries + 1))
        if [ $retries -lt $max_retries ]; then
            print_info "等待 order-service 就绪... ($retries/$max_retries)"
            sleep 2
        fi
    done

    if [ $retries -eq $max_retries ]; then
        print_warning "order-service 在 $((max_retries * 2)) 秒内未就绪，API Gateway 可能无法连接"
    fi
fi

# 启动扩展服务
print_info "启动扩展服务..."
for service_config in "${EXTENDED_SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    config=$(echo "$service_config" | cut -d: -f2)
    port=$(echo "$service_config" | cut -d: -f3)

    print_info "启动 $service (端口: $port)..."

    # 启动前再次检查并释放端口（防止在检查后端口被占用）
    if check_port $port; then
        print_warning "端口 $port 被占用，正在释放..."
        free_port $port
    fi

    log_file="logs/${service}.log"

    # 检查二进制文件是否存在且可执行（排除 Windows .exe 文件）
    if [ -f "bin/$service" ] && [ ! -f "bin/$service.exe" ] && file "bin/$service" 2>/dev/null | grep -q "executable" && ! file "bin/$service" 2>/dev/null | grep -q "PE32"; then
        ./bin/$service -f "$config" > "$log_file" 2>&1 &
    else
        # 使用 go run 作为回退方案
        go run "cmd/$service/main.go" -f "$config" > "$log_file" 2>&1 &
    fi

    PIDS+=($!)
    STARTED_SERVICES+=("$service:$port")
    sleep 2
done

# 启动后台服务（不需要端口检查）
print_info "启动后台服务..."
for service_config in "${BACKGROUND_SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    config=$(echo "$service_config" | cut -d: -f2)

    print_info "启动 $service (后台服务)..."

    log_file="logs/${service}.log"

    # 检查二进制文件是否存在且可执行
    if [ -f "bin/$service" ] && [ ! -f "bin/$service.exe" ] && file "bin/$service" 2>/dev/null | grep -q "executable" && ! file "bin/$service" 2>/dev/null | grep -q "PE32"; then
        ./bin/$service -f "$config" > "$log_file" 2>&1 &
    else
        # 使用 go run 作为回退方案
        go run "cmd/$service/main.go" -f "$config" > "$log_file" 2>&1 &
    fi

    PIDS+=($!)
    STARTED_SERVICES+=("$service:background")
    sleep 1
done

# 启动 API 网关（如果指定）
if [ "$GATEWAY" = true ]; then
    echo ""
    print_info "启动 API 网关..."
    GATEWAY_PORT=8080

    # 启动前再次检查并释放端口
    if check_port $GATEWAY_PORT; then
        print_warning "端口 $GATEWAY_PORT 被占用，正在释放..."
        free_port $GATEWAY_PORT
    fi

    GATEWAY_CONFIG="../configs/dev/gateway.yaml"
    if [ -f "$GATEWAY_CONFIG" ]; then
        log_file="logs/api-gateway.log"

        if [ -f "bin/api-gateway" ]; then
            ./bin/api-gateway -f "$GATEWAY_CONFIG" > "$log_file" 2>&1 &
        else
            go run cmd/api-gateway/main.go -f "$GATEWAY_CONFIG" > "$log_file" 2>&1 &
        fi

        PIDS+=($!)
        print_success "API 网关已启动 (端口: 8080)"
    else
        print_warning "网关配置文件不存在: $GATEWAY_CONFIG"
    fi
fi

# ============================================
# 8. 显示启动信息
# ============================================
echo ""
print_section "启动完成"

print_info "已启动的服务:"
echo ""

echo "基础设施服务:"
echo "  - Redis: localhost:6379"
echo "  - etcd: localhost:2379"
echo "  - MongoDB: localhost:27017"
echo "  - Elasticsearch: localhost:9200"
echo "  - Prometheus: http://localhost:9090"
echo "  - Kafka: localhost:9092"
echo "  - Kafka UI: http://localhost:18090"
echo ""

echo "微服务:"
for item in "${STARTED_SERVICES[@]}"; do
    service=$(echo "$item" | cut -d: -f1)
    port=$(echo "$item" | cut -d: -f2)
    if [ "$port" = "background" ]; then
        echo "  - $service: 后台服务"
    else
        echo "  - $service: gRPC 端口 $port"
    fi
done

if [ "$GATEWAY" = true ]; then
    echo "  - api-gateway: HTTP http://localhost:8080"
fi

echo ""
print_info "日志文件位置: logs/"
print_info "查看日志: tail -f logs/<service-name>.log"
echo ""
print_warning "按 Ctrl+C 停止所有服务"
echo ""

# ============================================
# 9. 等待用户中断并清理
# ============================================
cleanup() {
    echo ""
    print_section "正在停止所有服务"

    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            print_info "停止进程 $pid..."
            kill "$pid" 2>/dev/null || true
            wait "$pid" 2>/dev/null || true
        fi
    done

    echo ""
    print_success "所有服务已停止"
    exit 0
}

trap cleanup INT TERM

# 监控服务状态
while true; do
    dead_pids=()
    for i in "${!PIDS[@]}"; do
        pid=${PIDS[$i]}
        if ! kill -0 "$pid" 2>/dev/null; then
            dead_pids+=("$i")
        fi
    done

    if [ ${#dead_pids[@]} -gt 0 ]; then
        # 收集实际退出的服务名称
        exited_services=()
        for i in "${dead_pids[@]}"; do
            if [ $i -lt ${#STARTED_SERVICES[@]} ]; then
                # 微服务
                service=$(echo "${STARTED_SERVICES[$i]}" | cut -d: -f1)
                exited_services+=("$service")
            elif [ "$GATEWAY" = true ] && [ $i -ge ${#STARTED_SERVICES[@]} ]; then
                # API Gateway（在 STARTED_SERVICES 之后添加的）
                exited_services+=("api-gateway")
            fi
        done

        # 只有在确实有服务退出时才显示提示
        if [ ${#exited_services[@]} -gt 0 ]; then
            echo ""
            print_warning "以下服务已退出:"
            for service in "${exited_services[@]}"; do
                print_error "$service"
            done
        fi
    fi

    sleep 5
done
