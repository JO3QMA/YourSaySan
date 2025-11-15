package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func StopHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("stop")

	guildID := i.GuildID

	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"user_id":    i.Member.User.ID,
		"channel_id": i.ChannelID,
	}).Debug("Stop command started")

	// VC接続を取得
	conn, err := b.GetVoiceConnection(guildID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"guild_id": guildID,
		}).Debug("No voice connection found")
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
		logrus.WithError(err).WithFields(logrus.Fields{
			"guild_id": guildID,
		}).Error("Failed to stop playback")
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("読み上げの停止に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	logrus.WithFields(logrus.Fields{
		"guild_id": guildID,
	}).Info("Playback stopped successfully")

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "読み上げを停止しました。",
		},
	})
}

