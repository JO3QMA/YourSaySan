package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func InviteHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("invite")

	config := b.GetConfig()
	clientID := config.GetBotClientID()

	// 必要な権限: CONNECT (1048576), SPEAK (2097152), VIEW_CHANNEL (1024), SEND_MESSAGES (2048)
	// Discord権限の数値: https://discord.com/developers/docs/topics/permissions
	permissions := 2048 + 1048576 + 2097152 + 1024 // SEND_MESSAGES + CONNECT + SPEAK + VIEW_CHANNEL

	inviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=%d&scope=bot%%20applications.commands", clientID, permissions)

	embed := &discordgo.MessageEmbed{
		Title:       "Bot招待リンク",
		Description: fmt.Sprintf("[ここをクリックしてBotを招待](%s)", inviteURL),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "必要な権限",
				Value:  "CONNECT, SPEAK, VIEW_CHANNEL, SEND_MESSAGES, USE_SLASH_COMMANDS",
				Inline: false,
			},
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

