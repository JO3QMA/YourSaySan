// internal/interface/discord/handler_interaction.go
package discord

import (
	"github.com/bwmarrin/discordgo"
)

// handleInteractionCreate handles slash command interactions
func (b *Bot) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "ping":
		b.handlePingCommand(s, i)
	case "summon":
		b.handleSummonCommand(s, i)
	case "bye":
		b.handleByeCommand(s, i)
	case "stop":
		b.handleStopCommand(s, i)
	case "speaker":
		b.handleSpeakerCommand(s, i)
	case "speaker_list":
		b.handleSpeakerListCommand(s, i)
	case "help":
		b.handleHelpCommand(s, i)
	case "invite":
		b.handleInviteCommand(s, i)
	case "reconnect":
		b.handleReconnectCommand(s, i)
	default:
		b.respondWithError(s, i, "Unknown command")
	}
}

// respondWithMessage sends a response message
func (b *Bot) respondWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		b.logger.Errorf("Failed to respond to interaction: %v", err)
	}
}

// respondWithEmbed sends a response with an embed
func (b *Bot) respondWithEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		b.logger.Errorf("Failed to respond to interaction: %v", err)
	}
}

// respondWithError sends an error response
func (b *Bot) respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		b.logger.Errorf("Failed to respond with error: %v", err)
	}
}

