# syntax=docker/dockerfile:1
# ==================================
# 1. ベースステージ (全環境で共通)
# ==================================
FROM ruby:3.4.2 AS base

WORKDIR /app

# Bundler のインストール先を固定（キャッシュ安定化）
RUN bundle config set path vendor/bundle

# Optional audio tooling & runtime libs for Discord voice
# 新しいキャッシュ機能を使用してaptパッケージのキャッシュを効率的に管理
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg opus-tools libopus0 libopus-dev libsodium23

# ==================================
# 2. 開発環境用ステージ
# ==================================
FROM base AS development

# Gemfileをコピーして、すべてのGemをインストール
COPY Gemfile Gemfile.lock ./
# Bundlerのキャッシュを効率的に管理（vendor/bundleにキャッシュを設定）
RUN --mount=type=cache,target=/app/vendor/bundle \
    bundle install

# アプリケーションコードをコピー
COPY . .

# Dev Container起動時のデフォルトコマンド (特に不要ならなくてもOK)
CMD ["sleep", "infinity"]

# ==================================
# 3. 本番環境用ステージ
# ==================================
FROM base AS production

# Gemfileをコピーして、本番用のGemのみインストール
COPY Gemfile Gemfile.lock ./
# Bundlerのキャッシュを効率的に管理（本番用、vendor/bundleにキャッシュを設定）
RUN --mount=type=cache,target=/app/vendor/bundle \
    bundle install --without development test

# アプリケーションコードをコピー
COPY . .

ENV RACK_ENV=production
CMD ["ruby", "run.rb"]