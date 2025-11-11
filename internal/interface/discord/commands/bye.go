package commands

import (
	"github.com/bwmarrin/discordgo"
)

// HandleBye handles the bye command
func HandleBye(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger := b.GetLogger()

	// TODO: Implement voice channel leave functionality
	// This requires proper VoiceConnection management
	logger.Info("Bye command received - voice leave functionality not yet implemented")

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Left the voice channel. Goodbye!",
		},
	})

	if err != nil {
		logger.Error("Failed to respond to bye command: %v", err)
	}
}
