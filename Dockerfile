# syntax=docker/dockerfile:1
# ==================================
# 1. ビルドステージ
# ==================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Go modulesのダウンロードをキャッシュ化
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# ビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# ==================================
# 2. ランタイムステージ
# ==================================
FROM alpine:latest

# 必要なライブラリをインストール（音声処理用）
RUN apk --no-cache add ca-certificates ffmpeg opus-tools

WORKDIR /root/

# バイナリをコピー
COPY --from=builder /app/bot .

# 設定ファイルをコピー
COPY --from=builder /app/config.sample.yml ./config.yml

# ポート開放（必要に応じて）
EXPOSE 8080

# 実行
CMD ["./bot"]