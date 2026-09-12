FROM node:22-alpine AS builder
WORKDIR /app
COPY admin/package*.json ./
RUN npm ci
COPY admin/. .
RUN npx vite build

FROM alpine:3.21
COPY --from=builder /app/dist /dist
