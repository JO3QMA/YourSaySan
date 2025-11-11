// internal/domain/repository/speaker_repository.go
package repository

import (
	"context"

	"github.com/yoursaysan/discord-voicevox-bot/internal/domain/entity"
)

// SpeakerRepository defines the interface for speaker persistence
type SpeakerRepository interface {
	// Get retrieves the speaker configuration for a user
	Get(ctx context.Context, userID string) (*entity.Speaker, error)
	
	// Set saves the speaker configuration for a user
	Set(ctx context.Context, speaker *entity.Speaker) error
	
	// Delete removes the speaker configuration for a user
	Delete(ctx context.Context, userID string) error
	
	// Exists checks if a speaker configuration exists for a user
	Exists(ctx context.Context, userID string) (bool, error)
}

