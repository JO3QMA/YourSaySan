package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/your-org/yoursay-bot/internal/voice"
	"github.com/your-org/yoursay-bot/internal/voicevox"
)

// BotInterface はBotのインターフェース（循環参照を避けるため）
type BotInterface interface {
	GetConfig() ConfigInterface
	GetSession() *discordgo.Session
	GetState() StateInterface
	GetVoiceVox() VoiceVoxAPI
	GetSpeakerManager() SpeakerManagerAPI
	GetContext() context.Context
	GetVoiceConnection(guildID string) (*voice.Connection, error)
	SetVoiceConnection(guildID string, conn *voice.Connection)
	RemoveVoiceConnection(guildID string)
	GetActiveVoiceConnections() int
	GetTotalQueueSize() int
}

// ConfigInterface は設定のインターフェース
type ConfigInterface interface {
	GetBotClientID() string
	GetBotOwnerID() string
	GetBotStatus() string
}

// StateInterface は状態のインターフェース
type StateInterface interface {
	IsTextChannelActive(guildID, channelID string) bool
	AddTextChannel(guildID, channelID string)
	RemoveTextChannel(guildID, channelID string)
	GetGuildCount() int
}

// SpeakerManagerAPI は話者管理のインターフェース
type SpeakerManagerAPI interface {
	GetSpeaker(ctx context.Context, userID string) (int, error)
	SetSpeaker(ctx context.Context, userID string, speakerID int) error
	GetAvailableSpeakers(ctx context.Context) ([]voicevox.Speaker, error)
	ValidSpeaker(ctx context.Context, speakerID int) (bool, error)
}

// VoiceVoxAPI はVoiceVoxクライアントのインターフェース
type VoiceVoxAPI interface {
	Speak(ctx context.Context, text string, speakerID int) ([]byte, error)
	GetSpeakers(ctx context.Context) ([]voicevox.Speaker, error)
}

