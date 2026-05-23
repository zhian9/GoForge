FROM node:22-alpine AS builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/. .
RUN npm run build

FROM alpine:3.21
COPY --from=builder /app/dist /dist

