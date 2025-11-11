// internal/interface/discord/bot.go
package discord

import (
	"context"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/yoursaysan/discord-voicevox-bot/internal/infrastructure/config"
	"github.com/yoursaysan/discord-voicevox-bot/internal/infrastructure/logger"
	"github.com/yoursaysan/discord-voicevox-bot/internal/usecase/speaker"
	"github.com/yoursaysan/discord-voicevox-bot/internal/usecase/voice"
)

// Bot represents the Discord bot
type Bot struct {
	session        *discordgo.Session
	config         *config.Config
	logger         *logger.Logger
	speakerUseCase *speaker.UseCase
	voiceUseCase   *voice.UseCase
	
	// Voice connection management
	voiceConnections sync.Map // map[guildID]*discordgo.VoiceConnection
}

// NewBot creates a new Discord bot instance
func NewBot(
	cfg *config.Config,
	log *logger.Logger,
	speakerUseCase *speaker.UseCase,
	voiceUseCase *voice.UseCase,
) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.Bot.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	bot := &Bot{
		session:        session,
		config:         cfg,
		logger:         log,
		speakerUseCase: speakerUseCase,
		voiceUseCase:   voiceUseCase,
	}

	// Set bot intents
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsMessageContent

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	// Register event handlers
	b.registerHandlers()

	// Open the websocket connection
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	b.logger.Info("Bot is now running")

	// Wait for context cancellation
	<-ctx.Done()

	b.logger.Info("Shutting down bot...")
	return b.Stop()
}

// Stop stops the bot
func (b *Bot) Stop() error {
	// Disconnect all voice connections
	b.voiceConnections.Range(func(key, value interface{}) bool {
		if vc, ok := value.(*discordgo.VoiceConnection); ok {
			vc.Disconnect()
		}
		return true
	})

	// Close the discord session
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("failed to close discord session: %w", err)
	}

	b.logger.Info("Bot stopped successfully")
	return nil
}

// registerHandlers registers all event handlers
func (b *Bot) registerHandlers() {
	b.session.AddHandler(b.handleReady)
	b.session.AddHandler(b.handleInteractionCreate)
	b.session.AddHandler(b.handleMessageCreate)
	b.session.AddHandler(b.handleVoiceStateUpdate)
}

// GetVoiceConnection retrieves the voice connection for a guild
func (b *Bot) GetVoiceConnection(guildID string) (*discordgo.VoiceConnection, bool) {
	vc, ok := b.voiceConnections.Load(guildID)
	if !ok {
		return nil, false
	}
	return vc.(*discordgo.VoiceConnection), true
}

// SetVoiceConnection stores the voice connection for a guild
func (b *Bot) SetVoiceConnection(guildID string, vc *discordgo.VoiceConnection) {
	b.voiceConnections.Store(guildID, vc)
}

// RemoveVoiceConnection removes the voice connection for a guild
func (b *Bot) RemoveVoiceConnection(guildID string) {
	b.voiceConnections.Delete(guildID)
}

