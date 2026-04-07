package events

import (
	"context"
	"strings"
	"time"

	"github.com/JO3QMA/YourSaySan/internal/senryu"
	"github.com/JO3QMA/YourSaySan/pkg/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func MessageCreateHandler(b BotInterface) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// 1. Botのメッセージはスキップ
		if m.Author == nil || m.Author.Bot {
			return
		}

		// 2. 本文が空のメッセージはスキップ
		if len(m.Content) == 0 {
			return
		}

		// 3. スラッシュコマンドはスキップ
		if strings.HasPrefix(m.Content, "/") {
			return
		}

		cfg := b.GetConfig()

		// 川柳（5-7-5）: ギルド内の全チャンネルが対象（DM は GuildID なしのため除外）
		// 経路A: ちょうど3行 / 経路B: 改行なしで空白除去後の1文が17モーラ
		if cfg.GetSenryuEnabled() && m.GuildID != "" {
			var senryuThreeLines []string
			var senryuBlob string
			senryuMode := 0 // 0=なし, 1=三行, 2=連続
			if lines, ok := senryu.ThreeLines(m.Content); ok {
				senryuMode = 1
				senryuThreeLines = lines
			} else if blob, ok := senryu.IsUnbrokenSenryuCandidate(m.Content); ok {
				senryuMode = 2
				senryuBlob = blob
			}
			if senryuMode != 0 {
				channelID := m.ChannelID
				messageID := m.ID
				guildID := m.GuildID
				authorID := m.Author.ID
				replyText := cfg.GetSenryuReplyText()
				session := b.GetSession()
				mode := senryuMode
				linesCopy := senryuThreeLines
				blobCopy := senryuBlob
				b.RunWithSemaphore(func() {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					speakerID, err := b.GetSpeakerManager().GetSpeaker(ctx, authorID)
					if err != nil {
						logrus.WithError(err).WithField("user_id", authorID).Warn("Failed to get speaker for senryu check")
						speakerID = 2
					}

					vv := b.GetVoiceVox()
					var is575 bool
					if mode == 1 {
						is575, err = senryu.Is575Morae(ctx, vv, linesCopy, speakerID)
					} else {
						is575, err = senryu.Is575MoraeUnbroken(ctx, vv, blobCopy, speakerID)
					}
					if err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"guild_id":   guildID,
							"channel_id": channelID,
							"message_id": messageID,
						}).Warn("Senryu mora check failed")
						return
					}
					if !is575 {
						return
					}

					ref := &discordgo.MessageReference{
						MessageID: messageID,
						ChannelID: channelID,
						GuildID:   guildID,
					}
					if _, err := session.ChannelMessageSendReply(channelID, replyText, ref); err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"guild_id":   guildID,
							"channel_id": channelID,
							"message_id": messageID,
						}).Warn("Failed to send senryu reply")
					}
				})
			}
		}

		// 4. 読み上げ対象チャンネルかチェック
		if !b.GetState().IsTextChannelActive(m.GuildID, m.ChannelID) {
			return
		}

		// 読み上げ対象チャンネルのメッセージをログに記録（本文はプライバシー保護のため記録しない）
		logrus.WithFields(logrus.Fields{
			"guild_id":    m.GuildID,
			"channel_id":  m.ChannelID,
			"user_id":     m.Author.ID,
			"message_id":  m.ID,
			"content_len": len(m.Content),
		}).Debug("Message received for text-to-speech")

		// 5. VC接続を確認
		conn, err := b.GetVoiceConnection(m.GuildID)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"guild_id": m.GuildID,
			}).Trace("No voice connection found for guild")
			return
		}

		// 6. メッセージ変換
		transformedText := utils.TransformMessage(m.Content, cfg.GetVoiceVoxMaxMessageLength())

		if transformedText == "" {
			logrus.WithFields(logrus.Fields{
				"guild_id":   m.GuildID,
				"channel_id": m.ChannelID,
				"user_id":    m.Author.ID,
			}).Trace("Message transformed to empty string, skipping")
			return
		}

		logrus.WithFields(logrus.Fields{
			"guild_id":        m.GuildID,
			"user_id":         m.Author.ID,
			"original_len":    len(m.Content),
			"transformed_len": len(transformedText),
		}).Debug("Message transformed for TTS")

		// 7. 話者設定取得
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		speakerID, err := b.GetSpeakerManager().GetSpeaker(ctx, m.Author.ID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", m.Author.ID).Warn("Failed to get speaker")
			speakerID = 2 // デフォルト値
		}

		logrus.WithFields(logrus.Fields{
			"guild_id":   m.GuildID,
			"user_id":    m.Author.ID,
			"speaker_id": speakerID,
		}).Trace("Speaker ID retrieved")

		// 8. 音声生成
		startTime := time.Now()
		audioData, err := b.GetVoiceVox().Speak(ctx, transformedText, speakerID)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":    m.Author.ID,
				"speaker_id": speakerID,
				"text_len":   len(transformedText),
			}).Error("Failed to generate audio")
			return
		}

		// メトリクス記録
		duration := time.Since(startTime).Seconds()
		b.RecordAudioGenerationDuration(speakerID, duration)

		logrus.WithFields(logrus.Fields{
			"guild_id":     m.GuildID,
			"user_id":      m.Author.ID,
			"speaker_id":   speakerID,
			"audio_size":   len(audioData),
			"duration_sec": duration,
		}).Debug("Audio generated successfully")

		// 9. 音声再生
		if err := conn.Play(ctx, audioData); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"guild_id": m.GuildID,
			}).Error("Failed to play audio")
			return
		}

		queueSize := conn.QueueSize()
		logrus.WithFields(logrus.Fields{
			"guild_id":   m.GuildID,
			"queue_size": queueSize,
		}).Trace("Audio queued for playback")

		// キューサイズを更新
		b.SetQueueSize(m.GuildID, queueSize)
	}
}
