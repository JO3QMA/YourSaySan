package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const speakersPerPage = 20

func SpeakerListHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("speaker_list")

	options := i.ApplicationCommandData().Options
	page := 1
	if len(options) > 0 {
		page = int(options[0].IntValue())
		if page < 1 {
			page = 1
		}
	}

	ctx := b.GetContext()
	userID := i.Member.User.ID

	// 話者一覧を取得
	speakers, err := b.GetSpeakerManager().GetAvailableSpeakers(ctx)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("話者一覧の取得に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 現在のユーザーの話者設定を取得
	currentSpeakerID, _ := b.GetSpeakerManager().GetSpeaker(ctx, userID)

	// すべてのスタイルをフラット化
	type StyleInfo struct {
		SpeakerName string
		StyleName   string
		StyleID     int
	}

	var allStyles []StyleInfo
	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			allStyles = append(allStyles, StyleInfo{
				SpeakerName: speaker.Name,
				StyleName:   style.Name,
				StyleID:     style.ID,
			})
		}
	}

	// ページネーション
	totalPages := (len(allStyles) + speakersPerPage - 1) / speakersPerPage
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * speakersPerPage
	end := start + speakersPerPage
	if end > len(allStyles) {
		end = len(allStyles)
	}

	// Embedを作成
	var fields []*discordgo.MessageEmbedField
	for i := start; i < end; i++ {
		style := allStyles[i]
		marker := ""
		if style.StyleID == currentSpeakerID {
			marker = " ▶"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (%s)%s", style.SpeakerName, style.StyleName, marker),
			Value:  fmt.Sprintf("ID: %d", style.StyleID),
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "利用可能な話者一覧",
		Description: fmt.Sprintf("ページ %d / %d (全 %d 件)", page, totalPages, len(allStyles)),
		Fields:      fields,
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "▶ マークは現在の設定です",
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

