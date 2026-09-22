# ============================================
# GoForge - 后端服务镜像
# 编译所有 16 个 Go 微服务到同一镜像
# ============================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git tzdata

WORKDIR /build
COPY server/go.mod server/go.sum ./
RUN go mod download && go mod verify

COPY server/. .

RUN set -e; \
    mkdir -p bin; \
    for svc in user product order payment inventory cart promotion review logistics message search recommend file job seckill; do \
      go build -ldflags="-s -w" -o "bin/${svc}-service" "./cmd/${svc}-service"; \
    done; \
    go build -ldflags="-s -w" -o bin/api-gateway ./cmd/api-gateway; \
    go build -ldflags="-s -w" -o bin/order-service-consumer ./cmd/order-service-consumer

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /build/bin/ /app/bin/
COPY configs/docker/ /app/configs/
# Swagger 文档由 cmd/generate-swagger 生成（见 server/docs/swagger），
# 网关内嵌的 Swagger UI 会从 /app/docs/swagger 读取，因此必须一起进镜像
COPY server/docs/ /app/docs/
COPY deploy/docker/entrypoint.sh /app/

RUN chmod +x /app/entrypoint.sh && mkdir -p /app/logs /app/uploads /app/images

EXPOSE 8000 8080-8085 8090 8095 8006-8013

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["api-gateway"]
