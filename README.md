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

1.  `config.yml`を作成し、必要な情報を設定します。`config.yml.sample`を参考にしてください。
2.  Docker Composeを使用して、BotとVoicevox Engineを起動します。

    ```bash
    docker-compose up --build
    ```

3.  DiscordでBotを招待し、`!ping`コマンドを試してみてください。`Pong!`と返信されれば、Botは正常に動作しています。

## GitHub Actions による自動ビルド

このプロジェクトでは、GitHub Actionsを使用してDockerイメージの自動ビルドを行っています。

### 動作条件

- `main`または`master`ブランチへのプッシュ
- タグ付きリリース（`v*`形式）
- プルリクエスト

### ビルドされるイメージ

- **本番用イメージ**: `ghcr.io/{username}/{repository}:{tag}`
- **開発用イメージ**: `ghcr.io/{username}/{repository}:dev`

### 使用方法

1. GitHub Container Registryからイメージを取得:
   ```bash
   docker pull ghcr.io/{username}/{repository}:latest
   ```

2. ローカルで実行:
   ```bash
   docker run -d ghcr.io/{username}/{repository}:latest
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
2. `.env.sample` を `.env` にコピーして、必要な値を設定します。

   ```bash
   cp .env.sample .env
   # エディタで .env を開き、各値を入力
   ```

   - `DISCORD_BOT_TOKEN`（必須）
   - `DISCORD_CLIENT_ID`（必須）
   - `DISCORD_OWNER_ID`（任意）
   - `VOICEVOX_HOST`（任意。既定は `http://voicevox:50021`）

3. コンテナ内ターミナルで Bot を起動します。

   ```bash
   ruby run.rb
   ```

メモ:

- DevContainer は `.devcontainer/compose.devcontainer.yml` のみで起動します。`voicevox` サービスも自動で立ち上がります。
- ポート `50021` はホストにフォワードされます。

## 設定ファイル (config.yml)

```yaml
bot:
  token: 'YOUR_BOT_TOKEN'       # Discord BotのToken
  client_id: 'YOUR_CLIENT_ID'   # Discord BotのClient ID
  prefix: '!'                   # コマンドのプレフィックス
  status: '[TESTING] 読み上げBot' # Botのステータス
  auto_reconnect:
    enabled: true               # 自動再接続機能を有効にする
    heartbeat_timeout: 120      # ハートビートが途絶えてから再接続するまでの秒数
    check_interval: 60          # 接続状態チェックの間隔（秒）
    retry_delay: 30             # 再接続失敗時の再試行間隔（秒）

voicevox:
  max: 50                      # 最大文字数
  host: 'http://voicevox:50021' # Voicevox Engineのホスト

redis:
  host: 'redis'                # Redisのホスト
  port: 6379                   # Redisのポート
  db: 0                        # Redisのデータベース番号
```

## 機能

### 自動再接続機能

Cloudflare WorkersのアップデートなどによるWebSocket接続断を自動的に検知し、再接続を行います。

- **接続断検知**: WebSocket接続が切断された際に自動的に検知
- **ハートビート監視**: Discordからのハートビートを監視し、途絶えた場合に再接続
- **定期的な接続状態チェック**: 1分ごとに接続状態をチェック
- **指数バックオフ**: 再接続失敗時は徐々に待機時間を延長

### 話者設定機能

ユーザーごとにVoiceVoxの話者を設定できます。

- **個人設定**: 各ユーザーが自分の好みの話者を設定可能
- **永続化**: Redisを使用して話者設定を永続化
- **話者一覧**: `/speaker_list`コマンドで利用可能な話者を確認
- **設定変更**: `/speaker`コマンドで話者IDを指定して設定変更

### コマンド

*   `/ping`: Botの死活確認を行います。
*   `/summon`: 読み上げBotをVCに参加させます。
*   `/bye`: 読み上げBotをVCから退出させます。
*   `/reconnect`: Discordが調子悪いときなどに、手動で再接続します。
*   `/stop`: 読み上げを中断します。
*   `/speaker`: 話者を設定します（例: `/speaker 2`）。
*   `/speaker_list`: 利用可能な話者の一覧を表示します。

