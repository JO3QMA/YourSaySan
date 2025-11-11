package events

import (
	"discord-voice-bot/internal/infrastructure/logger"
	"discord-voice-bot/internal/usecase/speaker"
	"discord-voice-bot/internal/usecase/voice"

	"github.com/bwmarrin/discordgo"
)

// BotInterface defines the interface that events need from the bot
type BotInterface interface {
	GetLogger() logger.Logger
	GetPrefix() string
	GetClientID() string
	GetStatus() string
	GetSpeakerUseCase() *speaker.UseCase
	GetVoiceUseCase() *voice.UseCase
	GetSession() *discordgo.Session
}
