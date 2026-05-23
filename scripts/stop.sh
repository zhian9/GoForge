#!/bin/bash
# ============================================
# GoForge E-Commerce 一键停止脚本
# ============================================

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}停止所有服务...${NC}"

# 停止微服务
powershell -Command "
  \$names = @('user-service','product-service','order-service','payment-service','inventory-service','cart-service','promotion-service','review-service','logistics-service','message-service','search-service','recommend-service','file-service','job-service','seckill-service','api-gateway');
  \$killed = 0;
  foreach (\$n in \$names) {
    \$p = Get-Process -Name \$n -ErrorAction SilentlyContinue;
    if (\$p) { \$p | Stop-Process -Force; \$killed += (\$p | Measure-Object).Count }
  }
  Write-Host \"  Stopped \$killed processes\"
" 2>/dev/null

echo -e "${GREEN}全部服务已停止${NC}"
