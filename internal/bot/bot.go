package bot

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/JO3QMA/YourSaySan/internal/commands"
	"github.com/JO3QMA/YourSaySan/internal/events"
	"github.com/JO3QMA/YourSaySan/internal/speaker"
	"github.com/JO3QMA/YourSaySan/internal/voice"
	"github.com/JO3QMA/YourSaySan/internal/voicevox"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	session *discordgo.Session
	config  *Config
	state   *State

	// 共有リソース
	voicevox       commands.VoiceVoxAPI       // インターフェース（テスト容易性のため）
	speakerManager commands.SpeakerManagerAPI // インターフェース

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

	// コマンドレジストリ
	commandRegistry *commands.Registry

	// HTTPサーバー（ヘルスチェック/メトリクス）
	httpServer *http.Server
}

func NewBot(configPath string) (*Bot, error) {
	// 1. 設定ファイル読み込み
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	b := &Bot{
		config:          config,
		state:           NewState(),
		voiceConns:      make(map[string]*voice.Connection),
		ctx:             ctx,
		cancel:          cancel,
		maxGoroutines:   100,
		goroutineSem:    make(chan struct{}, 100),
		commandRegistry: nil, // Start()で初期化
	}

	return b, nil
}

func (b *Bot) Start() error {
	// 1. 設定ファイル読み込み（NewBot時点で完了）
	logrus.Debug("Config file already loaded")

	// 2. Redis接続
	logrus.WithFields(logrus.Fields{
		"host": b.config.Redis.Host,
		"port": b.config.Redis.Port,
		"db":   b.config.Redis.DB,
	}).Info("Connecting to Redis")
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", b.config.Redis.Host, b.config.Redis.Port),
		DB:   b.config.Redis.DB,
	})
	if err := redisClient.Ping(b.ctx).Err(); err != nil {
		logrus.WithError(err).Error("Failed to connect to Redis")
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	logrus.Info("Redis connection established")

	// 3. VoiceVoxクライアント初期化
	logrus.WithField("host", b.config.VoiceVox.Host).Info("Initializing VoiceVox client")
	voicevoxClient := voicevox.NewClient(b.config.VoiceVox.Host)
	b.voicevox = voicevoxClient
	logrus.Debug("VoiceVox client initialized")

	// 4. SpeakerManager初期化
	logrus.Debug("Initializing SpeakerManager")
	speakerManager, err := speaker.NewManager(redisClient, voicevoxClient)
	if err != nil {
		logrus.WithError(err).Error("Failed to create speaker manager")
		return fmt.Errorf("failed to create speaker manager: %w", err)
	}
	b.speakerManager = speakerManager
	logrus.Debug("SpeakerManager initialized")

	// 5. Discord接続
	logrus.Info("Creating Discord session")
	session, err := discordgo.New("Bot " + b.config.Bot.Token)
	if err != nil {
		logrus.WithError(err).Error("Failed to create Discord session")
		return fmt.Errorf("failed to create Discord session: %w", err)
	}
	b.session = session
	logrus.Debug("Discord session created")

	// 6. コマンド準備（レジストリ作成とインタラクションハンドラ登録）
	logrus.Debug("Preparing commands")
	if err := b.PrepareCommands(); err != nil {
		logrus.WithError(err).Error("Failed to prepare commands")
		return fmt.Errorf("failed to prepare commands: %w", err)
	}
	logrus.Debug("Commands prepared")

	// 7. イベントハンドラ登録
	logrus.Debug("Registering event handlers")
	if err := b.RegisterEvents(); err != nil {
		logrus.WithError(err).Error("Failed to register events")
		return fmt.Errorf("failed to register events: %w", err)
	}
	logrus.Debug("Event handlers registered")

	// 8. HTTPサーバー起動（ヘルスチェック/メトリクス）
	logrus.Debug("Starting HTTP server")
	go b.startHTTPServer()
	logrus.Debug("HTTP server started")

	// 9. Bot Ready（Discord接続開始）
	// コマンドのDiscord登録はReadyイベントハンドラ内で実行される
	logrus.Info("Opening Discord connection")
	if err := b.session.Open(); err != nil {
		logrus.WithError(err).Error("Failed to open Discord connection")
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}
	logrus.Debug("Discord connection opened")

	return nil
}

func (b *Bot) Stop() error {
	const shutdownTimeout = 30 * time.Second

	logrus.Debug("Starting bot shutdown process")

	// 1. コンテキストをキャンセル
	logrus.Debug("Cancelling context")
	b.cancel()

	// 2. 新しいイベントの処理を停止
	// 3. 進行中の読み上げを停止
	//    - すべてのPlayerに対してStop()を呼び出し（stopChanをクローズ）
	//    - 現在再生中のアイテムは完了させる（最大5秒待機）
	//    - キューに残っているアイテムは破棄（queue.Clear()）
	// 4. すべてのVC接続を切断
	logrus.Debug("Disconnecting all voice connections")
	b.connMu.Lock()
	connCount := len(b.voiceConns)
	for guildID, conn := range b.voiceConns {
		logrus.WithField("guild_id", guildID).Debug("Leaving voice channel")
		if err := conn.Leave(); err != nil {
			logrus.WithError(err).WithField("guild_id", guildID).Error("Error leaving voice channel")
		}
	}
	b.voiceConns = make(map[string]*voice.Connection)
	b.connMu.Unlock()
	logrus.WithField("disconnected_count", connCount).Info("All voice connections disconnected")

	// 5. HTTPサーバーを停止
	if b.httpServer != nil {
		logrus.Debug("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := b.httpServer.Shutdown(shutdownCtx); err != nil {
			logrus.WithError(err).Error("Error shutting down HTTP server")
		} else {
			logrus.Debug("HTTP server shut down successfully")
		}
	}

	// 6. Discordセッションを閉じる
	if b.session != nil {
		logrus.Debug("Closing Discord session")
		if err := b.session.Close(); err != nil {
			logrus.WithError(err).Error("Error closing Discord session")
		} else {
			logrus.Debug("Discord session closed")
		}
	}

	// 7. すべてのgoroutineの完了を待つ（タイムアウト付き）
	logrus.Debug("Waiting for all goroutines to complete")
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logrus.Info("Graceful shutdown completed")
		return nil
	case <-time.After(shutdownTimeout):
		logrus.Warn("Shutdown timeout exceeded, forcing exit")
		return fmt.Errorf("shutdown timeout")
	}
}

