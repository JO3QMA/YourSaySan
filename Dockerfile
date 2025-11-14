# syntax=docker/dockerfile:1
# ==================================
# 1. ビルドステージ
# ==================================
FROM golang:1.25-alpine AS builder

# ビルド依存関係
# Opusエンコーダー（USE_PION_OPUS=true）用: opus-dev, opusfile-dev（CGOビルド時）
RUN apk add --no-cache git make gcc musl-dev opus-dev opusfile-dev

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
# DCAエンコーダー（デフォルト）用: ffmpeg
# Opusエンコーダー（USE_PION_OPUS=true）用: opus-dev（CGOビルド時）
RUN apk add --no-cache ffmpeg opus-tools ca-certificates opus-dev

WORKDIR /app

# ビルド済みバイナリをコピー
COPY --from=builder /app/yoursay-bot .

# 設定ファイルをコピー
COPY config/config.yaml ./config/

CMD ["./yoursay-bot"]

# ==================================
# 3. 開発ステージ
# ==================================
FROM golang:1.25-alpine AS development

# ビルド依存関係と開発ツール
# Opusエンコーダー（USE_PION_OPUS=true）用: opus-dev, opusfile-dev（CGOビルド時）
RUN apk add --no-cache git make gcc musl-dev ffmpeg opus-tools ca-certificates opus-dev opusfile-dev

WORKDIR /app

# 依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# アプリケーションコードをコピー
COPY . .

# 開発用のエントリーポイント（必要に応じて変更可能）
CMD ["go", "run", "./cmd/bot"]

