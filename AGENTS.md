# AGENTS.md — YourSaySan コーディングエージェント向けガイド

このファイルは AI コーディングエージェントがこのリポジトリを理解・操作するための参照ドキュメントです。

---

## プロジェクト概要

**YourSaySan** は Discord のテキストチャンネルに投稿されたメッセージを VoiceVox Engine で音声合成し、ボイスチャンネルで読み上げる Bot です。

- 言語: **Go 1.25**
- Discord ライブラリ: `github.com/bwmarrin/discordgo`
- TTS エンジン: [VoiceVox Engine](https://voicevox.hiroshiba.jp/)（HTTP API）
- ストレージ: Redis（ユーザーごとの話者設定を永続化）
- ロギング: Logrus（構造化ログ）
- 設定: 環境変数（`godotenv` + `os.Getenv`）

---

## ディレクトリ構成

```
/app
├── cmd/bot/main.go              # エントリーポイント（シグナル処理・グレースフルシャットダウン）
├── internal/
│   ├── bot/                     # Bot コア
│   │   ├── bot.go               # Bot 構造体・ライフサイクル管理
│   │   ├── config.go            # 設定ロード（環境変数）
│   │   ├── state.go             # Bot 状態管理
│   │   └── http.go              # ヘルスチェック・メトリクス HTTP サーバー（:8080）
│   ├── commands/                # スラッシュコマンドハンドラ
│   │   ├── registry.go          # コマンド登録・ディスパッチ
│   │   ├── bot_interface.go     # commands パッケージが依存する Bot インターフェース
│   │   └── *.go                 # 各コマンド実装
│   ├── events/                  # Discord イベントハンドラ
│   │   ├── bot_interface.go     # events パッケージが依存する Bot インターフェース
│   │   ├── message_create.go    # メッセージ受信 → TTS キューイング
│   │   ├── voice_state_update.go# VC 入退出検知（自動切断）
│   │   ├── ready.go             # Discord Ready イベント
│   │   └── disconnect.go        # Discord 切断イベント
│   ├── voice/                   # VC 接続・音声再生
│   │   ├── connection.go        # VC 接続管理
│   │   ├── player.go            # 音声キュー再生ループ
│   │   ├── queue.go             # スレッドセーフなオーディオキュー
│   │   ├── encoder.go           # エンコーダーインターフェース
│   │   ├── encoder_dca.go       # DCA エンコーダー（デフォルト）
│   │   └── encoder_opus.go      # Pion Opus エンコーダー（USE_PION_OPUS=true）
│   ├── voicevox/                # VoiceVox API クライアント
│   │   ├── client.go            # HTTP クライアント（リトライ・指数バックオフ付き）
│   │   └── types.go             # API レスポンス型定義
│   ├── speaker/                 # 話者設定管理
│   │   ├── manager.go           # Redis + LRU キャッシュによる話者設定管理
│   │   └── interface.go         # SpeakerManager インターフェース
│   └── errors/errors.go         # ドメインエラー定義
├── pkg/utils/                   # 汎用ユーティリティ
│   ├── logger.go                # Logrus 初期化
│   └── message.go               # メッセージ前処理
├── config/config.yaml           # 設定ファイル（環境変数展開: ${VAR:-default}）
├── compose.yaml                  # 本番 Docker Compose（bot + voicevox + redis）
├── Dockerfile                   # マルチステージビルド（builder / runtime / development）
└── .devcontainer/               # DevContainer 設定
    ├── devcontainer.json
    └── compose.devcontainer.yml
```

---

## アーキテクチャ概要

### 依存関係の方向

```
cmd/bot → internal/bot → internal/commands, events, voice, voicevox, speaker
```

`commands` と `events` は `internal/bot` に直接依存せず、**インターフェース**（`bot_interface.go`）を通じて Bot に依存します。これにより循環参照を防ぎ、テスト容易性を確保しています。

### 音声再生パイプライン

```
Discord メッセージ受信
  → message_create.go でテキスト前処理
  → VoiceVox API（/audio_query + /synthesis）で WAV 生成
  → DCA または Pion Opus エンコーダーで Opus に変換
  → voice.Queue に積む
  → voice.Player が goroutine でキューを順次再生
  → discordgo の VoiceConnection.OpusSend に送信
```

### マルチギルド対応

`Bot.voiceConns` マップ（`map[string]*voice.Connection`）でギルドごとに VC 接続を管理。`sync.RWMutex` で保護。

---

## ビルドと実行

### ローカルビルド

```bash
# 依存関係ダウンロード
go mod download

# ビルド（CGO 有効、Opus ライブラリが必要）
CGO_ENABLED=1 go build -o yoursay-bot ./cmd/bot

# 実行
./yoursay-bot -config config/config.yaml
```

### Docker Compose で起動（本番）

```bash
# .env ファイルを作成
cp .env.example .env  # なければ手動で作成

# 起動
docker compose up --build
```

### 開発用（コードの変更を反映しながら実行）

```bash
go run ./cmd/bot
```

---

## 設定（環境変数）

設定は環境変数から読み込まれます。`.env`ファイルが存在する場合は、自動的に読み込まれます。

**`Config`構造体（`internal/bot/config.go`）:**

```go
type Config struct {
	Bot struct {
		Token    string // DISCORD_BOT_TOKEN
		ClientID string // DISCORD_CLIENT_ID
		Status   string // DISCORD_BOT_STATUS
		OwnerID  string // DISCORD_OWNER_ID
	}
	VoiceVox struct {
		MaxChars         int    // VOICEVOX_MAX_CHARS
		MaxMessageLength int    // VOICEVOX_MAX_MESSAGE_LENGTH
		Host             string // VOICEVOX_HOST
	}
	Redis struct {
		Host string // REDIS_HOST
		Port int    // REDIS_PORT
		DB   int    // REDIS_DB
	}
}
```

**必須環境変数:**
- `DISCORD_BOT_TOKEN` — Discord Bot トークン
- `DISCORD_CLIENT_ID` — Discord Bot クライアント ID

**任意環境変数:**
- `DISCORD_OWNER_ID` — Bot オーナーの Discord ユーザー ID（デフォルト: `123456789012345678`）
- `DISCORD_BOT_STATUS` — Bot のステータス（デフォルト: `[TESTING] 読み上げBot`）
- `VOICEVOX_HOST` — VoiceVox Engine のホスト URL（デフォルト: `http://voicevox:50021`）
- `VOICEVOX_MAX_CHARS` — 1回の読み上げ最大文字数（デフォルト: `200`）
- `VOICEVOX_MAX_MESSAGE_LENGTH` — メッセージの最大長（デフォルト: `50`）
- `REDIS_HOST` — Redis ホスト（デフォルト: `redis`）
- `REDIS_PORT` — Redis ポート（デフォルト: `6379`）
- `REDIS_DB` — Redis DB番号（デフォルト: `0`）
- `USE_PION_OPUS` — `true` にすると DCA の代わりに Pion Opus エンコーダーを使用（デフォルト: `false`）
- `LOG_LEVEL` — ログレベル（`trace`/`debug`/`info`/`warn`/`error`/`fatal`、デフォルト: `info`）
- `LOG_FORMAT` — ログ形式（`text`/`json`、デフォルト: `text`）
- `HTTP_PORT` — ヘルスチェックサーバーのポート（デフォルト: `8080`）

---

## DevContainer

```bash
# VS Code / Cursor で「Reopen in Container」を実行
# postCreateCommand により以下が自動実行される:
#   - go mod download
#   - go install github.com/mrjoshuak/godoc-mcp@latest
```

コンテナ内での Bot 起動:

```bash
go run ./cmd/bot
```

VoiceVox Engine は `compose.devcontainer.yml` に含まれており、自動起動します。

---

## スラッシュコマンド一覧

| コマンド | 説明 |
|----------|------|
| `/ping` | Bot の死活確認 |
| `/help` | コマンド一覧表示 |
| `/invite` | Bot 招待 URL 表示 |
| `/summon` | VC に参加 |
| `/bye` | VC から退出 |
| `/reconnect` | VC に手動再接続 |
| `/stop` | 読み上げ中断（キューをクリア） |
| `/speaker <ID>` | 話者を変更 |
| `/speaker_list` | 利用可能な話者一覧表示 |

---

## コーディング規約

### インターフェース設計

- `internal/commands/bot_interface.go` と `internal/events/bot_interface.go` は、各パッケージが必要とする Bot の機能のみを定義する（インターフェース分離の原則）
- 新しいコマンドを追加する場合は `internal/commands/` にファイルを作成し、`commands.Registry` に登録する

### エラー処理

- `internal/errors/errors.go` にドメインエラーを定義する
- ラップされたエラーは `fmt.Errorf("context: %w", err)` 形式で返す
- goroutine 内のパニックは `safeGoroutine` でリカバリする

### ロギング

- Logrus を使用。`logrus.WithFields()` でコンテキストを付与する
- **メッセージ内容はプライバシー保護のためログに記録しない**
- ログレベルの使い分けは README のロギング設定セクションを参照

### 並行処理

- VC 接続マップは `sync.RWMutex` で保護する（読み取りは `RLock`、書き込みは `Lock`）
- goroutine の上限は `goroutineSem`（セマフォ、上限 100）で管理する
- goroutine の起動は `runWithSemaphore` 経由で行う

### Golang ライブラリ調査

- ライブラリの仕様を調べる場合は `godoc-mcp` の `get_doc` 関数を使う

---

## HTTP エンドポイント（:8080）

| パス | メソッド | 説明 |
|------|----------|------|
| `/health` | GET | Liveness チェック（常に 200 OK） |
| `/health/ready` | GET | Readiness チェック（Discord・VoiceVox・Redis の死活確認） |
| `/metrics` | GET | メトリクス（未実装、将来対応） |

---

## CI/CD

GitHub Actions（`.github/workflows/`）で以下を自動実行:

- `main`/`master` ブランチへの push → GHCR に `ghcr.io/jo3qma/yoursaysan:main` をビルド・プッシュ
- タグ付きリリース（`v*`）→ セマンティックバージョンタグでプッシュ
- PR → ビルドのみ（プッシュなし）
