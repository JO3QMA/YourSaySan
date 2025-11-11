package discord

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"discord-voice-bot/internal/infrastructure/config"
	"discord-voice-bot/internal/infrastructure/logger"
	"discord-voice-bot/internal/interface/discord/commands"
	"discord-voice-bot/internal/interface/discord/events"
	"discord-voice-bot/internal/interface/redis"
	"discord-voice-bot/internal/interface/voicevox"
	"discord-voice-bot/internal/usecase/speaker"
	"discord-voice-bot/internal/usecase/voice"

	"github.com/bwmarrin/discordgo"

	redisPkg "github.com/redis/go-redis/v9"
)

// Bot represents the Discord bot
type Bot struct {
	session        *discordgo.Session
	config         *config.Config
	logger         logger.Logger
	speakerUseCase *speaker.UseCase
	voiceUseCase   *voice.UseCase
	redisClient    *redisPkg.Client
	voicevoxClient *voicevox.Client
}

// NewBot creates a new Discord bot
func NewBot(cfg *config.Config) (*Bot, error) {
	// Create Discord session
	session, err := discordgo.New("Bot " + cfg.Bot.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Create logger
	log := logger.NewDefaultLogger(true) // TODO: Make configurable

	// Create Redis client
	rdb := redisPkg.NewClient(&redisPkg.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	})

	// Create VoiceVox client
	vvClient := voicevox.NewClient(cfg.VoiceVox.Host)

	// Create repositories
	speakerRepo := redis.NewRepository(rdb)

	// Create use cases
	speakerUC := speaker.NewUseCase(speakerRepo, vvClient)
	voiceUC := voice.NewUseCase(speakerRepo, vvClient)

	bot := &Bot{
		session:        session,
		config:         cfg,
		logger:         log,
		speakerUseCase: speakerUC,
		voiceUseCase:   voiceUC,
		redisClient:    rdb,
		voicevoxClient: vvClient,
	}

	// Register event handlers
	bot.registerEventHandlers()

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	// Open Discord connection
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}

	b.logger.Info("Bot is now running. Press CTRL-C to exit.")

	// Wait for termination signal
	return b.waitForShutdown(ctx)
}

// Stop stops the bot
func (b *Bot) Stop() error {
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("failed to close Discord connection: %w", err)
	}

	if err := b.redisClient.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}

	return nil
}

// registerEventHandlers registers Discord event handlers
func (b *Bot) registerEventHandlers() {
	// Register ready event
	b.session.AddHandler(b.onReady)

	// Register interaction create event (for slash commands)
	b.session.AddHandler(b.onInteractionCreate)

	// Register message create event
	b.session.AddHandler(b.onMessageCreate)

	// Register voice state update event (for auto-disconnect)
	b.session.AddHandler(b.onVoiceStateUpdate)
}

// waitForShutdown waits for shutdown signal
func (b *Bot) waitForShutdown(ctx context.Context) error {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	select {
	case <-sc:
		b.logger.Info("Shutting down...")
		return b.Stop()
	case <-ctx.Done():
		b.logger.Info("Context cancelled, shutting down...")
		return b.Stop()
	}
}

// GetSession returns the Discord session
func (b *Bot) GetSession() *discordgo.Session {
	return b.session
}

// GetConfig returns the bot configuration
func (b *Bot) GetConfig() *config.Config {
	return b.config
}

// GetLogger returns the logger
func (b *Bot) GetLogger() logger.Logger {
	return b.logger
}

// GetSpeakerUseCase returns the speaker use case
func (b *Bot) GetSpeakerUseCase() *speaker.UseCase {
	return b.speakerUseCase
}

// GetVoiceUseCase returns the voice use case
func (b *Bot) GetVoiceUseCase() *voice.UseCase {
	return b.voiceUseCase
}

// GetPrefix returns the bot prefix
func (b *Bot) GetPrefix() string {
	return b.config.Bot.Prefix
}

// GetClientID returns the bot client ID
func (b *Bot) GetClientID() string {
	return b.config.Bot.ClientID
}

// GetStatus returns the bot status
func (b *Bot) GetStatus() string {
	return b.config.Bot.Status
}

// onReady handles the ready event
func (b *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	events.HandleReady(b, s, event)
}

// onInteractionCreate handles slash command interactions
func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	commandName := i.ApplicationCommandData().Name

	switch commandName {
	case "ping":
		commands.HandlePing(b, s, i)
	case "summon":
		commands.HandleSummon(b, s, i)
	case "bye":
		commands.HandleBye(b, s, i)
	case "stop":
		commands.HandleStop(b, s, i)
	case "speaker":
		commands.HandleSpeaker(b, s, i)
	case "speaker_list":
		commands.HandleSpeakerList(b, s, i)
	case "help":
		commands.HandleHelp(b, s, i)
	case "invite":
		commands.HandleInvite(b, s, i)
	case "reconnect":
		commands.HandleReconnect(b, s, i)
	default:
		b.logger.Error("Unknown command: %s", commandName)
	}
}

// onMessageCreate handles message create events
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	events.HandleMessageCreate(b, s, m)
}

// onVoiceStateUpdate handles voice state update events
func (b *Bot) onVoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	events.HandleVoiceStateUpdate(b, s, vsu)
}
