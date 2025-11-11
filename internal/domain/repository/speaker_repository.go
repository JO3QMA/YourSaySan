package repository

import (
	"context"
	"discord-voice-bot/internal/domain/entity"
)

// SpeakerRepository defines the interface for speaker data persistence
type SpeakerRepository interface {
	// GetSpeaker retrieves the speaker setting for a user
	GetSpeaker(ctx context.Context, userID string) (*entity.Speaker, error)

	// SetSpeaker saves the speaker setting for a user
	SetSpeaker(ctx context.Context, userID string, speaker *entity.Speaker) error

	// DeleteSpeaker removes the speaker setting for a user
	DeleteSpeaker(ctx context.Context, userID string) error
}
