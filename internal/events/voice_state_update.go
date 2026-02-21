package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func VoiceStateUpdateHandler(b BotInterface) func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		// Bot自身のVC接続を取得
		guild, err := s.Guild(vs.GuildID)
		if err != nil {
			return
		}

		// BotのVC接続を確認
		botUserID := s.State.User.ID
		var botVoiceState *discordgo.VoiceState
		for _, voiceState := range guild.VoiceStates {
			if voiceState.UserID == botUserID {
				botVoiceState = voiceState
				break
			}
		}
		if botVoiceState == nil {
			return // BotはVCに接続していない
		}

		// 同じVCチャンネルのメンバーをカウント
		memberCount := 0
		for _, voiceState := range guild.VoiceStates {
			if voiceState.ChannelID == botVoiceState.ChannelID {
				// Bot以外のメンバーをカウント
				if voiceState.UserID != botUserID {
					memberCount++
				}
			}
		}

		// Bot以外のメンバーが0人の場合、切断
		if memberCount == 0 {
			conn, err := b.GetVoiceConnection(vs.GuildID)
			if err == nil {
				logrus.WithField("guild_id", vs.GuildID).Info("Auto-disconnecting from empty voice channel")
				if err := conn.Leave(); err != nil {
					logrus.WithError(err).Error("Failed to leave voice channel")
				}
				b.RemoveVoiceConnection(vs.GuildID)
			}
		}
	}
}

