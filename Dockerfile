# ============================================
# Multi-stage Dockerfile for elotus_test
# ============================================

# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Install git (required for go mod download)
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./server

# Stage 2: Run
FROM alpine:latest

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy required directories
COPY --from=builder /app/server/db ./server/db
COPY --from=builder /app/server/html ./server/html

# Create env file for Docker
RUN echo 'environment: "production"' > /app/server/.env.docker.yaml && \
    echo 'server_name: "elotus-auth"' >> /app/server/.env.docker.yaml && \
    echo 'database_config_file_path: "server/db/database.docker.yaml"' >> /app/server/.env.docker.yaml && \
    echo 'redis_config_file_path: "server/db/redis.docker.yaml"' >> /app/server/.env.docker.yaml && \
    echo 'jwt_signing_key: "your-super-secret-jwt-key-change-in-production-min-32-chars"' >> /app/server/.env.docker.yaml && \
    echo 'jwt_token_duration: "24h"' >> /app/server/.env.docker.yaml && \
    echo 'token_revoke_duration: "24h"' >> /app/server/.env.docker.yaml && \
    echo 'backend:' >> /app/server/.env.docker.yaml && \
    echo '  host_http: "0.0.0.0"' >> /app/server/.env.docker.yaml && \
    echo '  port: "8080"' >> /app/server/.env.docker.yaml && \
    echo 'time_zone_offset: 7' >> /app/server/.env.docker.yaml && \
    echo 'time_zone_name: "Asia/Ho_Chi_Minh"' >> /app/server/.env.docker.yaml && \
    echo 'features:' >> /app/server/.env.docker.yaml && \
    echo '  enable_registration: true' >> /app/server/.env.docker.yaml && \
    echo '  enable_token_revoke: true' >> /app/server/.env.docker.yaml

# Create tmp directory for uploads
RUN mkdir -p /app/tmp/images

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application with docker env mode
CMD ["./main", "-envMode", "docker"]

