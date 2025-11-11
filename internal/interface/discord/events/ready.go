package events

import (
	"github.com/bwmarrin/discordgo"
)

// HandleReady handles the ready event
func HandleReady(b BotInterface, s *discordgo.Session, event *discordgo.Ready) {
	logger := b.GetLogger()
	status := b.GetStatus()
	clientID := b.GetClientID()

	logger.Info("Bot is ready! Logged in as: %s#%s", event.User.Username, event.User.Discriminator)

	// Set bot status
	err := s.UpdateGameStatus(0, status)
	if err != nil {
		logger.Error("Failed to set status: %v", err)
	}

	// Register slash commands
	if err := registerCommands(s, clientID); err != nil {
		logger.Error("Failed to register commands: %v", err)
	}

	logger.Info("Bot initialization completed successfully")
}

// registerCommands registers slash commands globally
func registerCommands(s *discordgo.Session, clientID string) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Check bot latency",
		},
		{
			Name:        "summon",
			Description: "Summon the bot to your voice channel",
		},
		{
			Name:        "bye",
			Description: "Disconnect the bot from voice channel",
		},
		{
			Name:        "stop",
			Description: "Stop current audio playback",
		},
		{
			Name:        "speaker",
			Description: "Set your speaker voice",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Speaker ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "speaker_list",
			Description: "Show available speakers",
		},
		{
			Name:        "help",
			Description: "Show help information",
		},
		{
			Name:        "invite",
			Description: "Get bot invite link",
		},
		{
			Name:        "reconnect",
			Description: "Reconnect to voice channel",
		},
	}

	_, err := s.ApplicationCommandBulkOverwrite(clientID, "", commands)
	return err
}
