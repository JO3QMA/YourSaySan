package speaker

import (
	"context"
	"fmt"

	"discord-voice-bot/internal/domain/entity"
	"discord-voice-bot/internal/domain/repository"
	"discord-voice-bot/internal/interface/voicevox"
)

// UseCase represents speaker use case
type UseCase struct {
	speakerRepo repository.SpeakerRepository
	voicevoxClient *voicevox.Client
}

// NewUseCase creates a new speaker use case
func NewUseCase(speakerRepo repository.SpeakerRepository, voicevoxClient *voicevox.Client) *UseCase {
	return &UseCase{
		speakerRepo: speakerRepo,
		voicevoxClient: voicevoxClient,
	}
}

// GetUserSpeaker gets the speaker setting for a user
func (uc *UseCase) GetUserSpeaker(ctx context.Context, userID string) (*entity.Speaker, error) {
	speaker, err := uc.speakerRepo.GetSpeaker(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user speaker: %w", err)
	}
	return speaker, nil
}

// SetUserSpeaker sets the speaker for a user
func (uc *UseCase) SetUserSpeaker(ctx context.Context, userID string, speakerID int, speakerName string) error {
	// Validate speaker exists in VoiceVox
	if err := uc.validateSpeaker(ctx, speakerID); err != nil {
		return fmt.Errorf("invalid speaker: %w", err)
	}

	speaker := &entity.Speaker{
		ID:   speakerID,
		Name: speakerName,
	}

	if err := uc.speakerRepo.SetSpeaker(ctx, userID, speaker); err != nil {
		return fmt.Errorf("failed to set user speaker: %w", err)
	}

	return nil
}

// GetAvailableSpeakers gets all available speakers from VoiceVox
func (uc *UseCase) GetAvailableSpeakers(ctx context.Context) ([]*entity.Speaker, error) {
	speakers, err := uc.voicevoxClient.GetSpeakers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get speakers from VoiceVox: %w", err)
	}

	var result []*entity.Speaker
	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			result = append(result, &entity.Speaker{
				ID:   style.ID,
				Name: fmt.Sprintf("%s (%s)", speaker.Name, style.Name),
			})
		}
	}

	return result, nil
}

// validateSpeaker validates if a speaker ID exists in VoiceVox
func (uc *UseCase) validateSpeaker(ctx context.Context, speakerID int) error {
	speakers, err := uc.voicevoxClient.GetSpeakers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get speakers for validation: %w", err)
	}

	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			if style.ID == speakerID {
				return nil // Valid speaker
			}
		}
	}

	return fmt.Errorf("speaker ID %d not found", speakerID)
}
