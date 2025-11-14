package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/your-org/yoursay-bot/internal/voice"
)

func ReconnectHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("reconnect")

	guildID := i.GuildID
	userID := i.Member.User.ID

	// ユーザーのVC接続を取得
	vs, err := s.State.VoiceState(guildID, userID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "VCに接続していないため、再接続できません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	channelID := vs.ChannelID

	// 既存の接続を切断
	existingConn, err := b.GetVoiceConnection(guildID)
	if err == nil {
		existingConn.Leave()
		b.RemoveVoiceConnection(guildID)
	}

	// 新しい接続を作成
	conn := voice.NewConnection(s, 50)
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("VC <#%s> に再接続しました。", channelID),
		},
	})
}

