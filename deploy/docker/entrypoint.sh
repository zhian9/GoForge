#!/bin/sh
set -e

SERVICE="${1:-api-gateway}"
CONFIG_DIR="${CONFIG_DIR:-/app/configs}"

case "$SERVICE" in
  api-gateway)
    exec /app/bin/api-gateway -f "${CONFIG_DIR}/gateway.yaml"
    ;;
  order-service-consumer)
    exec /app/bin/order-service-consumer -f "${CONFIG_DIR}/order-config.yaml"
    ;;
  user|product|order|payment|inventory|cart|promotion|review|logistics|message|search|recommend|file|job|seckill)
    exec "/app/bin/${SERVICE}-service" -f "${CONFIG_DIR}/${SERVICE}-config.yaml"
    ;;
  *)
    echo "Unknown service: $SERVICE"
    exit 1
    ;;
esac
