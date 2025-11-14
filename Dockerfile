# syntax=docker/dockerfile:1
# ==================================
# 1. ビルドステージ
# ==================================
FROM golang:1.21-alpine AS builder

# ビルド依存関係
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# go.modとgo.sumをコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# アプリケーションコードをコピー
COPY . .

# ビルド
RUN CGO_ENABLED=1 go build -o yoursay-bot ./cmd/bot

# ==================================
# 2. ランタイムステージ
# ==================================
FROM alpine:latest

# システム依存関係
RUN apk add --no-cache ffmpeg opus-tools ca-certificates

WORKDIR /app

# ビルド済みバイナリをコピー
COPY --from=builder /app/yoursay-bot .

# 設定ファイルをコピー
COPY config/config.yaml ./config/

CMD ["./yoursay-bot"]

