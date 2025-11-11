package commands

import (
	"github.com/bwmarrin/discordgo"
)

// HandleReconnect handles the reconnect command
func HandleReconnect(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger := b.GetLogger()

	// Find user's voice channel
	vs, err := findUserVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You must be in a voice channel to reconnect!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to reconnect command: %v", err)
		}
		return
	}

	// TODO: Implement voice channel leave functionality
	// This requires proper VoiceConnection management
	logger.Info("Reconnect command - voice leave functionality not yet implemented")

	// Rejoin voice channel
	_, err = s.ChannelVoiceJoin(i.GuildID, vs.ChannelID, false, true)
	if err != nil {
		logger.Error("Failed to rejoin voice channel: %v", err)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to reconnect to voice channel. Please try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to reconnect command: %v", err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Reconnected to your voice channel!",
		},
	})

	if err != nil {
		logger.Error("Failed to respond to reconnect command: %v", err)
	}
}
