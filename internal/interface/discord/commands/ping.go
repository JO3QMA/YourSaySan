package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// HandlePing handles the ping command
func HandlePing(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	start := time.Now()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})

	if err != nil {
		b.GetLogger().Error("Failed to respond to ping: %v", err)
		return
	}

	// Calculate and update latency
	latency := time.Since(start)

	content := fmt.Sprintf("Pong! Latency: %v", latency)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	if err != nil {
		b.GetLogger().Error("Failed to edit ping response: %v", err)
	}
}
