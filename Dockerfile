# syntax=docker/dockerfile:1
# ==================================
# 1. ビルドステージ
# ==================================
FROM golang:1.25-trixie AS builder

# ビルド依存関係
# Opusエンコーダー（USE_PION_OPUS=true）用: libopus-dev, libopusfile-dev（CGOビルド時）
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    gcc \
    g++ \
    pkg-config \
    libopus-dev \
    libopusfile-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

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
FROM debian:trixie-slim

# システム依存関係
# DCAエンコーダー（デフォルト）用: ffmpeg
# Opusエンコーダー（USE_PION_OPUS=true）用: libopus0/libopusfile0（ランタイム）
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    opus-tools \
    ca-certificates \
    libopus0 \
    libopusfile0 \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# ビルド済みバイナリをコピー
COPY --from=builder /app/yoursay-bot .

# 設定ファイルをコピー
COPY config/config.yaml ./config/

CMD ["./yoursay-bot"]

# ==================================
# 3. 開発ステージ
# ==================================
FROM golang:1.25-trixie AS development

# ビルド依存関係と開発ツール
# Opusエンコーダー（USE_PION_OPUS=true）用: libopus-dev, libopusfile-dev（CGOビルド時）
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    gcc \
    g++ \
    pkg-config \
    ffmpeg \
    opus-tools \
    ca-certificates \
    libopus-dev \
    libopusfile-dev \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# アプリケーションコードをコピー
COPY . .

# 開発用のエントリーポイント（必要に応じて変更可能）
CMD ["go", "run", "./cmd/bot"]

