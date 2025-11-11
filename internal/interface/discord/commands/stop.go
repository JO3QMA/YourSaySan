package commands

import (
	"github.com/bwmarrin/discordgo"
)

// HandleStop handles the stop command
func HandleStop(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger := b.GetLogger()

	// TODO: Implement actual audio stopping logic
	// This would involve stopping any currently playing audio streams
	// For now, just respond with a message

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Audio playback stopped.",
		},
	})

	if err != nil {
		logger.Error("Failed to respond to stop command: %v", err)
	}
}
