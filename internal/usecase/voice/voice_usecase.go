// internal/usecase/voice/voice_usecase.go
package voice

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/yoursaysan/discord-voicevox-bot/internal/domain/entity"
	"github.com/yoursaysan/discord-voicevox-bot/internal/interface/voicevox"
	"github.com/yoursaysan/discord-voicevox-bot/internal/usecase/speaker"
)

// UseCase handles voice generation business logic
type UseCase struct {
	speakerUseCase *speaker.UseCase
	voicevoxClient *voicevox.Client
}

// NewUseCase creates a new voice use case
func NewUseCase(speakerUseCase *speaker.UseCase, voicevoxClient *voicevox.Client) *UseCase {
	return &UseCase{
		speakerUseCase: speakerUseCase,
		voicevoxClient: voicevoxClient,
	}
}

// GenerateVoice generates audio from a message
func (uc *UseCase) GenerateVoice(ctx context.Context, message *entity.Message) ([]byte, error) {
	if err := message.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	// Clean the message text
	cleanedText := uc.cleanMessageText(message.Text)
	if cleanedText == "" {
		return nil, entity.ErrEmptyMessage
	}

	// Get the speaker for the user
	speaker, err := uc.speakerUseCase.GetSpeaker(ctx, message.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get speaker: %w", err)
	}

	// Generate the voice
	audioData, err := uc.voicevoxClient.GenerateVoice(ctx, cleanedText, speaker.SpeakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate voice: %w", err)
	}

	return audioData, nil
}

// cleanMessageText removes URLs, mentions, and other unnecessary content from the message
func (uc *UseCase) cleanMessageText(text string) string {
	// Remove URLs
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	text = urlPattern.ReplaceAllString(text, "URL")

	// Remove user mentions <@123456789>
	mentionPattern := regexp.MustCompile(`<@!?\d+>`)
	text = mentionPattern.ReplaceAllString(text, "")

	// Remove role mentions <@&123456789>
	roleMentionPattern := regexp.MustCompile(`<@&\d+>`)
	text = roleMentionPattern.ReplaceAllString(text, "")

	// Remove channel mentions <#123456789>
	channelMentionPattern := regexp.MustCompile(`<#\d+>`)
	text = channelMentionPattern.ReplaceAllString(text, "")

	// Remove custom emojis <:name:123456789>
	emojiPattern := regexp.MustCompile(`<a?:\w+:\d+>`)
	text = emojiPattern.ReplaceAllString(text, "")

	// Remove code blocks ```code```
	codeBlockPattern := regexp.MustCompile("```[^`]*```")
	text = codeBlockPattern.ReplaceAllString(text, "コードブロック")

	// Remove inline code `code`
	inlineCodePattern := regexp.MustCompile("`[^`]+`")
	text = inlineCodePattern.ReplaceAllString(text, "")

	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

