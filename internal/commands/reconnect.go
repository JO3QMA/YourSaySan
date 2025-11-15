package commands

import (
	"fmt"

	"github.com/JO3QMA/YourSaySan/internal/voice"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func ReconnectHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("reconnect")

	guildID := i.GuildID
	userID := i.Member.User.ID

	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"user_id":    userID,
		"channel_id": i.ChannelID,
	}).Debug("Reconnect command started")

	// ユーザーのVC接続を取得
	vs, err := s.State.VoiceState(guildID, userID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"guild_id": guildID,
			"user_id":  userID,
		}).Debug("User not connected to voice channel")
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "VCに接続していないため、再接続できません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	channelID := vs.ChannelID
	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Debug("User voice channel found")

	// 既存の接続を切断
	existingConn, err := b.GetVoiceConnection(guildID)
	if err == nil {
		logrus.WithFields(logrus.Fields{
			"guild_id": guildID,
		}).Debug("Disconnecting existing voice connection")
		existingConn.Leave()
		b.RemoveVoiceConnection(guildID)
	}

	// 新しい接続を作成
	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Debug("Creating new voice connection for reconnect")
	conn, err := voice.NewConnection(s, 50)
	if err != nil {
		logrus.WithError(err).Error("Failed to create voice connection")
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("VC接続の作成に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	ctx := b.GetContext()

	if err := conn.Join(ctx, guildID, channelID); err != nil {
		logrus.WithError(err).Error("Failed to reconnect to voice channel")
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("VCへの再接続に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Botに接続を登録
	b.SetVoiceConnection(guildID, conn)

	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Info("Successfully reconnected to voice channel")

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("VC <#%s> に再接続しました。", channelID),
		},
	})
}
