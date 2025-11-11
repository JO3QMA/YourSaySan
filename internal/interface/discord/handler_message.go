// internal/interface/discord/handler_message.go
package discord

import (
	"bytes"
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yoursaysan/discord-voicevox-bot/internal/domain/entity"
)

// handleMessageCreate handles message creation events for voice reading
func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.Bot {
		return
	}

	// Check if bot is in a voice channel in this guild
	vc, ok := b.GetVoiceConnection(m.GuildID)
	if !ok {
		return
	}

	// Check if the message is from a channel the bot is monitoring
	// (In this case, we read all messages from the guild)
	
	// Create a message entity
	message := entity.NewMessage(
		m.ID,
		m.Content,
		m.Author.ID,
		m.GuildID,
		m.ChannelID,
	)

	// Generate voice
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	audioData, err := b.voiceUseCase.GenerateVoice(ctx, message)
	if err != nil {
		b.logger.Errorf("Failed to generate voice: %v", err)
		return
	}

	// Play the audio
	if err := b.playAudio(vc, audioData); err != nil {
		b.logger.Errorf("Failed to play audio: %v", err)
		return
	}
}

// playAudio plays audio data in a voice connection
func (b *Bot) playAudio(vc *discordgo.VoiceConnection, audioData []byte) error {
	// Wait for the voice connection to be ready
	if !vc.Ready {
		return nil
	}

	// Start speaking
	vc.Speaking(true)
	defer vc.Speaking(false)

	// Use DCA (Discord Compatible Audio) format
	// VoiceVox returns WAV format, we need to convert it to Opus
	// For now, we'll use ffmpeg to convert on-the-fly
	return b.playAudioWithFFmpeg(vc, audioData)
}

// playAudioWithFFmpeg converts WAV to Opus using ffmpeg and plays it
func (b *Bot) playAudioWithFFmpeg(vc *discordgo.VoiceConnection, wavData []byte) error {
	// Note: This requires ffmpeg to be installed
	// We use ffmpeg to convert WAV to Opus format that Discord expects
	
	// For now, we'll use a simple implementation
	// In production, you might want to use a proper audio processing library
	
	// Create a buffer for the audio data
	buffer := bytes.NewReader(wavData)
	
	// TODO: Implement proper audio conversion and streaming
	// This is a placeholder that will need proper implementation
	// with ffmpeg or another audio processing library
	
	// For now, just log that we would play the audio
	b.logger.Debugf("Would play %d bytes of audio", buffer.Len())
	
	return nil
}

