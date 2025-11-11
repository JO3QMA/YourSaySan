package commands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleSpeaker handles the speaker command
func HandleSpeaker(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger := b.GetLogger()
	speakerUC := b.GetSpeakerUseCase()

	// Get the speaker ID from command options
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a speaker ID.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to speaker command: %v", err)
		}
		return
	}

	speakerID := int(options[0].IntValue())

	// Get available speakers to validate and get name
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
			logger.Error("Failed to respond to speaker command: %v", err)
		}
		return
	}

	// Find speaker name
	var speakerName string
	for _, speaker := range speakers {
		if speaker.ID == speakerID {
			speakerName = speaker.Name
			break
		}
	}

	if speakerName == "" {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Speaker ID %d not found. Use `/speaker_list` to see available speakers.", speakerID),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to speaker command: %v", err)
		}
		return
	}

	// Set speaker for user
	err = speakerUC.SetUserSpeaker(context.Background(), i.Member.User.ID, speakerID, speakerName)
	if err != nil {
		logger.Error("Failed to set user speaker: %v", err)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to set speaker. Please try again later.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to speaker command: %v", err)
		}
		return
	}

	// Respond with success
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Speaker set to: %s (ID: %d)", speakerName, speakerID),
		},
	})

	if err != nil {
		logger.Error("Failed to respond to speaker command: %v", err)
	}
}
