package errors

import "errors"

var (
	// VoiceVox関連
	ErrVoiceVoxUnavailable   = errors.New("voicevox engine unavailable")
	ErrVoiceVoxTimeout        = errors.New("voicevox request timeout")
	ErrVoiceVoxInvalidSpeaker = errors.New("invalid speaker ID")

	// Redis関連
	ErrRedisUnavailable  = errors.New("redis unavailable")
	ErrRedisConnection   = errors.New("redis connection error")

	// Voice関連
	ErrNotInVoiceChannel = errors.New("user not in voice channel")
	ErrAlreadyConnected  = errors.New("already connected to voice")
	ErrQueueFull         = errors.New("audio queue is full")
	ErrQueueTimeout      = errors.New("audio queue timeout")
	ErrAudioTooLarge     = errors.New("audio data too large (max 1MB)")

	// 一般エラー
	ErrInvalidGuildID   = errors.New("invalid guild ID")
	ErrInvalidChannelID = errors.New("invalid channel ID")
	ErrPermissionDenied  = errors.New("permission denied")
)

