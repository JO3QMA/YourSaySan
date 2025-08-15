# ==================================
# 1. ベースステージ (全環境で共通)
# ==================================
FROM ruby:3.4.2 AS base

WORKDIR /app

# Bundler のインストール先を固定（キャッシュ安定化）
RUN bundle config set path vendor/bundle

# Optional audio tooling & runtime libs for Discord voice
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg opus-tools libopus0 libsodium23 && rm -rf /var/lib/apt/lists/*

# ==================================
# 2. 開発環境用ステージ
# ==================================
FROM base AS development

# Gemfileをコピーして、すべてのGemをインストール
COPY Gemfile Gemfile.lock ./
RUN bundle install

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
RUN bundle install --without development test

# アプリケーションコードをコピー
COPY . .

ENV RACK_ENV=production
CMD ["ruby", "run.rb"]