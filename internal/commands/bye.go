package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func ByeHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("bye")

	guildID := i.GuildID
	channelID := i.ChannelID

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
		// エラーを無視して続行
	}

	// VCから切断
	if err := conn.Leave(); err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("VCからの切断に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Botから接続を削除
	b.RemoveVoiceConnection(guildID)

	// 読み上げ対象チャンネルを削除
	b.GetState().RemoveTextChannel(guildID, channelID)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "VCから退出しました。",
		},
	})
}

