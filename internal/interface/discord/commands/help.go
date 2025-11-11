// internal/interface/discord/commands/help.go
package commands

import (
	"github.com/bwmarrin/discordgo"
)

// HandleHelp handles the help command
func HandleHelp() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "📖 ヘルプ",
		Description: "VoiceVox音声読み上げBotのコマンド一覧",
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "/ping",
				Value:  "Botの応答速度を確認します",
				Inline: false,
			},
			{
				Name:   "/summon",
				Value:  "Botをボイスチャンネルに召喚します",
				Inline: false,
			},
			{
				Name:   "/bye",
				Value:  "Botをボイスチャンネルから退出させます",
				Inline: false,
			},
			{
				Name:   "/stop",
				Value:  "現在再生中の音声を停止します",
				Inline: false,
			},
			{
				Name:   "/speaker <id>",
				Value:  "話者を設定します（IDは/speaker_listで確認）",
				Inline: false,
			},
			{
				Name:   "/speaker_list",
				Value:  "利用可能な話者の一覧を表示します",
				Inline: false,
			},
			{
				Name:   "/reconnect",
				Value:  "ボイスチャンネルに再接続します",
				Inline: false,
			},
			{
				Name:   "/invite",
				Value:  "Botの招待リンクを表示します",
				Inline: false,
			},
			{
				Name:   "/help",
				Value:  "このヘルプを表示します",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "VoiceVox Discord Bot",
		},
	}
}

