package commands

import (
	"fmt"

	"github.com/JO3QMA/YourSaySan/internal/voice"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func SummonHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("summon")

	guildID := i.GuildID
	userID := i.Member.User.ID

	// ユーザーのVC接続を取得
	vs, err := s.State.VoiceState(guildID, userID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "VCに接続していないため、Botを参加させることができません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	channelID := vs.ChannelID

	// 既存の接続を確認
	existingConn, err := b.GetVoiceConnection(guildID)
	if err == nil {
		// 既に接続している場合は、同じチャンネルかチェック
		if existingConn.GetChannelID() == channelID {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "既にこのVCに接続しています。",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		// 別のチャンネルに接続している場合は切断
		existingConn.Leave()
		b.RemoveVoiceConnection(guildID)
	}

	// 新しい接続を作成
	conn, err := voice.NewConnection(s, 50) // 最大キューサイズ: 50
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
		logrus.WithError(err).Error("Failed to join voice channel")
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("VCへの接続に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Botに接続を登録
	b.SetVoiceConnection(guildID, conn)

	// 読み上げ対象チャンネルを追加
	b.GetState().AddTextChannel(guildID, i.ChannelID)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("VC <#%s> に参加しました。", channelID),
		},
	})
}
