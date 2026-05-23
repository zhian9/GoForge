#!/bin/bash

# ============================================
# 数据库初始化脚本
# ============================================

set -e

# 配置变量
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASS="${DB_PASS:-123456}"
DB_NAME="${DB_NAME:-go_forge}"

echo "============================================"
echo "Go 微服务电商项目 - 数据库初始化"
echo "============================================"
echo "数据库主机: $DB_HOST"
echo "数据库端口: $DB_PORT"
echo "数据库用户: $DB_USER"
echo "数据库名称: $DB_NAME"
echo "============================================"

# 检查 MySQL 客户端是否安装
if ! command -v mysql &> /dev/null; then
    echo "错误: 未找到 mysql 客户端，请先安装 MySQL 客户端"
    exit 1
fi

# 构建 MySQL 连接命令
MYSQL_CMD="mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER}"
if [ -n "$DB_PASS" ]; then
    MYSQL_CMD="${MYSQL_CMD} -p${DB_PASS}"
fi

# 创建数据库
echo "正在创建数据库..."
$MYSQL_CMD -e "CREATE DATABASE IF NOT EXISTS ${DB_NAME} DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
echo "✓ 数据库创建成功"

# 导入表结构
echo "正在导入表结构..."
$MYSQL_CMD ${DB_NAME} < "$(dirname "$0")/../server/database/schema.sql"
echo "✓ 表结构导入成功"

# 询问是否创建分表
read -p "是否创建分表？(y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "正在创建分表..."
    $MYSQL_CMD ${DB_NAME} < "$(dirname "$0")/../server/database/sharding.sql"
    echo "✓ 分表创建成功"
fi

echo "============================================"
echo "数据库初始化完成！"
echo "============================================"
