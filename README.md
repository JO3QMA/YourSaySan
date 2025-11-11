# YourSaySan - Discord VoiceVox Bot

Discord音声読み上げBot（VoiceVox Engine連携、Redis永続化）のGolang実装版です。

## 概要

このBotは、Discordのテキストチャンネルに投稿されたメッセージを、VoiceVox Engineを使用して音声に変換し、ボイスチャンネルで読み上げます。

## アーキテクチャ

このプロジェクトは**クリーンアーキテクチャ**に基づいて設計されています：

```
.
├── cmd/bot/                    # エントリーポイント
├── internal/
│   ├── domain/                 # ドメイン層
│   │   ├── entity/            # エンティティ
│   │   └── repository/        # リポジトリインターフェース
│   ├── usecase/               # ユースケース層
│   │   ├── speaker/           # 話者管理
│   │   └── voice/             # 音声生成
│   ├── interface/             # インターフェース層
│   │   ├── discord/           # Discord Bot実装
│   │   ├── voicevox/          # VoiceVox APIクライアント
│   │   └── redis/             # Redis実装
│   └── infrastructure/        # インフラ層
│       ├── config/            # 設定管理
│       └── logger/            # ロギング
└── pkg/                       # 共通パッケージ
```

## 機能

- ✅ テキストメッセージの音声読み上げ
- ✅ 話者の選択・変更
- ✅ ユーザーごとの話者設定保存（Redis）
- ✅ スラッシュコマンド対応
- ✅ 自動切断（無人時）
- ✅ VoiceVox Engine統合

## スラッシュコマンド

| コマンド | 説明 |
|---------|------|
| `/ping` | Botの応答速度を確認 |
| `/summon` | Botをボイスチャンネルに召喚 |
| `/bye` | Botをボイスチャンネルから退出 |
| `/stop` | 現在再生中の音声を停止 |
| `/speaker <id>` | 話者を設定 |
| `/speaker_list` | 利用可能な話者の一覧を表示 |
| `/reconnect` | ボイスチャンネルに再接続 |
| `/invite` | Botの招待リンクを表示 |
| `/help` | ヘルプを表示 |

## セットアップ

### 必要要件

- Go 1.23以上
- Docker & Docker Compose
- Discord Bot Token
- ffmpeg（音声処理用）

### 環境変数

`.env`ファイルを作成し、以下の環境変数を設定してください：

```env
DISCORD_BOT_TOKEN=your_discord_bot_token
DISCORD_CLIENT_ID=your_discord_client_id
DISCORD_OWNER_ID=your_discord_user_id
VOICEVOX_HOST=http://voicevox:50021
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_DB=0
```

### ローカル開発

```bash
# 依存関係のインストール
go mod download

# ビルド
go build -o bot ./cmd/bot

# 実行
./bot -config config.sample.yml
```

### Docker Compose

```bash
# ビルドと起動
docker compose up -d

# ログ確認
docker compose logs -f yousaysan

# 停止
docker compose down
```

## 技術スタック

- **言語**: Go 1.23
- **Discord**: [discordgo](https://github.com/bwmarrin/discordgo)
- **Redis**: [go-redis](https://github.com/redis/go-redis)
- **設定**: YAML (gopkg.in/yaml.v3)
- **音声エンジン**: [VoiceVox Engine](https://github.com/VOICEVOX/voicevox_engine)

## 設計原則

このプロジェクトは以下の原則に従っています：

- **DRY (Don't Repeat Yourself)**: コードの重複を避ける
- **SOLID原則**: 
  - 単一責任の原則
  - オープン/クローズドの原則
  - リスコフの置換原則
  - インターフェース分離の原則
  - 依存性逆転の原則
- **クリーンアーキテクチャ**: レイヤー間の依存関係を適切に管理

## ライセンス

このプロジェクトのライセンスは未定です。

## クレジット

- [VoiceVox](https://voicevox.hiroshiba.jp/) - 音声合成エンジン
- [DiscordGo](https://github.com/bwmarrin/discordgo) - Discord Go ライブラリ
