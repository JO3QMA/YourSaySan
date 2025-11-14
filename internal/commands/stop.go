package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func StopHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("stop")

	guildID := i.GuildID

	// VC接続を取得
	conn, err := b.GetVoiceConnection(guildID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "VCに接続していません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 再生を停止
	if err := conn.Stop(); err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("読み上げの停止に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "読み上げを停止しました。",
		},
	})
}

