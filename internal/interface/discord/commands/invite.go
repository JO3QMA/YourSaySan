package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleInvite handles the invite command
func HandleInvite(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	clientID := b.GetClientID()

	// Generate invite link with necessary permissions
	permissions := discordgo.PermissionViewChannel |
		discordgo.PermissionSendMessages |
		discordgo.PermissionUseSlashCommands |
		discordgo.PermissionVoiceConnect |
		discordgo.PermissionVoiceSpeak

	inviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=%d&scope=bot%%20applications.commands",
		clientID, permissions)

	content := fmt.Sprintf("**Invite me to your server!**\n\n[Click here to invite](%s)", inviteURL)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})

	if err != nil {
		b.GetLogger().Error("Failed to respond to invite command: %v", err)
	}
}
