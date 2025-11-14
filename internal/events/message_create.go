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
		if m.Author.Bot {
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
		if !b.GetState().IsTextChannelActive(m.GuildID, m.ChannelID) {
			return
		}

		// 5. VC接続を確認
		conn, err := b.GetVoiceConnection(m.GuildID)
		if err != nil {
			return
		}

		// 6. メッセージ変換
		config := b.GetConfig()
		transformedText := utils.TransformMessage(m.Content, config.GetVoiceVoxMaxMessageLength())

		if transformedText == "" {
			return
		}

		// 7. 話者設定取得
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		speakerID, err := b.GetSpeakerManager().GetSpeaker(ctx, m.Author.ID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", m.Author.ID).Warn("Failed to get speaker")
			speakerID = 2 // デフォルト値
		}

		// 8. 音声生成
		startTime := time.Now()
		audioData, err := b.GetVoiceVox().Speak(ctx, transformedText, speakerID)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":    m.Author.ID,
				"speaker_id": speakerID,
				"text":       transformedText,
			}).Error("Failed to generate audio")
			return
		}

		// メトリクス記録
		duration := time.Since(startTime).Seconds()
		b.RecordAudioGenerationDuration(speakerID, duration)

		// 9. 音声再生
		if err := conn.Play(ctx, audioData); err != nil {
			logrus.WithError(err).Error("Failed to play audio")
			return
		}

		// キューサイズを更新
		b.SetQueueSize(m.GuildID, conn.QueueSize())
	}
}
