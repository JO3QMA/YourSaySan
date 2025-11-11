package events

import (
	"context"
	"fmt"
	"strings"

	"discord-voice-bot/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
)

// HandleMessageCreate handles message create events for text-to-speech
func HandleMessageCreate(b BotInterface, s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := b.GetLogger()
	prefix := b.GetPrefix()

	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	// Ignore messages not starting with prefix
	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	// Remove prefix and trim spaces
	text := strings.TrimSpace(strings.TrimPrefix(m.Content, prefix))
	if text == "" {
		return
	}

	// Check if user is in a voice channel
	vs, err := findUserVoiceState(s, m.GuildID, m.Author.ID)
	if err != nil {
		logger.Debug("User not in voice channel: %v", err)
		return
	}

	// Check if bot is in the same voice channel
	botVoiceState, err := findUserVoiceState(s, m.GuildID, s.State.User.ID)
	if err != nil || botVoiceState.ChannelID != vs.ChannelID {
		logger.Debug("Bot not in same voice channel")
		return
	}

	// Create message entity
	message := &entity.Message{
		Text:      text,
		UserID:    m.Author.ID,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
	}

	// Generate voice audio
	voiceUC := b.GetVoiceUseCase()
	audio, err := voiceUC.ProcessMessage(context.Background(), message)
	if err != nil {
		logger.Error("Failed to generate voice: %v", err)
		return
	}

	// Play audio in voice channel
	if err := playAudio(s, m.GuildID, audio); err != nil {
		logger.Error("Failed to play audio: %v", err)
	}
}

// findUserVoiceState finds a user's voice state in a guild
func findUserVoiceState(s *discordgo.Session, guildID, userID string) (*discordgo.VoiceState, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs, nil
		}
	}

	return nil, fmt.Errorf("user not in voice channel")
}

// playAudio plays audio data in a voice channel (placeholder implementation)
func playAudio(s *discordgo.Session, guildID string, audio []byte) error {
	// TODO: Implement actual audio playback using discordgo's voice connection
	// This requires:
	// 1. Joining voice channel
	// 2. Creating DCA audio from raw audio data
	// 3. Sending audio to voice connection

	// For now, just log that we would play audio
	fmt.Printf("Would play %d bytes of audio in guild %s\n", len(audio), guildID)
	return nil
}
