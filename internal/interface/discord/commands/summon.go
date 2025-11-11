package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleSummon handles the summon command
func HandleSummon(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger := b.GetLogger()

	// Find user's voice channel
	vs, err := findUserVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You must be in a voice channel to summon me!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to summon command: %v", err)
		}
		return
	}

	// Join voice channel
	_, err = s.ChannelVoiceJoin(i.GuildID, vs.ChannelID, false, true)
	if err != nil {
		logger.Error("Failed to join voice channel: %v", err)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to join voice channel. Please try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Error("Failed to respond to summon command: %v", err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Joined your voice channel! Use the prefix command to start text-to-speech.",
		},
	})

	if err != nil {
		logger.Error("Failed to respond to summon command: %v", err)
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
