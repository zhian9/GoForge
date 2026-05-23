#!/bin/bash
# ============================================
# GoForge E-Commerce 一键启动脚本
# 用法: bash start.sh            # 完整启动
#       bash start.sh -no-build  # 跳过编译
#       bash start.sh -no-infra  # 跳过 Docker
# ============================================
set -e

# 加载 .env
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/../.env" ]; then
  set -a && . "$SCRIPT_DIR/../.env" && set +a
  echo "Loaded .env"
fi

PROJECT_ROOT="$SCRIPT_DIR/../server"
BIN_DIR="$PROJECT_ROOT/bin"
LOG_DIR="$PROJECT_ROOT/logs"
CONFIG_DIR="$PROJECT_ROOT/../configs/dev"

SKIP_BUILD=false
SKIP_INFRA=false

for arg in "$@"; do
  case "$arg" in
    -no-build) SKIP_BUILD=true ;;
    -no-infra)  SKIP_INFRA=true  ;;
  esac
done

# ─── 颜色 ───
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

echo ""
echo -e "${CYAN}============================================${NC}"
echo -e "${CYAN}   GoForge E-Commerce 一键启动${NC}"
echo -e "${CYAN}============================================${NC}"
echo ""

# ─── Step 1: Docker ───
if [ "$SKIP_INFRA" = true ]; then
  echo -e "${YELLOW}[1/4] 跳过 Docker 基础设施${NC}"
else
  echo -e "${YELLOW}[1/4] 启动 Docker 基础设施...${NC}"
  cd "$PROJECT_ROOT"
  docker compose -f docker-compose-infra.yml up -d 2>/dev/null
  echo -e "${GREEN}  Docker 基础设施已就绪${NC}"
fi

# ─── Step 2: Build ───
if [ "$SKIP_BUILD" = true ]; then
  echo -e "${YELLOW}[2/4] 跳过编译${NC}"
else
  echo -e "${YELLOW}[2/4] 编译所有服务...${NC}"
  mkdir -p "$BIN_DIR"
  cd "$PROJECT_ROOT"

  SERVICES=(
    "user-service" "product-service" "order-service" "payment-service"
    "inventory-service" "cart-service" "promotion-service" "review-service"
    "logistics-service" "message-service" "search-service" "recommend-service"
    "file-service" "job-service" "seckill-service" "api-gateway"
  )

  for svc in "${SERVICES[@]}"; do
    printf "  Building %-20s ... " "$svc"
    if go build -o "$BIN_DIR/$svc.exe" "./cmd/$svc" 2>/dev/null; then
      echo -e "${GREEN}OK${NC}"
    else
      echo -e "${RED}FAILED${NC}"
      exit 1
    fi
  done
  echo -e "${GREEN}  全部编译完成${NC}"
fi

# ─── Step 3: Start Microservices ───
echo ""
echo -e "${YELLOW}[3/4] 启动微服务...${NC}"
cd "$PROJECT_ROOT"
mkdir -p "$LOG_DIR"

# 停止旧进程
powershell -Command "Get-Process -Name 'user-service','product-service','order-service','payment-service','inventory-service','cart-service','promotion-service','review-service','logistics-service','message-service','search-service','recommend-service','file-service','job-service','seckill-service','api-gateway' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue" 2>/dev/null
sleep 2

PORTS=(
  "8000:user" "8081:product" "8082:order" "8083:payment"
  "8084:inventory" "8085:cart" "8006:promotion" "8007:review"
  "8008:logistics" "8009:message" "8010:search" "8011:recommend"
  "8012:file" "8013:job" "8090:seckill"
)

started=0
failed=0

for entry in "${PORTS[@]}"; do
  port="${entry%%:*}"
  name="${entry##*:}"
  exe="$BIN_DIR/$name-service.exe"
  cfg="$CONFIG_DIR/$name-config.yaml"

  if [ -f "$exe" ]; then
    nohup "$exe" -f "$cfg" > "$LOG_DIR/$name-service.log" 2>&1 &
  else
    nohup go run "cmd/$name-service/main.go" -f "$cfg" > "$LOG_DIR/$name-service.log" 2>&1 &
  fi

  # 等待端口监听
  for i in $(seq 1 20); do
    if netstat -ano 2>/dev/null | grep -q ":$port .*LISTENING"; then
      echo -e "  ${GREEN}[+]${NC} $name-service :$port"
      ((started++))
      break
    fi
    sleep 1
  done

  if ! netstat -ano 2>/dev/null | grep -q ":$port .*LISTENING"; then
    echo -e "  ${RED}[-]${NC} $name-service :$port TIMEOUT"
    ((failed++))
  fi
done

echo ""
echo -e "  微服务: ${GREEN}$started${NC} 启动, ${RED}$failed${NC} 失败"

# ─── Step 4: Gateway ───
echo ""
echo -e "${YELLOW}[4/4] 启动 API Gateway...${NC}"

if [ $failed -gt 0 ]; then
  echo -e "  ${RED}部分上游服务未启动，跳过 Gateway${NC}"
else
  cd "$PROJECT_ROOT"
  nohup "$BIN_DIR/api-gateway.exe" -f "$CONFIG_DIR/gateway.yaml" > "$LOG_DIR/api-gateway.log" 2>&1 &
  sleep 6
  if netstat -ano 2>/dev/null | grep -q ":8080 .*LISTENING"; then
    echo -e "  ${GREEN}[+]${NC} API Gateway :8080"
    echo -e "  ${GREEN}[+]${NC} Swagger UI  :8095"
  else
    echo -e "  ${RED}[-]${NC} Gateway 启动失败"
  fi
fi

# ─── Summary ───
echo ""
echo -e "${CYAN}============================================${NC}"
echo -e "${CYAN}   启动完成${NC}"
echo -e "${CYAN}============================================${NC}"
echo ""
echo -e "  API 入口:  ${GREEN}http://localhost:8080/api/v1${NC}"
echo -e "  Swagger:   ${GREEN}http://localhost:8095${NC}"
echo -e "  日志目录:  ${YELLOW}$LOG_DIR${NC}"
echo ""
