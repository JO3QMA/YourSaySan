package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HandleSpeakerList handles the speaker_list command
func HandleSpeakerList(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger := b.GetLogger()
	speakerUC := b.GetSpeakerUseCase()

	// Get available speakers
	speakers, err := speakerUC.GetAvailableSpeakers(context.Background())
	if err != nil {
		logger.Error("Failed to get available speakers: %v", err)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to get speaker list. Please try again later.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to speaker_list command: %v", err)
		}
		return
	}

	// Build response content
	var content strings.Builder
	content.WriteString("Available speakers:\n")

	for i, speaker := range speakers {
		if i >= 20 { // Discord has message length limits, limit to 20 speakers
			content.WriteString(fmt.Sprintf("... and %d more speakers", len(speakers)-20))
			break
		}
		content.WriteString(fmt.Sprintf("• ID: %d - %s\n", speaker.ID, speaker.Name))
	}

	content.WriteString("\nUse `/speaker <id>` to set your speaker.")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content.String(),
		},
	})

	if err != nil {
		logger.Error("Failed to respond to speaker_list command: %v", err)
	}
}
