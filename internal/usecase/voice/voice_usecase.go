package voice

import (
	"context"
	"fmt"

	"discord-voice-bot/internal/domain/entity"
	"discord-voice-bot/internal/domain/repository"
	"discord-voice-bot/internal/interface/voicevox"
)

// UseCase represents voice generation use case
type UseCase struct {
	speakerRepo repository.SpeakerRepository
	voicevoxClient *voicevox.Client
}

// NewUseCase creates a new voice use case
func NewUseCase(speakerRepo repository.SpeakerRepository, voicevoxClient *voicevox.Client) *UseCase {
	return &UseCase{
		speakerRepo: speakerRepo,
		voicevoxClient: voicevoxClient,
	}
}

// GenerateVoice generates voice audio from text for a user
func (uc *UseCase) GenerateVoice(ctx context.Context, userID string, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Get user's speaker setting
	speaker, err := uc.speakerRepo.GetSpeaker(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user speaker: %w", err)
	}

	// Generate audio using VoiceVox
	audio, err := uc.voicevoxClient.GenerateAudio(ctx, text, speaker.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate audio: %w", err)
	}

	return audio, nil
}

// GenerateVoiceWithSpeaker generates voice audio with specific speaker
func (uc *UseCase) GenerateVoiceWithSpeaker(ctx context.Context, text string, speakerID int) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Generate audio using VoiceVox with specified speaker
	audio, err := uc.voicevoxClient.GenerateAudio(ctx, text, speakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate audio: %w", err)
	}

	return audio, nil
}

// ProcessMessage processes a Discord message for voice generation
func (uc *UseCase) ProcessMessage(ctx context.Context, message *entity.Message) ([]byte, error) {
	if message == nil || !message.IsValid() {
		return nil, fmt.Errorf("invalid message")
	}

	return uc.GenerateVoice(ctx, message.UserID, message.Text)
}
