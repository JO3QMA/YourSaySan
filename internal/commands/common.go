package commands

import (
	"github.com/bwmarrin/discordgo"
)

// RegisterAllCommands はすべてのコマンドを登録する
func RegisterAllCommands(b BotInterface) *Registry {
	reg := NewRegistry(b)

	// 各コマンドを登録
	reg.Register("ping", CommandInfo{
		Name:        "ping",
		Description: "Botの死活確認",
		Options:     nil,
	}, PingHandler)

	reg.Register("help", CommandInfo{
		Name:        "help",
		Description: "利用可能なコマンドの一覧または詳細を表示",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "command",
				Description: "コマンド名",
				Required:    false,
			},
		},
	}, HelpHandler)

	reg.Register("invite", CommandInfo{
		Name:        "invite",
		Description: "Botを他のサーバーに招待するためのURLを表示",
		Options:     nil,
	}, InviteHandler)

	reg.Register("summon", CommandInfo{
		Name:        "summon",
		Description: "BotをVCに参加させる",
		Options:     nil,
	}, SummonHandler)

	reg.Register("bye", CommandInfo{
		Name:        "bye",
		Description: "BotをVCから退出させる",
		Options:     nil,
	}, ByeHandler)

	reg.Register("reconnect", CommandInfo{
		Name:        "reconnect",
		Description: "VC接続を再接続する",
		Options:     nil,
	}, ReconnectHandler)

	reg.Register("stop", CommandInfo{
		Name:        "stop",
		Description: "現在の読み上げを中断する",
		Options:     nil,
	}, StopHandler)

	reg.Register("speaker", CommandInfo{
		Name:        "speaker",
		Description: "ユーザーの話者を設定する",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "speaker_id",
				Description: "話者ID",
				Required:    true,
			},
		},
	}, SpeakerHandler)

	reg.Register("speaker_list", CommandInfo{
		Name:        "speaker_list",
		Description: "利用可能な話者の一覧を表示",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "page",
				Description: "ページ番号",
				Required:    false,
			},
		},
	}, SpeakerListHandler)

	reg.Register("status", CommandInfo{
		Name:        "status",
		Description: "Botの状態情報を表示（開発者用）",
		Options:     nil,
	}, StatusHandler)

	return reg
}

