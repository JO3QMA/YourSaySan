package commands

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func SpeakerHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("speaker")

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "話者IDを指定してください。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	speakerID := int(options[0].IntValue())
	userID := i.Member.User.ID
	ctx := b.GetContext()

	// 話者IDの検証
	valid, err := b.GetSpeakerManager().ValidSpeaker(ctx, speakerID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("話者IDの検証に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if !valid {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("無効な話者IDです: %d", speakerID),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 話者設定を保存
	if err := b.GetSpeakerManager().SetSpeaker(ctx, userID, speakerID); err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("話者設定の保存に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 話者名を取得
	speakers, err := b.GetSpeakerManager().GetAvailableSpeakers(ctx)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("話者ID %d に設定しました。", speakerID),
			},
		})
	}

	var speakerName string
	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			if style.ID == speakerID {
				speakerName = fmt.Sprintf("%s (%s)", speaker.Name, style.Name)
				break
			}
		}
		if speakerName != "" {
			break
		}
	}

	if speakerName == "" {
		speakerName = strconv.Itoa(speakerID)
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("話者を %s (ID: %d) に設定しました。", speakerName, speakerID),
		},
	})
}

