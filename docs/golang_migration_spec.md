# Discord チャット読み上げBot - Golang移行仕様書

## 目次

1. [概要](#概要)
2. [プロジェクト構造](#プロジェクト構造)
3. [主要コンポーネント設計](#主要コンポーネント設計)
4. [データ構造](#データ構造)
5. [API連携仕様](#api連携仕様)
6. [音声処理アーキテクチャ](#音声処理アーキテクチャ)
7. [イベントハンドリング](#イベントハンドリング)
8. [スラッシュコマンド仕様](#スラッシュコマンド仕様)
9. [設定管理](#設定管理)
10. [並行処理とリソース管理](#並行処理とリソース管理)
11. [エラーハンドリング](#エラーハンドリング)
12. [インターフェース設計](#インターフェース設計)
13. [ログと監視](#ログと監視)
14. [依存関係](#依存関係)
15. [移行時の注意点](#移行時の注意点)

---

## 概要

このドキュメントは、Ruby製のDiscordチャット読み上げBotをGolangへ移行する際に必要な技術仕様を定義します。

### 移行の目的

- パフォーマンス向上（並行処理の効率化）
- 型安全性の向上
- メモリ使用量の最適化
- 将来の拡張性向上

### 保持すべき機能

- Discordのテキストチャンネルのメッセージを自動で音声読み上げ
- ユーザーごとの話者設定（VoiceVoxの話者を個別に設定可能）
- スラッシュコマンドによる操作
- Redisによる話者設定の永続化
- 自動再接続機能
- 自動切断機能（VCにBotのみが残った場合）

---

## プロジェクト構造

```
/app
├── main.go                          # エントリーポイント
├── cmd/
│   └── bot/
│       └── main.go                  # Bot起動コマンド
├── internal/
│   ├── bot/
│   │   ├── bot.go                  # Botのメイン構造体
│   │   ├── config.go               # 設定管理
│   │   └── state.go                # Bot状態管理
│   ├── voicevox/
│   │   ├── client.go               # VoiceVox APIクライアント
│   │   ├── types.go                # VoiceVox関連の型定義
│   │   └── speaker.go              # 話者情報管理
│   ├── speaker/
│   │   ├── manager.go              # 話者設定管理
│   │   └── redis.go                # Redis連携
│   ├── commands/
│   │   ├── registry.go             # コマンド登録
│   │   ├── ping.go                 # /ping コマンド
│   │   ├── help.go                 # /help コマンド
│   │   ├── invite.go               # /invite コマンド
│   │   ├── summon.go               # /summon コマンド
│   │   ├── bye.go                  # /bye コマンド
│   │   ├── reconnect.go           # /reconnect コマンド
│   │   ├── stop.go                 # /stop コマンド
│   │   ├── speaker.go              # /speaker コマンド
│   │   ├── speaker_list.go         # /speaker_list コマンド
│   │   └── status.go               # /status コマンド（開発者用）
│   ├── events/
│   │   ├── ready.go                # Readyイベント
│   │   ├── message_create.go       # メッセージ読み上げイベント
│   │   ├── voice_state_update.go  # 自動切断イベント
│   │   └── disconnect.go          # 自動再接続イベント
│   └── voice/
│       ├── connection.go           # VC接続管理
│       ├── player.go               # 音声再生管理
│       └── queue.go                # 音声再生キュー
├── pkg/
│   └── utils/
│       ├── message.go              # メッセージ変換ユーティリティ
│       └── logger.go               # ログユーティリティ
├── config/
│   └── config.yaml                 # 設定ファイル
├── go.mod                          # Go依存関係
├── go.sum                          # Go依存関係チェックサム
├── Dockerfile                      # Dockerイメージ定義
└── compose.yml                     # Docker Compose設定
```

---

## 主要コンポーネント設計

### 1. Bot構造体（`internal/bot/bot.go`）

```go
type Bot struct {
    session *discordgo.Session
    config  *Config
    state   *State
    
    // 共有リソース
    voicevox       VoiceVoxAPI      // インターフェース（テスト容易性のため）
    speakerManager SpeakerManagerAPI // インターフェース
    
    // マルチギルド対応: ギルドごとのVC接続管理
    voiceConns map[string]*voice.Connection // guildID -> connection
    connMu     sync.RWMutex
    
    // 並行処理制御
    mu sync.RWMutex
    wg sync.WaitGroup
    
    // コンテキスト
    ctx    context.Context
    cancel context.CancelFunc
    
    // リソース制限
    maxGoroutines int
    goroutineSem  chan struct{} // セマフォ（goroutine数の制限用）
}

// goroutineSemの使用例
func (b *Bot) runWithSemaphore(fn func()) {
    // セマフォを取得（ブロック可能）
    b.goroutineSem <- struct{}{}
    defer func() { <-b.goroutineSem }() // 解放
    
    b.wg.Add(1)
    go b.safeGoroutine(func() {
        defer b.wg.Done()
        fn()
    })
}
```

**責務:**
- Discordセッションの管理
- 設定ファイルの読み込み
- コマンドとイベントハンドラの登録
- 共有リソース（VoiceVox、SpeakerManager）の管理
- マルチギルド対応のVC接続管理
- ライフサイクル管理（起動、停止、グレースフルシャットダウン）

**主要メソッド:**
- `NewBot(configPath string) (*Bot, error)`: Botインスタンス作成
- `Start() error`: Bot起動
- `Stop() error`: Bot停止（タイムアウト付き）
- `RegisterCommands() error`: スラッシュコマンド登録
- `RegisterEvents() error`: イベントハンドラ登録
- `GetVoiceConnection(guildID string) (*voice.Connection, error)`: ギルドのVC接続取得
- `SetVoiceConnection(guildID string, conn *voice.Connection)`: VC接続設定
- `RemoveVoiceConnection(guildID string)`: VC接続削除

**Bot起動シーケンス:**
```go
func (b *Bot) Start() error {
    // 1. 設定ファイル読み込み（NewBot時点で完了）
    
    // 2. Redis接続
    redisClient := redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%d", b.config.Redis.Host, b.config.Redis.Port),
        DB:   b.config.Redis.DB,
    })
    if err := redisClient.Ping(b.ctx).Err(); err != nil {
        return fmt.Errorf("failed to connect to Redis: %w", err)
    }
    
    // 3. VoiceVoxクライアント初期化
    voicevoxClient := voicevox.NewClient(b.config.VoiceVox.Host)
    
    // 4. SpeakerManager初期化
    speakerManager := speaker.NewManager(redisClient, voicevoxClient)
    
    // 5. Discord接続
    session, err := discordgo.New("Bot " + b.config.Bot.Token)
    if err != nil {
        return fmt.Errorf("failed to create Discord session: %w", err)
    }
    b.session = session
    
    // 6. コマンド登録
    if err := b.RegisterCommands(); err != nil {
        return fmt.Errorf("failed to register commands: %w", err)
    }
    
    // 7. イベントハンドラ登録
    if err := b.RegisterEvents(); err != nil {
        return fmt.Errorf("failed to register events: %w", err)
    }
    
    // 8. HTTPサーバー起動（ヘルスチェック/メトリクス）
    go b.startHTTPServer()
    
    // 9. Bot Ready（Discord接続開始）
    if err := b.session.Open(); err != nil {
        return fmt.Errorf("failed to open Discord connection: %w", err)
    }
    
    return nil
}
```

### 2. VoiceVoxクライアント（`internal/voicevox/client.go`）

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
    
    // タイムアウト設定
    connectTimeout time.Duration // 接続タイムアウト: 3秒
    readTimeout    time.Duration // 読み込みタイムアウト: 10秒
    
    // リトライ設定
    maxRetries      int           // 最大リトライ回数: 3回
    retryBackoff    time.Duration // 初期バックオフ: 100ms
    retryBackoffMax time.Duration // 最大バックオフ: 2秒
    
    // レート制限
    rateLimiter *rate.Limiter // 1秒あたりのリクエスト数制限
}

// VoiceVox Engine APIの実際のレスポンス構造
// 参考: https://voicevox.github.io/voicevox_engine/api/
type Speaker struct {
    Name       string  `json:"name"`
    SpeakerUUID string `json:"speaker_uuid"`
    Styles     []Style `json:"styles"`
}

type Style struct {
    Name string `json:"name"`
    ID   int    `json:"id"`
}

// 話者IDはStyle.IDを使用する（例: 四国めたんのノーマル = 2）

type AudioQuery struct {
    AccentPhrases []AccentPhrase `json:"accent_phrases"`
    SpeedScale    float64        `json:"speedScale"`
    PitchScale    float64        `json:"pitchScale"`
    IntonationScale float64      `json:"intonationScale"`
    VolumeScale   float64        `json:"volumeScale"`
    PrePhonemeLength float64     `json:"prePhonemeLength"`
    PostPhonemeLength float64    `json:"postPhonemeLength"`
    OutputSamplingRate int       `json:"outputSamplingRate"`
    OutputStereo      bool       `json:"outputStereo"`
    Kana              string     `json:"kana,omitempty"`
}

type AccentPhrase struct {
    Moras      []Mora `json:"moras"`
    Accent     int    `json:"accent"`
    PauseMora  *Mora  `json:"pauseMora,omitempty"`
    IsInterrogative bool `json:"isInterrogative"`
}

type Mora struct {
    Text            string  `json:"text"`
    Consonant       *string `json:"consonant,omitempty"`
    ConsonantLength *float64 `json:"consonantLength,omitempty"`
    Vowel           string  `json:"vowel"`
    VowelLength     float64 `json:"vowelLength"`
    Pitch           float64 `json:"pitch"`
}

func (c *Client) Speak(ctx context.Context, text string, speakerID int) ([]byte, error)
func (c *Client) GetSpeakers(ctx context.Context) ([]Speaker, error)
func (c *Client) speakWithRetry(ctx context.Context, text string, speakerID int) ([]byte, error)
```

**責務:**
- VoiceVox Engine APIとの通信
- テキストから音声データ（WAV形式）への変換
- 話者一覧の取得
- エラーハンドリングとリトライ（指数バックオフ）
- レート制限の管理

**APIエンドポイント:**
- `POST /audio_query?text={text}&speaker={speaker_id}`: テキストから音声クエリを生成
- `POST /synthesis?speaker={speaker_id}`: クエリから音声データ（WAV形式）を生成
- `GET /speakers`: 利用可能な話者一覧を取得

**参考:**
- [VoiceVox Engine API ドキュメント](https://voicevox.github.io/voicevox_engine/api/)
- 実際のAPIレスポンス構造は上記ドキュメントを参照

**タイムアウト設定:**
- 接続タイムアウト: 3秒
- 読み込みタイムアウト: 10秒

**リトライ戦略:**
- 最大リトライ回数: 3回
- 指数バックオフ: 初回100ms、2回目200ms、3回目400ms（最大2秒）
- リトライ対象: ネットワークエラー、タイムアウト、5xxエラー
- リトライ対象外: 4xxエラー（クライアントエラー）

**レート制限:**
- デフォルト: 10リクエスト/秒（設定可能）
- `golang.org/x/time/rate`を使用

**同時実行数制限:**
- 複数ギルドで同時に大量のメッセージが投稿された場合、VoiceVox APIを過負荷にする可能性がある
- **対策**: 
  - ギルドごとにキューで順次処理: 1ギルドあたり1並行再生
  - VoiceVox API呼び出しは複数ギルドから同時発生する可能性あり
  - レート制限（10リクエスト/秒）で全体の負荷を制御
  - 例: 20ギルドで同時にメッセージ投稿 → 最大20並行リクエスト（レート制限で調整）
- **推奨設定**: 高負荷環境では、レート制限を5リクエスト/秒に下げることを検討

**HTTPクライアントの設定:**
```go
httpClient: &http.Client{
    Timeout: 10 * time.Second,  // 全体のタイムアウト（読み込みタイムアウト）
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout: 3 * time.Second,  // 接続タイムアウト
        }).DialContext,
        ResponseHeaderTimeout: 10 * time.Second,  // レスポンスヘッダーのタイムアウト
        MaxIdleConns: 100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout: 90 * time.Second,
    },
}
```

**コネクションプーリング:**
- `MaxIdleConns`: 100
- `MaxIdleConnsPerHost`: 10
- `IdleConnTimeout`: 90秒

### 3. SpeakerManager（`internal/speaker/manager.go`）

```go
type Manager struct {
    redis RedisClient // インターフェース（テスト容易性のため）
    voicevox VoiceVoxAPI // 話者一覧取得用
    
    // メモリキャッシュ（LRUキャッシュ）
    // github.com/hashicorp/golang-lru/v2 を使用
    cache      *lru.Cache[string, *cacheEntry]
    cacheTTL   time.Duration // キャッシュTTL: 5分
    maxCacheSize int         // 最大キャッシュサイズ: 1000件
    
    // 話者一覧キャッシュ
    speakersCache      []voicevox.Speaker
    speakersCacheTime  time.Time
    speakersCacheTTL   time.Duration // 話者一覧キャッシュTTL: 1時間
    speakersCacheMu    sync.RWMutex
}

type cacheEntry struct {
    speakerID int
    expires  time.Time
}

func (m *Manager) GetSpeaker(ctx context.Context, userID string) (int, error)
func (m *Manager) SetSpeaker(ctx context.Context, userID string, speakerID int) error
func (m *Manager) GetAvailableSpeakers(ctx context.Context) ([]voicevox.Speaker, error)
func (m *Manager) ValidSpeaker(ctx context.Context, speakerID int) (bool, error)
func (m *Manager) invalidateCache(userID string) // キャッシュ無効化
func (m *Manager) cleanupExpiredCache() // 期限切れキャッシュの削除
```

**責務:**
- ユーザーごとの話者設定の管理
- Redisとの連携（永続化）
- メモリキャッシュによる高速化（LRU方式）
- 話者IDの検証
- キャッシュの無効化とクリーンアップ

**Redisキー形式:**
- `speaker:{user_id}`: ユーザーの話者ID（整数）

**デフォルト値:**
- 話者IDが未設定の場合: `2`

**キャッシュ戦略:**
- **メモリキャッシュ**: `github.com/hashicorp/golang-lru/v2`を使用したLRU方式、最大1000件、TTL 5分
- **話者一覧キャッシュ**: TTL 1時間（変更頻度が低いため）
- **キャッシュ無効化**: `SetSpeaker`実行時に該当ユーザーのキャッシュを無効化（`cache.Remove(userID)`）
- **キャッシュクリーンアップ**: LRUライブラリが自動的に古いエントリを削除。TTLチェックは`GetSpeaker`時に実行し、期限切れの場合はRedisから再取得

**実装例:**
```go
func NewManager(redis RedisClient, voicevox VoiceVoxAPI) (*Manager, error) {
    cache, err := lru.New[string, *cacheEntry](1000)
    if err != nil {
        return nil, err
    }
    return &Manager{
        redis: redis,
        voicevox: voicevox,
        cache: cache,
        cacheTTL: 5 * time.Minute,
        maxCacheSize: 1000,
        speakersCacheTTL: 1 * time.Hour,
    }, nil
}
```

**複数インスタンス対応:**
- Redisを共通のデータソースとして使用
- キャッシュは各インスタンスで独立（整合性はRedisで保証）
- 話者設定変更時はRedisに書き込み、他のインスタンスのキャッシュは自然に期限切れになる

**キャッシュ整合性の考慮:**
- **許容できる遅延**: 最大5分間（TTL）は古いデータを返す可能性がある
- **理由**: 話者設定の変更頻度が低く、5分程度の遅延は許容範囲
- **代替案（将来の拡張）**: 許容できない場合は、Redis Pub/Subでキャッシュ無効化を通知する実装を検討

**Redisフェイルオーバー戦略:**
```go
// Redisダウン時の戦略:
// 1. 初回エラー: デフォルト値（話者ID: 2）を使用、エラーログを記録
// 2. 5分ごとに再接続を試行（バックグラウンドgoroutine）
// 3. 再接続成功時:
//    - キャッシュをクリア（invalidateCache）
//    - 正常動作を再開
//    - ログに復旧を記録
// 4. 長時間ダウン時（30分以上）:
//    - 警告ログを出力
//    - メトリクスでアラート（オプション）
```

**実装例:**
```go
func (m *Manager) reconnectLoop(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := m.redis.Ping(ctx).Err(); err != nil {
                logrus.WithError(err).Warn("Redis still unavailable")
                continue
            }
            // 再接続成功
            logrus.Info("Redis reconnected, clearing cache")
            m.cache.Purge() // キャッシュをクリア
        }
    }
}
```

### 4. Voice接続管理（`internal/voice/connection.go`）

```go
type Connection struct {
    session    *discordgo.Session
    guildID    string
    channelID  string
    connection *discordgo.VoiceConnection
    player     *Player
    mu         sync.RWMutex
}

type Player struct {
    queue    *Queue
    playing  atomic.Bool  // 並行アクセスがあるためatomic.Boolを使用（Go 1.19+で導入）
    stopChan chan struct{}
    mu       sync.Mutex
}

func (c *Connection) Join(ctx context.Context, guildID, channelID string) error
func (c *Connection) Leave() error
func (c *Connection) Play(ctx context.Context, audioData []byte) error
func (c *Connection) Stop() error
```

**責務:**
- Discord VCへの接続管理
- 音声データの再生
- 再生キュー管理
- 再生の中断制御

---

## データ構造

### 設定構造体（`internal/bot/config.go`）

```go
type Config struct {
    Bot struct {
        Token    string `yaml:"token"`
        ClientID string `yaml:"client_id"`
        Status   string `yaml:"status"`
        OwnerID  string `yaml:"owner"`
    } `yaml:"bot"`
    
    VoiceVox struct {
        MaxChars        int    `yaml:"max_chars"`         // VoiceVox APIの最大文字数（デフォルト: 200）
        MaxMessageLength int   `yaml:"max_message_length"` // 読み上げメッセージの最大長（デフォルト: 50）
        Host            string `yaml:"host"`
    } `yaml:"voicevox"`
    
    Redis struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
        DB   int    `yaml:"db"`
    } `yaml:"redis"`
}
```

### Bot状態（`internal/bot/state.go`）

```go
type GuildState struct {
    // VC接続情報
    VoiceConn      *voice.Connection
    VoiceChannelID string
    
    // 読み上げ対象チャンネル（現在は単一チャンネルのみ対応）
    // 将来的な拡張（複数チャンネル対応）を見越してmapで実装
    TextChannelIDs map[string]bool // channelID -> bool
    
    // 読み上げ状態
    IsReading      atomic.Bool  // Go 1.19+で導入
    LastActivity   time.Time
    
    // ロック
    mu sync.RWMutex
}

type State struct {
    // ギルドごとの状態管理
    Guilds map[string]*GuildState // guildID -> state
    mu     sync.RWMutex
}

func (s *State) GetGuildState(guildID string) *GuildState
func (s *State) SetGuildState(guildID string, state *GuildState)
func (s *State) RemoveGuildState(guildID string)
func (s *State) IsTextChannelActive(guildID, channelID string) bool
func (s *State) AddTextChannel(guildID, channelID string)
func (s *State) RemoveTextChannel(guildID, channelID string)
```

**状態遷移:**
- **未接続** → `/summon` → **接続中**
- **接続中** → `/bye` → **未接続**
- **接続中** → 自動切断（Botのみ） → **未接続**
- **接続中** → `/reconnect` → **接続中**（再接続）

### コマンド情報（`internal/commands/registry.go`）

```go
type CommandInfo struct {
    Name        string
    Description string
    Options     []*discordgo.ApplicationCommandOption
}

type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error
```

---

## API連携仕様

### Discord API

**使用ライブラリ:** `github.com/bwmarrin/discordgo`

**必要な権限:**
- `SEND_MESSAGES`: メッセージ送信
- `CONNECT`: VC接続
- `SPEAK`: 音声再生
- `VIEW_CHANNEL`: チャンネル閲覧
- `USE_SLASH_COMMANDS`: スラッシュコマンド使用

**イベント:**
- `Ready`: Bot起動完了
- `MessageCreate`: メッセージ投稿
- `VoiceStateUpdate`: VC状態変更
- `Disconnect`: WebSocket切断

### VoiceVox Engine API

**ベースURL:** 環境変数 `VOICEVOX_HOST`（デフォルト: `http://voicevox:50021`）

**エンドポイント:**

1. **POST /audio_query**
   - リクエスト: `?text={text}&speaker={speaker_id}`
   - レスポンス: `AudioQuery` JSON

2. **POST /synthesis**
   - リクエスト: `AudioQuery` JSON + `?speaker={speaker_id}`
   - レスポンス: WAV形式のバイナリデータ

3. **GET /speakers**
   - レスポンス: 話者一覧のJSON配列

**エラーハンドリング:**
- HTTPステータスコードが200以外の場合はエラーを返す
- タイムアウト時はエラーログを出力し、`nil`を返す

### Redis API

**使用ライブラリ:** `github.com/redis/go-redis/v9`

**接続情報:**
- ホスト: 環境変数 `REDIS_HOST`（デフォルト: `redis`）
- ポート: 環境変数 `REDIS_PORT`（デフォルト: `6379`）
- DB: 環境変数 `REDIS_DB`（デフォルト: `0`）

**操作:**
- `GET speaker:{user_id}`: 話者ID取得
- `SET speaker:{user_id} {speaker_id}`: 話者ID設定

**エラーハンドリング:**
- 接続エラー時はログを出力し、デフォルト値を使用
- キーが存在しない場合はデフォルト値（2）を返す

---

## 音声処理アーキテクチャ

### 音声処理パイプライン

DiscordのVCで音声を再生するには、WAV形式の音声データをOpus形式にエンコードする必要があります。この処理は以下のパイプラインで実装します：

```
VoiceVox API (WAV) → WAVファイル → DCAエンコーダー → Opusストリーム → Discord VC
```

### 1. DCAエンコーディング（`internal/voice/encoder.go`）

```go
import (
    "github.com/jonas747/dca"
    "github.com/layeh/gopus"
)

type Encoder struct {
    options *dca.EncodeOptions
}

type EncodeOptions struct {
    FrameSize   int    // フレームサイズ: 960
    Bitrate     int    // ビットレート: 64kbps
    SampleRate  int    // サンプルレート: 48000Hz
    Channels    int    // チャンネル数: 2（ステレオ）
    Application int    // Opusアプリケーション: gopus.AudioApplicationVoip
}

func NewEncoder() *Encoder {
    return &Encoder{
        options: &dca.EncodeOptions{
            FrameSize:  960,
            Bitrate:    64,
            SampleRate: 48000,
            Channels:   2,
            Application: gopus.AudioApplicationVoip,
        },
    }
}

func (e *Encoder) EncodeFile(ctx context.Context, wavPath string) (<-chan []byte, error)
func (e *Encoder) EncodeBytes(ctx context.Context, wavData []byte) (<-chan []byte, error)
// 注意: DCAライブラリはファイルベースのため、EncodeBytesは内部で一時ファイルを作成する
// 実装例:
//   1. wavDataを一時ファイルに書き込み
//   2. EncodeFileを呼び出し
//   3. エンコード完了後に一時ファイルを削除

// エラーケース:
var (
    ErrDiskFull        = errors.New("disk full")
    ErrPermissionDenied = errors.New("permission denied")
    ErrFFmpegNotFound  = errors.New("ffmpeg not found")
)

// EncodeBytesのエラーハンドリング:
// - ディスク容量不足: ErrDiskFullを返す
// - 権限エラー: ErrPermissionDeniedを返す
// - ffmpeg実行失敗: ErrFFmpegNotFoundを返す（ffmpegがPATHにない場合）
// - 一時ファイル作成失敗: 適切なエラーを返す
```

**責務:**
- WAVファイル/データをOpus形式にエンコード
- DCAライブラリを使用してエンコード処理
- ストリーミング形式でデータを返す（チャンネル経由）

**依存関係:**
- `github.com/jonas747/dca`: DCAエンコーディングライブラリ
- `github.com/layeh/gopus`: Opusエンコーディングライブラリ
- `ffmpeg`: システムにインストールが必要（DCAが内部的に使用）

**エンコーディングパラメータ:**
- **フレームサイズ**: 960サンプル（20ms @ 48kHz）
- **ビットレート**: 64kbps（Discordの推奨値）
- **サンプルレート**: 48000Hz（Discordの標準）
- **チャンネル数**: 2（ステレオ）
- **アプリケーション**: Voip（音声通話用）

### 2. 音声再生キュー（`internal/voice/queue.go`）

```go
type AudioItem struct {
    Data      []byte  // WAVデータ（可変サイズ、最大1MB/アイテム）
    GuildID   string
    ChannelID string
    UserID    string
    Timestamp time.Time
}

// メモリ使用量の制限:
// - キューアイテムの最大サイズ: 1MB/アイテム
// - 超過する場合はエラーを返す（ErrAudioTooLarge）
// - 全ギルド合計の推定メモリ: 最大50MB（50アイテム × 10ギルド × 1MB）
// - 実際の使用量は音声の長さに依存（約50KB/メッセージが目安）

type Queue struct {
    items    chan AudioItem
    maxSize  int           // 最大キューサイズ: 50件
    timeout  time.Duration // タイムアウト: 30秒
    mu       sync.Mutex
    closed   atomic.Bool
}

func NewQueue(maxSize int) *Queue {
    return &Queue{
        items:   make(chan AudioItem, maxSize),
        maxSize: maxSize,
        timeout: 30 * time.Second,
    }
}

func (q *Queue) Push(ctx context.Context, item AudioItem) error
func (q *Queue) Pop(ctx context.Context) (AudioItem, error)
func (q *Queue) Clear() // キューをクリア
func (q *Queue) Size() int // 現在のキューサイズ
func (q *Queue) Close() // キューを閉じる
```

**責務:**
- 音声データのキュー管理
- 背圧（backpressure）処理
- タイムアウト処理
- 古いメッセージの破棄戦略

**キュー戦略:**
- **最大サイズ**: 50件（設定可能）
- **キューが満杯の場合**: 古いメッセージを破棄（FIFO方式）
- **タイムアウト**: 30秒以内に処理されない場合はエラーを返す
- **バッファサイズ**: チャンネルのバッファサイズ = 最大キューサイズ

**背圧処理:**
- キューが満杯の場合、`Push`は即座にエラーを返す（ブロックしない）
- エラーメッセージ: "音声キューが満杯です。しばらく待ってから再試行してください"

### 3. 音声プレイヤー（`internal/voice/player.go`）

```go
type Player struct {
    queue     *Queue
    encoder   *Encoder
    conn      *discordgo.VoiceConnection
    playing   atomic.Bool
    stopChan  chan struct{}
    doneChan  chan struct{}
    mu        sync.Mutex
    wg        sync.WaitGroup
}

func NewPlayer(conn *discordgo.VoiceConnection, queue *Queue, encoder *Encoder) *Player

func (p *Player) Start(ctx context.Context) error
func (p *Player) Stop() error
func (p *Player) IsPlaying() bool
func (p *Player) playLoop(ctx context.Context) // 内部メソッド
func (p *Player) playAudio(ctx context.Context, item AudioItem) error // 内部メソッド
```

**責務:**
- 音声データの再生制御
- キューからデータを取得して再生
- 再生の中断制御
- エラーハンドリング

**再生フロー:**
1. `Start()`で再生ループを開始
2. キューから音声データを取得（`Pop`）
3. WAVデータをOpus形式にエンコード
4. `VoiceConnection.Speaking(true)`で再生開始
5. OpusストリームをDiscordに送信
6. 再生完了後、`VoiceConnection.Speaking(false)`
7. 次のアイテムを処理（ループ）

**エラーハンドリング:**
- エンコードエラー: ログ出力して次のアイテムへ
- 送信エラー: ログ出力して次のアイテムへ
- パニック: `recover`でキャッチしてログ出力、再生を継続

### 4. Voice接続管理（更新版）

```go
type Connection struct {
    session    *discordgo.Session
    guildID    string
    channelID  string
    connection *discordgo.VoiceConnection
    player     *Player  // Playerがencoderとqueueを所有
    mu         sync.RWMutex
}

func (c *Connection) Join(ctx context.Context, guildID, channelID string) error
// Joinは構造体のguildID/channelIDを更新し、Discord VCに接続する
func (c *Connection) Leave() error
func (c *Connection) Play(ctx context.Context, audioData []byte) error
func (c *Connection) Stop() error
func (c *Connection) QueueSize() int
```

**設計方針:**
- `Player`が`Encoder`と`Queue`を所有する
- `Connection`は`Player`への参照のみを持つ
- これにより、音声再生の責務が`Player`に集約され、設計が明確になる

**NewConnectionの実装例:**
```go
func NewConnection(session *discordgo.Session, maxQueueSize int) *Connection {
    queue := NewQueue(maxQueueSize)
    encoder := NewEncoder()
    player := NewPlayer(queue, encoder) // connは後でSetConnectionで設定
    
    return &Connection{
        session: session,
        player:  player,
    }
}
```

**Playerの初期化:**
```go
// connは後で設定する設計
func NewPlayer(queue *Queue, encoder *Encoder) *Player {
    return &Player{
        queue:    queue,
        encoder:  encoder,
        conn:     nil, // Join時にSetConnectionで設定
        playing:  atomic.Bool{},
        stopChan: make(chan struct{}),
        doneChan: make(chan struct{}),
    }
}

func (p *Player) SetConnection(conn *discordgo.VoiceConnection) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.conn = conn
}
```

**Joinメソッドの実装:**
```go
func (c *Connection) Join(ctx context.Context, guildID, channelID string) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // 構造体のフィールドを更新
    c.guildID = guildID
    c.channelID = channelID
    
    // Discord VCに接続
    vc, err := c.session.ChannelVoiceJoin(guildID, channelID, false, true)
    if err != nil {
        return fmt.Errorf("failed to join voice channel: %w", err)
    }
    
    c.connection = vc
    // Playerに接続を設定
    c.player.SetConnection(vc)
    
    return nil
}
```

**責務:**
- Discord VCへの接続管理
- 音声データの再生（キュー経由）
- 再生キュー管理
- 再生の中断制御

**一時ファイル処理:**
- WAVデータは一時ファイルに書き込む（`os.CreateTemp()`）
- 再生完了後に`defer os.Remove()`で削除
- パニック時も確実に削除するため、`defer`を使用
- 一時ファイルのパス: `/tmp/yoursay-*.wav`

**メモリ使用量の制限:**
- 大きな音声データ（> 1MB）は必ず一時ファイルに書き込む
- メモリに保持する場合は最大10MBまで
- 超過する場合は一時ファイルを使用

---

## イベントハンドリング

### 1. Readyイベント（`internal/events/ready.go`）

```go
func ReadyHandler(s *discordgo.Session, event *discordgo.Ready) {
    // Botステータス設定
    s.UpdateGameStatus(0, config.Bot.Status)
    
    // ログ出力
    log.Println("Bot is Ready!")
}
```

**処理内容:**
- Botのステータス（ゲーム表示）を設定
- コンソールに「Bot is Ready!」を出力

### 2. MessageCreateイベント（`internal/events/message_create.go`）

```go
func MessageCreateHandler(bot *bot.Bot) func(s *discordgo.Session, m *discordgo.MessageCreate) {
    return func(s *discordgo.Session, m *discordgo.MessageCreate) {
        // 1. 読み上げ対象チャンネルかチェック
        // 2. スラッシュコマンド（/で始まる）は読み上げない
        // 3. 本文が空のメッセージは読み上げない
        // 4. メッセージ変換
        // 5. 話者設定取得
        // 6. 音声生成
        // 7. 音声再生
    }
}
```

**処理フロー:**

1. **チャンネルチェック**
   - `bot.state.TextChannels`にチャンネルIDが存在するか確認

2. **メッセージフィルタリング**（軽いチェックから順に実行）
   - `m.Author.Bot`をチェック（最も一般的なケース）
   - 本文が空のメッセージはスキップ（`len(m.Content) == 0`）
   - スラッシュコマンド（`/`で始まる）はスキップ（`strings.HasPrefix(m.Content, "/")`）
   - 読み上げ対象チャンネルかチェック（最後に実行、最も重い処理）

3. **メッセージ変換**（`pkg/utils/message.go`）
   - メンション（`<@user_id>`）を表示名（`@username`）に変換
   - チャンネルメンション（`<#channel_id>`）をチャンネル名（`#channel-name`）に変換
   - ロールメンション（`<@&role_id>`）をロール名（`@role-name`）に変換
   - カスタム絵文字（`<:name:id>`）を絵文字名（`:name:`）に変換
   - URLを「URL省略」に置換（正規表現: `https?://[^\s<>\"'()]+`）
     - より厳密なマッチングで、句読点やカッコを含まないようにする
     - 実装時にテストで検証が必要
   - Markdown記法の除去（`**bold**` → `bold`、`*italic*` → `italic`など）
   - 最大文字数（デフォルト50文字）を超える場合は「以下略」を追加
   - 改行を空白に変換（読み上げの自然さのため）

4. **話者設定取得**
   - `SpeakerManager.GetSpeaker(ctx, userID)`を呼び出し

5. **音声生成**
   - `VoiceVox.Speak(ctx, text, speakerID)`を呼び出し
   - WAV形式のバイナリデータを取得

6. **音声再生**
   - 一時ファイルにWAVデータを書き込み（またはメモリから直接再生）
   - `VoiceConnection.Speaking(true)`で再生開始
   - OpusエンコードしてDiscordに送信
   - 再生完了後、`VoiceConnection.Speaking(false)`

### 3. VoiceStateUpdateイベント（`internal/events/voice_state_update.go`）

```go
func VoiceStateUpdateHandler(bot *bot.Bot) func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
    return func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
        // VCにBotのみが残っている場合、自動切断
    }
}
```

**処理内容:**
```go
func VoiceStateUpdateHandler(bot *bot.Bot) func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
    return func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
        // 1. Bot自身のVC接続を取得
        guild, err := s.Guild(vs.GuildID)
        if err != nil {
            return
        }
        
        // 2. BotのVC接続を確認
        botVoiceState := guild.VoiceStates[bot.session.State.User.ID]
        if botVoiceState == nil {
            return // BotはVCに接続していない
        }
        
        // 3. 同じVCチャンネルのメンバーをカウント
        memberCount := 0
        for _, voiceState := range guild.VoiceStates {
            if voiceState.ChannelID == botVoiceState.ChannelID {
                // Bot以外のメンバーをカウント
                if voiceState.UserID != bot.session.State.User.ID {
                    memberCount++
                }
            }
        }
        
        // 4. Bot以外のメンバーが0人の場合、切断
        if memberCount == 0 {
            conn, err := bot.GetVoiceConnection(vs.GuildID)
            if err == nil {
                conn.Leave()
            }
        }
    }
}
```

### 4. Disconnectイベント（`internal/events/disconnect.go`）

```go
func DisconnectHandler(s *discordgo.Session, event *discordgo.Disconnect) {
    log.Println("WebSocket disconnected, reconnecting...")
    // discordgoの自動再接続機能に任せる
}
```

**処理内容:**
- 切断を検知してログ出力
- discordgoの自動再接続機能に任せる
- 再接続完了時にログ出力

---

## スラッシュコマンド仕様

### コマンド登録（`internal/commands/registry.go`）

```go
type Registry struct {
    commands map[string]CommandHandler
    infos    map[string]CommandInfo
}

func (r *Registry) Register(name string, info CommandInfo, handler CommandHandler)
func (r *Registry) RegisterAll(s *discordgo.Session, guildID string) error
```

### コマンド一覧

#### 1. `/ping`

**説明:** Botの死活確認

**処理:**
```go
func PingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "Pong!",
        },
    })
}
```

#### 2. `/help`

**説明:** 利用可能なコマンドの一覧または詳細を表示

**オプション:**
- `command` (string, optional): コマンド名

**処理:**
- オプションなし: 全コマンドの一覧を表示
- オプションあり: 指定コマンドの詳細を表示

#### 3. `/invite`

**説明:** Botを他のサーバーに招待するためのURLを表示

**処理:**
- OAuth2認証URLを生成
- 必要な権限を含める
- Embed形式で表示

#### 4. `/summon`

**説明:** BotをVCに参加させる

**処理:**
1. コマンド実行者のVCを取得
2. VCに接続
3. 実行したテキストチャンネルを読み上げ対象に追加
4. 成功メッセージを返信

#### 5. `/bye`

**説明:** BotをVCから退出させる

**処理:**
1. サーバーからVC接続を切断
2. 読み上げ対象チャンネルをクリア（`state.RemoveTextChannel(guildID, channelID)`）
3. 音声キューをクリア（`player.queue.Clear()`）
4. 再生を停止（`player.Stop()`）
5. 成功メッセージを返信

#### 6. `/reconnect`

**説明:** VC接続を再接続する

**処理:**
1. 現在のVC接続を切断
2. コマンド実行者のVCに再接続
3. 成功メッセージを返信

#### 7. `/stop`

**説明:** 現在の読み上げを中断する

**処理:**
1. `VoiceConnection.Stop()`で再生を停止
2. 成功メッセージを返信

#### 8. `/speaker`

**説明:** ユーザーの話者を設定する

**オプション:**
- `speaker_id` (int, required): 話者ID

**処理:**
1. VoiceVoxから利用可能な話者一覧を取得
2. 指定された話者IDが有効かチェック
3. Redisに話者設定を保存
4. 設定した話者名を返信

#### 9. `/speaker_list`

**説明:** 利用可能な話者の一覧を表示

**オプション:**
- `page` (int, optional, default: 1): ページ番号

**処理:**
1. VoiceVoxから利用可能な話者一覧を取得
2. 現在のユーザーの話者設定を取得
3. ページネーション（1ページ20件）で表示
4. 現在の設定を強調表示（▶マーク）
5. Embed形式で表示

#### 10. `/status`

**説明:** Botの状態情報を表示（開発者用）

**権限:** Botオーナーのみ実行可能

**処理:**
1. オーナーIDをチェック
2. 以下の情報をEmbed形式で表示:
   - Botの稼働時間
   - 接続中のギルド数
   - アクティブなVC接続数
   - 音声キューの合計サイズ
   - メモリ使用量
   - VoiceVox APIの状態
   - Redis接続状態

**注意:** `/eval`コマンドはセキュリティ上の理由により実装しません。Golangでは動的コード実行は推奨されず、代替として`/status`コマンドでBotの状態を確認できます。

### スラッシュコマンドの権限管理

**権限レベル:**

1. **全員実行可能**（デフォルト）
   - `/ping`: 死活確認
   - `/help`: ヘルプ表示
   - `/invite`: 招待リンク表示
   - `/summon`: VC参加（VCに接続できる権限が必要）
   - `/bye`: VC退出
   - `/reconnect`: VC再接続
   - `/stop`: 読み上げ中断
   - `/speaker`: 話者設定（自分のみ）
   - `/speaker_list`: 話者一覧表示

2. **Botオーナーのみ**
   - `/status`: Bot状態表示

**権限チェックの実装:**
```go
func isOwner(userID string, ownerID string) bool {
    return userID == ownerID
}

// コマンドハンドラ内で使用
if command == "status" {
    if !isOwner(i.Member.User.ID, config.Bot.OwnerID) {
        return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "このコマンドはBotオーナーのみ実行可能です。",
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })
    }
}
```

**Discord権限:**
- `/summon`実行時: ユーザーがVCに接続している必要がある
- Botが必要な権限: `CONNECT`, `SPEAK`, `VIEW_CHANNEL`, `SEND_MESSAGES`

---

## 設定管理

### 設定ファイル（`config/config.yaml`）

```yaml
bot:
  token: ${DISCORD_BOT_TOKEN}                    # 環境変数から取得
  client_id: ${DISCORD_CLIENT_ID}                # 環境変数から取得
  status: "[TESTING] 読み上げBot"
  owner: ${DISCORD_OWNER_ID:-123456789012345678} # デフォルト値あり

voicevox:
  max_chars: 200        # VoiceVox APIの最大文字数
  max_message_length: 50 # 読み上げメッセージの最大長（超過時は「以下略」を追加）
  host: ${VOICEVOX_HOST:-http://voicevox:50021}  # デフォルト値あり

redis:
  host: ${REDIS_HOST:-redis}                     # デフォルト値あり
  port: ${REDIS_PORT:-6379}                      # デフォルト値あり
  db: ${REDIS_DB:-0}                             # デフォルト値あり
```

### 環境変数展開

Golangでは、以下のライブラリを使用して環境変数を展開:
- `github.com/joho/godotenv`: `.env`ファイル読み込み
- `os.ExpandEnv()`: 環境変数展開
- または `github.com/spf13/viper`: 設定管理（推奨）

### 設定読み込み（`internal/bot/config.go`）

```go
import (
    "github.com/spf13/viper"
    "github.com/joho/godotenv"
)

func LoadConfig(path string) (*Config, error) {
    // 1. .envファイル読み込み
    if err := godotenv.Load(); err != nil {
        logrus.Warn("No .env file found, using environment variables")
    }
    
    // 2. Viperで設定ファイル読み込み
    viper.SetConfigFile(path)
    viper.SetConfigType("yaml")
    
    // 3. 環境変数の自動読み込み（プレフィックスなし）
    // 設定ファイルの${DISCORD_BOT_TOKEN}形式と互換性を保つため
    viper.AutomaticEnv()
    // 環境変数名はそのまま使用（DISCORD_BOT_TOKEN等）
    // 必要に応じて個別にバインド:
    viper.BindEnv("bot.token", "DISCORD_BOT_TOKEN")
    viper.BindEnv("bot.client_id", "DISCORD_CLIENT_ID")
    viper.BindEnv("bot.owner", "DISCORD_OWNER_ID")
    viper.BindEnv("voicevox.host", "VOICEVOX_HOST")
    viper.BindEnv("redis.host", "REDIS_HOST")
    viper.BindEnv("redis.port", "REDIS_PORT")
    viper.BindEnv("redis.db", "REDIS_DB")
    
    // 注意: ViperのBindEnvは設定ファイル内の${VAR}を展開しない
    // YAML内の${DISCORD_BOT_TOKEN}は手動で展開する必要がある
    // 方法: os.ExpandEnv()を使用して環境変数を展開
    
    // 4. 設定ファイル読み込み
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    // 4-1. 環境変数展開（YAML内の${VAR}形式を展開）
    // 方法1: 手動で展開（推奨）
    for _, key := range viper.AllKeys() {
        val := viper.GetString(key)
        if strings.Contains(val, "${") {
            expanded := os.ExpandEnv(val)
            viper.Set(key, expanded)
        }
    }
    // 方法2: Viperの標準機能のみ使用する場合
    // YAMLには実際の値を書くか、環境変数で上書き（viper.BindEnvでバインド）
    
    // 5. Config構造体にマッピング
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    // 6. 設定バリデーション
    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return &config, nil
}

func validateConfig(config *Config) error {
    if config.Bot.Token == "" {
        return errors.New("bot token is required")
    }
    if config.Bot.ClientID == "" {
        return errors.New("bot client ID is required")
    }
    if config.VoiceVox.Host == "" {
        return errors.New("voicevox host is required")
    }
    if config.VoiceVox.MaxChars <= 0 {
        return errors.New("voicevox max chars must be positive")
    }
    return nil
}
```

**設定リロード:**
- 本番環境では設定ファイルのホットリロードは実装しない（再起動が必要）
- 開発環境では必要に応じて実装可能（`viper.WatchConfig()`を使用）

**デフォルト値の管理:**
- Viperの`SetDefault()`でデフォルト値を設定
- 環境変数で上書き可能
- 設定ファイルで上書き可能（優先順位: 環境変数 > 設定ファイル > デフォルト値）

**Config構造体のネスト:**
- Viperは`mapstructure`タグを自動的に認識するため、`yaml`タグだけで十分
- ネストされた構造体は`viper.Unmarshal()`で自動的にマッピングされる
- 例: `config.Bot.Token`は`viper.GetString("bot.token")`でアクセス可能

---

## 並行処理とリソース管理

### 並行処理の考慮事項

1. **読み上げキュー**
   - 複数のメッセージが同時に投稿された場合、キューで順次処理
   - `sync.Mutex`でキューへのアクセスを保護

2. **話者設定キャッシュ**
   - `sync.RWMutex`で読み取りロックを最適化
   - キャッシュミス時のみRedisにアクセス

3. **VoiceVox API呼び出し**
   - `context.Context`でタイムアウト制御
   - 並行リクエストは可能だが、レート制限に注意

4. **Discord API呼び出し**
   - discordgoは内部で並行処理を管理
   - レート制限は自動で処理される

### リソース管理

1. **一時ファイル**
   - `os.CreateTemp()`で一時ファイル作成
   - 再生完了後に`os.Remove()`で削除
   - `defer`で確実にクリーンアップ

2. **音声データ**
   - メモリに保持する場合はサイズ制限を設ける
   - 大きな音声データは一時ファイルに書き込む

3. **コンテキスト管理**
   - `context.WithTimeout()`でタイムアウト設定
   - `context.WithCancel()`でキャンセレーション制御

### グレースフルシャットダウン

```go
func (b *Bot) Stop() error {
    const shutdownTimeout = 30 * time.Second
    
    // 1. コンテキストをキャンセル
    b.cancel()
    
    // 2. 新しいイベントの処理を停止
    // 3. 進行中の読み上げを停止
    //    - すべてのPlayerに対してStop()を呼び出し（stopChanをクローズ）
    //    - 現在再生中のアイテムは完了させる（最大5秒待機）
    //    - キューに残っているアイテムは破棄（queue.Clear()）
    // 4. すべてのVC接続を切断
    for guildID, conn := range b.voiceConns {
        if err := conn.Leave(); err != nil {
            log.Printf("Error leaving voice channel for guild %s: %v", guildID, err)
        }
    }
    
    // 5. すべてのgoroutineの完了を待つ（タイムアウト付き）
    done := make(chan struct{})
    go func() {
        b.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        log.Println("Graceful shutdown completed")
        return nil
    case <-time.After(shutdownTimeout):
        log.Println("Shutdown timeout exceeded, forcing exit")
        return errors.New("shutdown timeout")
    }
}
```

**シャットダウン戦略:**
- **タイムアウト**: 30秒（設定可能）
- **タイムアウト超過時**: エラーを返すが、プロセスは終了を続行
- **リソースクリーンアップ**: すべてのVC接続、キュー、一時ファイルをクリーンアップ
- **ログ出力**: シャットダウンの各段階でログを出力

---

## エラーハンドリング

### エラーハンドリング戦略

1. **センチネルエラーの定義**
```go
var (
    // VoiceVox関連
    ErrVoiceVoxUnavailable = errors.New("voicevox engine unavailable")
    ErrVoiceVoxTimeout     = errors.New("voicevox request timeout")
    ErrVoiceVoxInvalidSpeaker = errors.New("invalid speaker ID")
    
    // Redis関連
    ErrRedisUnavailable = errors.New("redis unavailable")
    ErrRedisConnection  = errors.New("redis connection error")
    
    // Voice関連
    ErrNotInVoiceChannel = errors.New("user not in voice channel")
    ErrAlreadyConnected  = errors.New("already connected to voice")
    ErrQueueFull         = errors.New("audio queue is full")
    ErrQueueTimeout      = errors.New("audio queue timeout")
    ErrAudioTooLarge     = errors.New("audio data too large (max 1MB)")
    
    // 一般エラー
    ErrInvalidGuildID    = errors.New("invalid guild ID")
    ErrInvalidChannelID  = errors.New("invalid channel ID")
    ErrPermissionDenied  = errors.New("permission denied")
)
```

2. **エラーラッピング**
```go
import "fmt"

// errors.Wrapの代わりにfmt.Errorfと%wを使用
if err != nil {
    return fmt.Errorf("failed to get speaker: %w", err)
}
```

3. **エラーログ**
   - `github.com/sirupsen/logrus`を使用（構造化ログ）
   - エラー発生時にスタックトレースを出力
   - エラーレベルに応じてログレベルを設定

4. **リトライロジック**
   - VoiceVox API呼び出し失敗時はリトライ（最大3回）
   - 指数バックオフでリトライ間隔を調整
   - リトライ可能なエラーのみリトライ

5. **フォールバック**
   - Redis接続エラー時はメモリキャッシュのみ使用
   - VoiceVox APIエラー時はエラーメッセージを返信
   - エラー時もBotは継続動作（可能な限り）

### パニックリカバリー

**goroutine内でのパニック処理:**

```go
func (b *Bot) safeGoroutine(fn func()) {
    defer func() {
        if r := recover(); r != nil {
            logrus.WithFields(logrus.Fields{
                "panic": r,
                "stack": string(debug.Stack()),
            }).Error("Panic recovered in goroutine")
        }
    }()
    fn()
}

// 使用例
b.wg.Add(1)
go b.safeGoroutine(func() {
    defer b.wg.Done()
    // 処理
})
```

**音声再生ループでのパニック処理:**

```go
func (p *Player) playLoop(ctx context.Context) {
    defer func() {
        if r := recover(); r != nil {
            logrus.WithFields(logrus.Fields{
                "panic": r,
                "stack": string(debug.Stack()),
            }).Error("Panic in play loop")
            // 再生を継続（次のアイテムへ）
        }
    }()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-p.stopChan:
            return
        case item := <-p.queue.items:
            // キューから直接読み取る（busy waitを回避）
            if err := p.playAudio(ctx, item); err != nil {
                logrus.WithError(err).Error("Failed to play audio")
                continue
            }
        }
    }
}
```

### エラーケース

1. **VoiceVox API**
   - タイムアウト: 3秒（接続）、10秒（読み込み）
   - エラー時は`ErrVoiceVoxUnavailable`を返し、ログに記録
   - ユーザーには「音声生成に失敗しました」と通知
   - リトライ後も失敗する場合はエラーを返す

2. **Redis**
   - 接続エラー時は`ErrRedisUnavailable`を返し、ログを出力
   - デフォルト値を使用して処理を継続
   - 各コマンドで`nil`チェックを実施

3. **Discord接続**
   - WebSocket切断時は自動再接続
   - エラー発生時はスタックトレースを出力
   - VC接続エラー時は`ErrNotInVoiceChannel`を返す

4. **音声キュー**
   - キューが満杯の場合、`ErrQueueFull`を返す
   - タイムアウト時は`ErrQueueTimeout`を返す
   - エラー時もBotは継続動作

### エラー時のユーザーフィードバック戦略

**方針:**
- エラーはログに記録し、ユーザーには簡潔なメッセージを表示
- 致命的でないエラーはBotの動作を継続

**具体的な対応:**

1. **VoiceVox APIエラー時**
   - ログ: 詳細なエラー情報を記録
   - ユーザー: メッセージを投稿せず、静かにスキップ（読み上げが失敗したことを明示しない）
   - 理由: 頻繁なエラーメッセージはユーザー体験を損なう

2. **キューが満杯の場合**
   - ログ: 警告レベルで記録
   - ユーザー: メッセージを投稿せず、静かにスキップ
   - 理由: 一時的な負荷のため、ユーザーに通知する必要はない

3. **Redisエラー時**
   - ログ: エラーレベルで記録
   - ユーザー: デフォルト値（話者ID: 2）を使用して読み上げを継続
   - コマンド実行時: `/speaker`コマンドでエラーが発生した場合のみ、ユーザーに通知

4. **VC接続エラー時**
   - ログ: エラーレベルで記録
   - ユーザー: `/summon`コマンド実行時にエラーメッセージを返信
   - 例: "VCへの接続に失敗しました。権限を確認してください。"

---

## インターフェース設計

### インターフェース定義

テスト容易性と依存関係の逆転のため、主要なコンポーネントをインターフェースとして定義します。

```go
// internal/voicevox/interface.go
type VoiceVoxAPI interface {
    Speak(ctx context.Context, text string, speakerID int) ([]byte, error)
    GetSpeakers(ctx context.Context) ([]Speaker, error)
}

// internal/speaker/interface.go
type SpeakerManagerAPI interface {
    GetSpeaker(ctx context.Context, userID string) (int, error)
    SetSpeaker(ctx context.Context, userID string, speakerID int) error
    GetAvailableSpeakers(ctx context.Context) ([]voicevox.Speaker, error)
    ValidSpeaker(ctx context.Context, speakerID int) (bool, error)
}

// internal/speaker/redis_interface.go
type RedisClient interface {
    Get(ctx context.Context, key string) *redis.StringCmd
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
    // expiration に 0 を渡すと永続化（期限なし）
    Ping(ctx context.Context) *redis.StatusCmd
}

// 使用例: 話者設定は永続化するため expiration = 0
func (m *Manager) SetSpeaker(ctx context.Context, userID string, speakerID int) error {
    key := fmt.Sprintf("speaker:%s", userID)
    return m.redis.Set(ctx, key, speakerID, 0).Err() // 0 = 永続化
}

// internal/voice/interface.go
type VoiceConnection interface {
    Join(ctx context.Context, guildID, channelID string) error
    Leave() error
    Play(ctx context.Context, audioData []byte) error
    Stop() error
    QueueSize() int
}
```

### モック実装

テスト用のモック実装を提供します。

```go
// internal/voicevox/mock.go
type MockVoiceVox struct {
    SpeakFunc      func(ctx context.Context, text string, speakerID int) ([]byte, error)
    GetSpeakersFunc func(ctx context.Context) ([]Speaker, error)
}

func (m *MockVoiceVox) Speak(ctx context.Context, text string, speakerID int) ([]byte, error) {
    if m.SpeakFunc != nil {
        return m.SpeakFunc(ctx, text, speakerID)
    }
    return nil, errors.New("not implemented")
}

// 同様に他のインターフェースもモック実装
```

### テスト戦略

1. **単体テスト**
   - 各コンポーネントをインターフェース経由でテスト
   - モックを使用して外部依存を排除
   - **カバレッジ目標**:
     - コア機能（bot, voicevox, speaker, voice）: 80%以上
     - コマンドハンドラ: 60%以上（Discord APIのモックが複雑なため）
     - ユーティリティ（message, logger）: 90%以上

2. **統合テスト**
   - VoiceVox Engineとの統合テスト（テスト用コンテナを使用）
   - Redisとの統合テスト（テスト用Redisを使用）
   - 実際のAPIを呼び出す（必要に応じて）

3. **E2Eテスト**
   - Discord Botの動作確認（テスト用Discordサーバーを使用）
   - 実際のVCでの音声再生テスト

---

## ログと監視

### ログ戦略

**使用ライブラリ:** `github.com/sirupsen/logrus`

**ログレベル:**
- `DEBUG`: 詳細なデバッグ情報
- `INFO`: 一般的な情報（Bot起動、コマンド実行など）
- `WARN`: 警告（リトライ、フォールバックなど）
- `ERROR`: エラー（APIエラー、接続エラーなど）
- `FATAL`: 致命的なエラー（Bot起動失敗など）

**構造化ログのフィールド:**

```go
logrus.WithFields(logrus.Fields{
    "guild_id":   guildID,
    "channel_id": channelID,
    "user_id":    userID,
    "speaker_id": speakerID,
    "error":      err,
}).Error("Failed to generate audio")
```

**ログ出力形式:**
- 開発環境: テキスト形式（読みやすい）
- 本番環境: JSON形式（パースしやすい）

**ログローテーション:**
- ファイルサイズ: 100MB
- 保持ファイル数: 10個
- 圧縮: 有効

### メトリクス収集

**推奨ライブラリ:** `github.com/prometheus/client_golang`

**収集するメトリクス:**

```go
var (
    // コマンド実行回数
    commandsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "bot_commands_total",
            Help: "Total number of commands executed",
        },
        []string{"command"},
    )
    
    // 音声生成時間
    audioGenerationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "audio_generation_duration_seconds",
            Help: "Duration of audio generation",
        },
        []string{"speaker_id"},
    )
    
    // キューサイズ
    queueSize = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "audio_queue_size",
            Help: "Current size of audio queue",
        },
        []string{"guild_id"},
    )
    
    // アクティブなVC接続数
    activeVoiceConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_voice_connections",
            Help: "Number of active voice connections",
        },
    )
)

// メトリクスの初期化（init関数またはBot起動時）
func initMetrics() {
    prometheus.MustRegister(commandsTotal)
    prometheus.MustRegister(audioGenerationDuration)
    prometheus.MustRegister(queueSize)
    prometheus.MustRegister(activeVoiceConnections)
}
```

**HTTPサーバー:**
- **ポート**: 8080（設定可能、環境変数`HTTP_PORT`で変更可能）
- **エンドポイント**:
  - `/health`: 基本的なヘルスチェック
    - レスポンス: `{"status": "ok"}`
  - `/health/ready`: Readinessプローブ
    - Discord接続、VoiceVox API、Redis接続をチェック
    - すべて正常な場合のみ`200 OK`を返す
  - `/metrics`: Prometheus形式のメトリクスを提供

### ヘルスチェック

**実装例:**

```go
func (b *Bot) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (b *Bot) readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
    checks := map[string]bool{
        "discord":  b.session != nil && b.session.State != nil,
        "voicevox": b.checkVoiceVoxHealth(r.Context()),
        "redis":    b.checkRedisHealth(r.Context()),
    }
    
    allHealthy := true
    for _, healthy := range checks {
        if !healthy {
            allHealthy = false
            break
        }
    }
    
    if allHealthy {
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    json.NewEncoder(w).Encode(checks)
}
```

### トレーシング

**推奨ライブラリ:** OpenTelemetry（将来的な拡張）

- リクエストトレーシング
- 分散トレーシング（VoiceVox API呼び出しなど）
- パフォーマンス分析

---

## 依存関係

### 必須ライブラリ

```go
module github.com/JO3QMA/YourSaySan

go 1.21

require (
    // Discord Bot
    github.com/bwmarrin/discordgo v0.27.1
    
    // 音声処理
    github.com/jonas747/dca v0.0.0-20210930103944-155f5e5f0cc7
    github.com/layeh/gopus v0.0.0-202105022010-51f4f5be479f
    
    // Redis
    github.com/redis/go-redis/v9 v9.3.0
    
    // LRUキャッシュ
    github.com/hashicorp/golang-lru/v2 v2.0.7
    
    // 設定管理
    github.com/spf13/viper v1.18.2
    
    // ログ
    github.com/sirupsen/logrus v1.9.3
    
    // レート制限
    golang.org/x/time v0.5.0
    
    // HTTPクライアント（標準ライブラリで十分だが、必要に応じて）
    // github.com/go-resty/resty/v2 v2.11.0
    
    // メトリクス（オプション）
    github.com/prometheus/client_golang v1.18.0
)
```

### システム依存

**Dockerイメージに含める必要があるもの:**
- `ffmpeg`: DCAエンコーディングに必要
- `opus-tools`: Opusエンコーディングに必要（オプション）

**Dockerfile例:**

```dockerfile
FROM golang:1.21-alpine AS builder

# ビルド依存関係
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o yoursay-bot ./cmd/bot

# ランタイムイメージ
FROM alpine:latest

# システム依存関係
RUN apk add --no-cache ffmpeg opus-tools ca-certificates

WORKDIR /app
COPY --from=builder /app/yoursay-bot .

CMD ["./yoursay-bot"]
```

### バージョン管理

- **Go**: 1.21以上
- **Discordgo**: 最新の安定版
- **DCA**: 最新のコミット（メンテナンスが活発でないため）
- **Redis**: 6.0以上（go-redis/v9が対応）

### セキュリティ

- 依存関係の脆弱性スキャン: `go list -json -m all | nancy sleuth`
- 定期的な依存関係の更新
- セキュリティパッチの迅速な適用

---

## 移行時の注意点

### RubyからGolangへの主な違い

1. **型システム**
   - Golangは静的型付け、Rubyは動的型付け
   - 型アサーションや型変換が必要

2. **エラーハンドリング**
   - Golangは明示的なエラーハンドリング（`error`型）
   - `nil`チェックを忘れない

3. **並行処理**
   - Golangは`goroutine`と`channel`を使用
   - `sync`パッケージで同期制御

4. **メモリ管理**
   - Golangはガベージコレクション
   - 明示的なメモリ解放は不要だが、リソースのクリーンアップは必要

5. **動的実行**
   - Rubyの`eval`に相当する機能は標準では提供されていない
   - `/eval`コマンドはセキュリティ上の理由により実装しない
   - 代替として`/status`コマンドでBotの状態を確認可能

### 実装の優先順位

1. **Phase 1: 基本機能**
   - Bot起動・接続
   - 基本的なスラッシュコマンド（`/ping`, `/help`）
   - 設定ファイル読み込み

2. **Phase 2: コア機能**
   - VoiceVox API連携
   - Redis連携
   - メッセージ読み上げ

3. **Phase 3: 拡張機能**
   - 全スラッシュコマンド実装
   - 自動再接続・自動切断
   - エラーハンドリング強化

4. **Phase 4: 最適化**
   - パフォーマンスチューニング
   - メモリ使用量の最適化
   - ログの改善

### テスト戦略

1. **単体テスト**
   - 各コンポーネントの単体テスト
   - モックを使用したAPI連携テスト

2. **統合テスト**
   - VoiceVox Engineとの統合テスト
   - Redisとの統合テスト

3. **E2Eテスト**
   - Discord Botの動作確認
   - 実際のVCでの音声再生テスト

### デプロイメント

1. **Dockerイメージ**
   - マルチステージビルドでイメージサイズを最適化
   - ベースイメージは`golang:alpine`を使用

2. **環境変数**
   - `.env`ファイルまたは環境変数で設定
   - シークレット情報は環境変数で管理

3. **ログ**
   - 構造化ログ（JSON形式）を出力
   - ログレベルを設定可能に

---

## まとめ

この仕様書は、Ruby製のDiscordチャット読み上げBotをGolangへ移行する際に必要な技術仕様を定義しています。以下の重要な改善点を含んでいます：

### 主要な改善点

1. **マルチギルド対応**: ギルドごとの状態管理とVC接続管理を実装
2. **音声処理アーキテクチャ**: DCAエンコーディング、Opusストリーミング、キュー管理の詳細設計
3. **並行処理の強化**: キュー、背圧処理、リソース制限の明確な定義
4. **エラーハンドリング**: センチネルエラー、パニックリカバリー、エラーラッピング戦略
5. **インターフェース設計**: テスト容易性を向上させるためのインターフェース定義
6. **ログと監視**: 構造化ログ、メトリクス収集、ヘルスチェックエンドポイント
7. **セキュリティ**: `/eval`コマンドの削除、設定バリデーション、依存関係の脆弱性スキャン

### 実装時の注意点

- **音声処理**: ffmpegのインストールが必要（Dockerイメージに含める）
- **マルチギルド**: 各ギルドで独立したVC接続とキューを管理
- **リソース制限**: goroutine数、メモリ使用量、キューサイズの制限を設定
- **エラーハンドリング**: すべてのエラーケースを適切に処理し、Botは継続動作を維持

移行時は、段階的に実装を進め、各フェーズで動作確認を行うことを推奨します。特に音声処理とマルチギルド対応については、設計レビューを実施してから実装に着手してください。

