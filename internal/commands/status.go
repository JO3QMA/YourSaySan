package commands

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
)

var startTime = time.Now()

func StatusHandler(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	IncrementCommandCounter("status")

	// オーナーチェック
	config := b.GetConfig()
	userID := i.Member.User.ID
	if userID != config.GetBotOwnerID() {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "このコマンドはBotオーナーのみ実行可能です。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	ctx := b.GetContext()

	// 稼働時間
	uptime := time.Since(startTime)

	// 接続中のギルド数
	state := b.GetState()
	guildCount := state.GetGuildCount()

	// アクティブなVC接続数
	activeConnections := b.GetActiveVoiceConnections()

	// 音声キューの合計サイズ
	totalQueueSize := b.GetTotalQueueSize()

	// メモリ使用量
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsageMB := float64(m.Alloc) / 1024 / 1024

	// VoiceVox APIの状態
	voicevoxHealthy := false
	if b.GetVoiceVox() != nil {
		_, err := b.GetVoiceVox().GetSpeakers(ctx)
		voicevoxHealthy = err == nil
	}

	// Redis接続状態
	redisHealthy := b.GetSpeakerManager() != nil

	// Embedを作成
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "稼働時間",
			Value:  formatDuration(uptime),
			Inline: true,
		},
		{
			Name:   "接続中のギルド数",
			Value:  fmt.Sprintf("%d", guildCount),
			Inline: true,
		},
		{
			Name:   "アクティブなVC接続数",
			Value:  fmt.Sprintf("%d", activeConnections),
			Inline: true,
		},
		{
			Name:   "音声キューの合計サイズ",
			Value:  fmt.Sprintf("%d", totalQueueSize),
			Inline: true,
		},
		{
			Name:   "メモリ使用量",
			Value:  fmt.Sprintf("%.2f MB", memUsageMB),
			Inline: true,
		},
		{
			Name:   "VoiceVox API",
			Value:  formatHealth(voicevoxHealthy),
			Inline: true,
		},
		{
			Name:   "Redis接続",
			Value:  formatHealth(redisHealthy),
			Inline: true,
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:  "Bot状態情報",
		Fields: fields,
		Color:  0x00ff00,
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%d日 %d時間 %d分", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d時間 %d分 %d秒", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分 %d秒", minutes, seconds)
	}
	return fmt.Sprintf("%d秒", seconds)
}

func formatHealth(healthy bool) string {
	if healthy {
		return "✅ 正常"
	}
	return "❌ 異常"
}

