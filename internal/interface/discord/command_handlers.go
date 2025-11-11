// internal/interface/discord/command_handlers.go
package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/yoursaysan/discord-voicevox-bot/internal/interface/discord/commands"
	pkgerrors "github.com/yoursaysan/discord-voicevox-bot/pkg/errors"
)

// handlePingCommand handles the /ping command
func (b *Bot) handlePingCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	message := commands.HandlePing(s, i)
	b.respondWithMessage(s, i, message)
}

// handleHelpCommand handles the /help command
func (b *Bot) handleHelpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := commands.HandleHelp()
	b.respondWithEmbed(s, i, embed)
}

// handleInviteCommand handles the /invite command
func (b *Bot) handleInviteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	message := commands.HandleInvite(b.config.Bot.ClientID)
	b.respondWithMessage(s, i, message)
}

// handleSummonCommand handles the /summon command
func (b *Bot) handleSummonCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get the guild and user
	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		b.respondWithError(s, i, "サーバー情報の取得に失敗しました")
		return
	}

	// Find the user's voice channel
	var voiceChannelID string
	for _, vs := range guild.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			voiceChannelID = vs.ChannelID
			break
		}
	}

	if voiceChannelID == "" {
		b.respondWithError(s, i, "ボイスチャンネルに参加してください")
		return
	}

	// Join the voice channel
	vc, err := s.ChannelVoiceJoin(i.GuildID, voiceChannelID, false, true)
	if err != nil {
		b.logger.Errorf("Failed to join voice channel: %v", err)
		b.respondWithError(s, i, "ボイスチャンネルへの参加に失敗しました")
		return
	}

	b.SetVoiceConnection(i.GuildID, vc)
	b.respondWithMessage(s, i, "✅ ボイスチャンネルに参加しました！")
}

// handleByeCommand handles the /bye command
func (b *Bot) handleByeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	vc, ok := b.GetVoiceConnection(i.GuildID)
	if !ok {
		b.respondWithError(s, i, "Botはボイスチャンネルに参加していません")
		return
	}

	vc.Disconnect()
	b.RemoveVoiceConnection(i.GuildID)
	b.respondWithMessage(s, i, "👋 ボイスチャンネルから退出しました")
}

// handleStopCommand handles the /stop command
func (b *Bot) handleStopCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	vc, ok := b.GetVoiceConnection(i.GuildID)
	if !ok {
		b.respondWithError(s, i, "Botはボイスチャンネルに参加していません")
		return
	}

	vc.Speaking(false)
	b.respondWithMessage(s, i, "⏹️ 音声を停止しました")
}

// handleSpeakerCommand handles the /speaker command
func (b *Bot) handleSpeakerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		b.respondWithError(s, i, "話者IDを指定してください")
		return
	}

	speakerID := int(options[0].IntValue())
	ctx := context.Background()

	// Validate speaker ID
	if err := b.speakerUseCase.ValidateSpeakerID(ctx, speakerID); err != nil {
		if err == pkgerrors.ErrSpeakerNotFound {
			b.respondWithError(s, i, "指定された話者IDは存在しません")
			return
		}
		b.logger.Errorf("Failed to validate speaker ID: %v", err)
		b.respondWithError(s, i, "話者の検証に失敗しました")
		return
	}

	// Get speaker name
	speakers, err := b.speakerUseCase.GetAvailableSpeakers(ctx)
	if err != nil {
		b.logger.Errorf("Failed to get speakers: %v", err)
		b.respondWithError(s, i, "話者の取得に失敗しました")
		return
	}

	var speakerName string
	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			if style.ID == speakerID {
				speakerName = fmt.Sprintf("%s（%s）", speaker.Name, style.Name)
				break
			}
		}
		if speakerName != "" {
			break
		}
	}

	// Set speaker
	if err := b.speakerUseCase.SetSpeaker(ctx, i.Member.User.ID, speakerID, speakerName); err != nil {
		b.logger.Errorf("Failed to set speaker: %v", err)
		b.respondWithError(s, i, "話者の設定に失敗しました")
		return
	}

	b.respondWithMessage(s, i, fmt.Sprintf("✅ 話者を **%s** に設定しました", speakerName))
}

// handleSpeakerListCommand handles the /speaker_list command
func (b *Bot) handleSpeakerListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	speakers, err := b.speakerUseCase.GetAvailableSpeakers(ctx)
	if err != nil {
		b.logger.Errorf("Failed to get speakers: %v", err)
		b.respondWithError(s, i, "話者の取得に失敗しました")
		return
	}

	var builder strings.Builder
	builder.WriteString("**利用可能な話者一覧:**\n\n")

	for _, speaker := range speakers {
		builder.WriteString(fmt.Sprintf("**%s**\n", speaker.Name))
		for _, style := range speaker.Styles {
			builder.WriteString(fmt.Sprintf("  • ID: `%d` - %s\n", style.ID, style.Name))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("使用方法: `/speaker <ID>`")

	// Discord message limit is 2000 characters
	message := builder.String()
	if len(message) > 2000 {
		message = message[:1997] + "..."
	}

	b.respondWithMessage(s, i, message)
}

// handleReconnectCommand handles the /reconnect command
func (b *Bot) handleReconnectCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if bot is in a voice channel
	vc, ok := b.GetVoiceConnection(i.GuildID)
	if !ok {
		b.respondWithError(s, i, "Botはボイスチャンネルに参加していません")
		return
	}

	// Disconnect and reconnect
	channelID := vc.ChannelID
	vc.Disconnect()
	b.RemoveVoiceConnection(i.GuildID)

	// Rejoin the voice channel
	newVc, err := s.ChannelVoiceJoin(i.GuildID, channelID, false, true)
	if err != nil {
		b.logger.Errorf("Failed to reconnect to voice channel: %v", err)
		b.respondWithError(s, i, "ボイスチャンネルへの再接続に失敗しました")
		return
	}

	b.SetVoiceConnection(i.GuildID, newVc)
	b.respondWithMessage(s, i, "🔄 ボイスチャンネルに再接続しました")
}

