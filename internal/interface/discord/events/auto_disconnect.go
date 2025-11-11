package events

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// HandleVoiceStateUpdate handles voice state updates for auto-disconnect
func HandleVoiceStateUpdate(b BotInterface, s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	logger := b.GetLogger()

	// Ignore bot's own voice state updates
	if vsu.UserID == s.State.User.ID {
		return
	}

	// Check if bot should disconnect after being alone
	go func() {
		time.Sleep(5 * time.Minute) // Wait 5 minutes

		guild, err := s.State.Guild(vsu.GuildID)
		if err != nil {
			logger.Error("Failed to get guild for auto-disconnect check: %v", err)
			return
		}

		// Count non-bot users in bot's voice channel
		botVoiceState := findBotVoiceState(guild.VoiceStates, s.State.User.ID)
		if botVoiceState == nil {
			return // Bot not in voice channel
		}

		humanUsers := 0
		for _, vs := range guild.VoiceStates {
			if vs.ChannelID == botVoiceState.ChannelID && !isBot(s, vs.UserID) {
				humanUsers++
			}
		}

		// Disconnect if no human users left
		if humanUsers == 0 {
			logger.Info("Auto-disconnecting from empty voice channel in guild %s", vsu.GuildID)
			// TODO: Implement auto-disconnect functionality
			// This requires proper VoiceConnection management
		}
	}()
}

// findBotVoiceState finds the bot's voice state
func findBotVoiceState(voiceStates []*discordgo.VoiceState, botID string) *discordgo.VoiceState {
	for _, vs := range voiceStates {
		if vs.UserID == botID {
			return vs
		}
	}
	return nil
}

// isBot checks if a user is a bot
func isBot(s *discordgo.Session, userID string) bool {
	user, err := s.User(userID)
	return err == nil && user.Bot
}
