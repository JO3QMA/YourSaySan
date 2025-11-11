// internal/usecase/speaker/speaker_usecase.go
package speaker

import (
	"context"
	"fmt"

	"github.com/yoursaysan/discord-voicevox-bot/internal/domain/entity"
	"github.com/yoursaysan/discord-voicevox-bot/internal/domain/repository"
	"github.com/yoursaysan/discord-voicevox-bot/internal/interface/voicevox"
	pkgerrors "github.com/yoursaysan/discord-voicevox-bot/pkg/errors"
)

// UseCase handles speaker-related business logic
type UseCase struct {
	repo          repository.SpeakerRepository
	voicevoxClient *voicevox.Client
}

// NewUseCase creates a new speaker use case
func NewUseCase(repo repository.SpeakerRepository, voicevoxClient *voicevox.Client) *UseCase {
	return &UseCase{
		repo:          repo,
		voicevoxClient: voicevoxClient,
	}
}

// GetSpeaker retrieves the speaker configuration for a user
func (uc *UseCase) GetSpeaker(ctx context.Context, userID string) (*entity.Speaker, error) {
	speaker, err := uc.repo.Get(ctx, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			// Return default speaker if not found
			return entity.NewSpeaker(userID, 1, "四国めたん（ノーマル）"), nil
		}
		return nil, fmt.Errorf("failed to get speaker: %w", err)
	}
	return speaker, nil
}

// SetSpeaker saves the speaker configuration for a user
func (uc *UseCase) SetSpeaker(ctx context.Context, userID string, speakerID int, name string) error {
	// Validate that the speaker exists
	if err := uc.ValidateSpeakerID(ctx, speakerID); err != nil {
		return fmt.Errorf("invalid speaker ID: %w", err)
	}

	speaker := entity.NewSpeaker(userID, speakerID, name)
	if err := uc.repo.Set(ctx, speaker); err != nil {
		return fmt.Errorf("failed to set speaker: %w", err)
	}

	return nil
}

// DeleteSpeaker removes the speaker configuration for a user
func (uc *UseCase) DeleteSpeaker(ctx context.Context, userID string) error {
	if err := uc.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete speaker: %w", err)
	}
	return nil
}

// ValidateSpeakerID checks if a speaker ID is valid
func (uc *UseCase) ValidateSpeakerID(ctx context.Context, speakerID int) error {
	speakers, err := uc.voicevoxClient.GetSpeakers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get speakers from voicevox: %w", err)
	}

	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			if style.ID == speakerID {
				return nil
			}
		}
	}

	return pkgerrors.ErrSpeakerNotFound
}

// GetAvailableSpeakers retrieves all available speakers from VoiceVox
func (uc *UseCase) GetAvailableSpeakers(ctx context.Context) ([]voicevox.Speaker, error) {
	speakers, err := uc.voicevoxClient.GetSpeakers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get speakers from voicevox: %w", err)
	}
	return speakers, nil
}

