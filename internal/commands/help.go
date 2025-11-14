package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func HelpHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("help")

	options := i.ApplicationCommandData().Options
	var commandName string
	if len(options) > 0 {
		commandName = options[0].StringValue()
	}

	if commandName != "" {
		// 特定のコマンドの詳細を表示
		return showCommandDetail(b, s, i, commandName)
	}

	// 全コマンドの一覧を表示
	return showCommandList(b, s, i)
}

func showCommandList(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	commands := []string{
		"`/ping` - Botの死活確認",
		"`/help` - 利用可能なコマンドの一覧または詳細を表示",
		"`/invite` - Botを他のサーバーに招待するためのURLを表示",
		"`/summon` - BotをVCに参加させる",
		"`/bye` - BotをVCから退出させる",
		"`/reconnect` - VC接続を再接続する",
		"`/stop` - 現在の読み上げを中断する",
		"`/speaker` - ユーザーの話者を設定する",
		"`/speaker_list` - 利用可能な話者の一覧を表示",
		"`/status` - Botの状態情報を表示（開発者用）",
	}

	embed := &discordgo.MessageEmbed{
		Title:       "利用可能なコマンド",
		Description: strings.Join(commands, "\n"),
		Color:       0x00ff00,
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func showCommandDetail(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate, commandName string) error {
	descriptions := map[string]string{
		"ping":        "Botの死活確認を行います。",
		"help":        "利用可能なコマンドの一覧または詳細を表示します。",
		"invite":      "Botを他のサーバーに招待するためのURLを表示します。",
		"summon":      "BotをVCに参加させます。",
		"bye":         "BotをVCから退出させます。",
		"reconnect":   "VC接続を再接続します。",
		"stop":        "現在の読み上げを中断します。",
		"speaker":     "ユーザーの話者を設定します。",
		"speaker_list": "利用可能な話者の一覧を表示します。",
		"status":      "Botの状態情報を表示します（開発者用）。",
	}

	desc, exists := descriptions[commandName]
	if !exists {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("コマンド '%s' が見つかりません。", commandName),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("コマンド: /%s", commandName),
		Description: desc,
		Color:       0x00ff00,
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

