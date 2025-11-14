package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func ReadyHandler(b BotInterface) func(s *discordgo.Session, event *discordgo.Ready) {
	return func(s *discordgo.Session, event *discordgo.Ready) {
		config := b.GetConfig()
		// Botステータス設定
		s.UpdateGameStatus(0, config.GetBotStatus())

		// ログ出力
		logrus.Info("Bot is Ready!")
	}
}

