# Discord 音声読み上げBot

## 概要

このプロジェクトは、Discord上で音声読み上げを行うBotです。<br>
Voicevox Engineを使用して、テキストを音声に変換し、Discordのボイスチャットでユーザーの代わりに喋ります。

## 環境

*   Ruby 3.4.2
*   Docker
*   Docker Compose
*   GitHub Actions (Dockerイメージの自動ビルド)

## 必要なもの

*   Discord BotのToken
*   Discord BotのClient ID
*   Voicevox Engine (Dockerコンテナで起動)
*   Redis (話者設定の永続化用)

## 使い方

1.  `config.yml`を作成し、必要な情報を設定します。`config.sample.yml`を参考にしてください。
2.  Docker Composeを使用して、BotとVoicevox Engineを起動します。

    ```bash
    docker-compose up --build
    ```

3.  DiscordでBotを招待し、`/ping`コマンドを試してみてください。`Pong!`と返信されれば、Botは正常に動作しています。

## GitHub Actions による自動ビルド

このプロジェクトでは、GitHub Actionsを使用してDockerイメージの自動ビルドを行っています。

### 動作条件

- `main`または`master`ブランチへのプッシュ
- タグ付きリリース（`v*`形式）
- プルリクエスト

### ビルドされるイメージ

- **mainブランチ**: `ghcr.io/jo3qma/yoursaysan:main`
- **開発用イメージ**: `ghcr.io/jo3qma/yoursaysan:dev`
- **セマンティックバージョン**: `ghcr.io/jo3qma/yoursaysan:v1.0.0`（タグ付きリリース時）
- **メジャー・マイナーバージョン**: `ghcr.io/jo3qma/yoursaysan:1.0`（タグ付きリリース時）
- **SHAハッシュ**: `ghcr.io/jo3qma/yoursaysan:sha-{hash}`（コミット時）

### 使用方法

1. GitHub Container Registryからイメージを取得:
   ```bash
   # 最新のmainブランチのイメージ
   docker pull ghcr.io/jo3qma/yoursaysan:main
   
   # 開発用イメージ
   docker pull ghcr.io/jo3qma/yoursaysan:dev
   
   # 特定のバージョン（例: v1.0.0）
   docker pull ghcr.io/jo3qma/yoursaysan:v1.0.0
   ```

2. ローカルで実行:
   ```bash
   # mainブランチのイメージで実行
   docker run -d ghcr.io/jo3qma/yoursaysan:main
   
   # 開発用イメージで実行
   docker run -d ghcr.io/jo3qma/yoursaysan:dev
   ```

### 注意事項

- GitHub Container Registryへのプッシュには、リポジトリの設定で「Packages」の権限を有効にする必要があります
- プルリクエスト時はイメージのビルドのみ実行され、プッシュは行われません

## DevContainer での開発

### 前提

- Docker / Docker Compose
- VS Code もしくは Cursor の Dev Containers 機能（拡張）

### 手順

1. エディタでこのリポジトリを開き、「Reopen in Container」を実行します。
   - 初回起動時に `bundle install` が自動実行され、`config.yml` が無ければ `config.sample.yml` から生成されます。
2. 環境変数を設定します。以下の環境変数を設定してください：

   - `DISCORD_BOT_TOKEN`（必須）
   - `DISCORD_CLIENT_ID`（必須）
   - `DISCORD_OWNER_ID`（任意）
   - `VOICEVOX_HOST`（任意。既定は `http://voicevox:50021`）
   - `REDIS_HOST`（任意。既定は `redis`）
   - `REDIS_PORT`（任意。既定は `6379`）
   - `REDIS_DB`（任意。既定は `0`）

3. コンテナ内ターミナルで Bot を起動します。

   ```bash
   ruby run.rb
   ```

メモ:

- DevContainer は `.devcontainer/compose.devcontainer.yml` のみで起動します。`voicevox` サービスも自動で立ち上がります。
- 各サービスは内部ネットワークで通信するため、ホストへのポート公開は不要です。

## 設定ファイル (config.yml)

```yaml
bot:
  token: <%= ENV['DISCORD_BOT_TOKEN'] %>       # Discord BotのToken
  client_id: <%= ENV['DISCORD_CLIENT_ID'] %>   # Discord BotのClient ID
  prefix: '!'                   # コマンドのプレフィックス
  status: '[TESTING] 読み上げBot' # Botのステータス
  owner: <%= ENV['DISCORD_OWNER_ID'] || 123456789012345678 %> # BotのオーナーID

voicevox:
  max: 50                      # 最大文字数
  host: <%= ENV['VOICEVOX_HOST'] || 'http://voicevox:50021' %> # Voicevox Engineのホスト

redis:
  host: <%= ENV['REDIS_HOST'] || 'redis' %>    # Redisのホスト
  port: <%= ENV['REDIS_PORT'] || 6379 %>       # Redisのポート
  db: <%= ENV['REDIS_DB'] || 0 %>              # Redisのデータベース番号
```

## 機能

### 話者設定機能

ユーザーごとにVoiceVoxの話者を設定できます。

- **個人設定**: 各ユーザーが自分の好みの話者を設定可能
- **永続化**: Redisを使用して話者設定を永続化
- **話者一覧**: `/speaker_list`コマンドで利用可能な話者を確認
- **設定変更**: `/speaker`コマンドで話者IDを指定して設定変更

### コマンド

*   `/ping`: Botの死活確認を行います。
*   `/help`: 利用可能なコマンドの一覧を表示します。
*   `/invite`: Botを他のサーバーに招待するためのURLを表示します。
*   `/summon`: 読み上げBotをVCに参加させます。
*   `/bye`: 読み上げBotをVCから退出させます。
*   `/reconnect`: Discordが調子悪いときなどに、手動で再接続します。
*   `/stop`: 読み上げを中断します。
*   `/speaker`: 話者を設定します（例: `/speaker 2`）。
*   `/speaker_list`: 利用可能な話者の一覧を表示します。

