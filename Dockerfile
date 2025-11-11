# syntax=docker/dockerfile:1
# ==================================
# 1. ビルドステージ
# ==================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /bot ./cmd/bot

# ==================================
# 2. 開発環境用ステージ
# ==================================
FROM golang:1.23-alpine AS development

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ffmpeg opus-tools libopus ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install air for hot reloading (optional)
RUN go install github.com/air-verse/air@latest

# Dev Container起動時のデフォルトコマンド
CMD ["sleep", "infinity"]

# ==================================
# 3. 本番環境用ステージ
# ==================================
FROM alpine:latest AS production

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ffmpeg opus-tools libopus ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /bot /app/bot

# Copy configuration file
COPY config.sample.yml /app/config.sample.yml

# Create a non-root user
RUN addgroup -g 1000 bot && \
    adduser -D -u 1000 -G bot bot && \
    chown -R bot:bot /app

USER bot

ENV TZ=Asia/Tokyo

CMD ["/app/bot", "-config", "/app/config.sample.yml"]
