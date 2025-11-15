package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func ReadyHandler(b BotInterface) func(s *discordgo.Session, event *discordgo.Ready) {
	return func(s *discordgo.Session, event *discordgo.Ready) {
		logrus.WithFields(logrus.Fields{
			"user_id":    s.State.User.ID,
			"username":   s.State.User.Username,
			"guild_count": len(event.Guilds),
		}).Info("Discord Ready event received")

		config := b.GetConfig()
		// Botステータス設定
		s.UpdateGameStatus(0, config.GetBotStatus())
		logrus.WithField("status", config.GetBotStatus()).Debug("Bot status updated")

		// Discordにコマンドを登録（Readyイベント後なのでs.State.Userが利用可能）
		if err := b.RegisterCommandsToDiscord(); err != nil {
			logrus.WithError(err).Error("Failed to register commands to Discord")
		} else {
			logrus.Debug("Commands registered to Discord")
		}

		// ログ出力
		logrus.Info("Bot is Ready!")
	}
}