func (b *Bot) PrepareCommands() error {
	// コマンドレジストリを作成
	registry := commands.RegisterAllCommands(b)
	b.commandRegistry = registry

	// インタラクションハンドラを登録
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		registry.HandleInteraction(s, i)
	})

	return nil
}

func (b *Bot) RegisterCommandsToDiscord() error {
	if b.commandRegistry == nil {
		return fmt.Errorf("command registry is not initialized")
	}

	// Discordにコマンドを登録（Readyイベント後に呼ばれる）
	if err := b.commandRegistry.RegisterAll(b.session); err != nil {
		return fmt.Errorf("failed to register commands to Discord: %w", err)
	}

	return nil
}

func (b *Bot) RegisterEvents() error {
	// events.BotInterfaceを実装するラッパーを作成
	eventsBot := &eventsBotWrapper{bot: b}

	// Readyイベント
	b.session.AddHandler(events.ReadyHandler(eventsBot))

	// MessageCreateイベント
	b.session.AddHandler(events.MessageCreateHandler(eventsBot))

	// VoiceStateUpdateイベント
	b.session.AddHandler(events.VoiceStateUpdateHandler(eventsBot))

	// Disconnectイベント
	b.session.AddHandler(events.DisconnectHandler)

	return nil
}

// eventsBotWrapper はevents.BotInterfaceを実装するラッパー
type eventsBotWrapper struct {
	bot *Bot
}

func (w *eventsBotWrapper) GetConfig() events.ConfigInterface {
	return w.bot.config
}

func (w *eventsBotWrapper) GetState() events.StateInterface {
	return w.bot.state
}

func (w *eventsBotWrapper) GetVoiceVox() events.VoiceVoxAPI {
	return w.bot.voicevox
}

func (w *eventsBotWrapper) GetSpeakerManager() events.SpeakerManagerAPI {
	return w.bot.speakerManager
}

func (w *eventsBotWrapper) GetVoiceConnection(guildID string) (*voice.Connection, error) {
	return w.bot.GetVoiceConnection(guildID)
}

func (w *eventsBotWrapper) RemoveVoiceConnection(guildID string) {
	w.bot.RemoveVoiceConnection(guildID)
}

func (w *eventsBotWrapper) RecordAudioGenerationDuration(speakerID int, duration float64) {
	w.bot.RecordAudioGenerationDuration(speakerID, duration)
}

func (w *eventsBotWrapper) SetQueueSize(guildID string, size int) {
	w.bot.SetQueueSize(guildID, size)
}

func (w *eventsBotWrapper) RegisterCommandsToDiscord() error {
	return w.bot.RegisterCommandsToDiscord()
}

func (b *Bot) GetVoiceConnection(guildID string) (*voice.Connection, error) {
	b.connMu.RLock()
	defer b.connMu.RUnlock()

	conn, ok := b.voiceConns[guildID]
	if !ok {
		return nil, fmt.Errorf("no voice connection for guild %s", guildID)
	}
	return conn, nil
}

func (b *Bot) SetVoiceConnection(guildID string, conn *voice.Connection) {
	b.connMu.Lock()
	defer b.connMu.Unlock()
	b.voiceConns[guildID] = conn
}

func (b *Bot) RemoveVoiceConnection(guildID string) {
	b.connMu.Lock()
	defer b.connMu.Unlock()
	delete(b.voiceConns, guildID)
}

func (b *Bot) GetActiveVoiceConnections() int {
	b.connMu.RLock()
	defer b.connMu.RUnlock()
	return len(b.voiceConns)
}

func (b *Bot) GetTotalQueueSize() int {
	b.connMu.RLock()
	defer b.connMu.RUnlock()

	total := 0
	for _, conn := range b.voiceConns {
		total += conn.QueueSize()
	}
	return total
}

// BotInterface実装（commands/eventsから使用）
func (b *Bot) GetConfig() commands.ConfigInterface {
	return b.config
}

func (b *Bot) GetSession() *discordgo.Session {
	return b.session
}

func (b *Bot) GetState() commands.StateInterface {
	return b.state
}

func (b *Bot) GetVoiceVox() commands.VoiceVoxAPI {
	return b.voicevox
}

func (b *Bot) GetSpeakerManager() commands.SpeakerManagerAPI {
	return b.speakerManager
}

func (b *Bot) GetContext() context.Context {
	return b.ctx
}

func (b *Bot) RecordAudioGenerationDuration(speakerID int, duration float64) {
	// メトリクス記録（将来の実装）
}

func (b *Bot) SetQueueSize(guildID string, size int) {
	// メトリクス記録（将来の実装）
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

func (b *Bot) safeGoroutine(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"panic": r,
			}).Error("Panic recovered in goroutine")
		}
	}()
	fn()
}
