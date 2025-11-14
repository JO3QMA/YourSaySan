package events

import (
	"context"

	"github.com/JO3QMA/YourSaySan/internal/voice"
)

// BotInterface はBotのインターフェース（循環参照を避けるため）
type BotInterface interface {
	GetConfig() ConfigInterface
	GetState() StateInterface
	GetVoiceVox() VoiceVoxAPI
	GetSpeakerManager() SpeakerManagerAPI
	GetVoiceConnection(guildID string) (*voice.Connection, error)
	RemoveVoiceConnection(guildID string)
	RecordAudioGenerationDuration(speakerID int, duration float64)
	SetQueueSize(guildID string, size int)
}

// ConfigInterface は設定のインターフェース
type ConfigInterface interface {
	GetVoiceVoxMaxMessageLength() int
	GetBotStatus() string
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
}
