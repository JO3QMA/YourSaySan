package events

import (
	"context"
	"strings"
	"time"

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

		// 4. 読み上げ対象チャンネルかチェック
		// summonで呼び出されていないチャンネルのメッセージはログに出さない
		if !b.GetState().IsTextChannelActive(m.GuildID, m.ChannelID) {
			return
		}

		// 読み上げ対象チャンネルのメッセージをログに記録
		// Debugレベルではメッセージ内容も記録
		logrus.WithFields(logrus.Fields{
			"guild_id":    m.GuildID,
			"channel_id":  m.ChannelID,
			"user_id":     m.Author.ID,
			"message_id":  m.ID,
			"content_len": len(m.Content),
			"content":     m.Content, // Debugレベルでメッセージ内容を記録
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
		config := b.GetConfig()
		transformedText := utils.TransformMessage(m.Content, config.GetVoiceVoxMaxMessageLength())

		if transformedText == "" {
			logrus.WithFields(logrus.Fields{
				"guild_id":   m.GuildID,
				"channel_id": m.ChannelID,
				"user_id":    m.Author.ID,
			}).Trace("Message transformed to empty string, skipping")
			return
		}

		logrus.WithFields(logrus.Fields{
			"guild_id":         m.GuildID,
			"user_id":          m.Author.ID,
			"original_len":     len(m.Content),
			"transformed_len":  len(transformedText),
			"transformed_text": transformedText, // Debugレベルで変換後のテキストも記録
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
