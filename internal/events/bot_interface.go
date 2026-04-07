package events

import (
	"context"

	"github.com/JO3QMA/YourSaySan/internal/voice"
	"github.com/bwmarrin/discordgo"
)

// BotInterface はBotのインターフェース（循環参照を避けるため）
type BotInterface interface {
	GetConfig() ConfigInterface
	GetState() StateInterface
	GetSession() *discordgo.Session
	GetVoiceVox() VoiceVoxAPI
	GetSpeakerManager() SpeakerManagerAPI
	GetVoiceConnection(guildID string) (*voice.Connection, error)
	RemoveVoiceConnection(guildID string)
	RecordAudioGenerationDuration(speakerID int, duration float64)
	SetQueueSize(guildID string, size int)
	RegisterCommandsToDiscord() error
	RunWithSemaphore(fn func())
}

// ConfigInterface は設定のインターフェース
type ConfigInterface interface {
	GetVoiceVoxMaxMessageLength() int
	GetBotStatus() string
	GetSenryuEnabled() bool
	GetSenryuReplyText() string
}

// StateInterface は状態のインターフェース
type StateInterface interface {
	IsTextChannelActive(guildID, channelID string) bool
}

// SpeakerManagerAPI は話者管理のインターフェース
type SpeakerManagerAPI interface {
	GetSpeaker(ctx context.Context, userID string) (int, error)
}

// VoiceVoxAPI はVoiceVoxクライアントのインターフェース
type VoiceVoxAPI interface {
	Speak(ctx context.Context, text string, speakerID int) ([]byte, error)
	CountMorae(ctx context.Context, text string, speakerID int) (int, error)
}
