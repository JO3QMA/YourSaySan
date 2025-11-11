// internal/interface/discord/handler_ready.go
package discord

import (
	"github.com/bwmarrin/discordgo"
)

// handleReady handles the ready event
func (b *Bot) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	b.logger.Infof("Bot is ready! Logged in as %s#%s", r.User.Username, r.User.Discriminator)

	// Set the bot status
	if err := s.UpdateGameStatus(0, b.config.Bot.Status); err != nil {
		b.logger.Errorf("Failed to update game status: %v", err)
	}

	// Register slash commands
	if err := b.registerSlashCommands(s); err != nil {
		b.logger.Errorf("Failed to register slash commands: %v", err)
		return
	}

	b.logger.Info("Slash commands registered successfully")
}

// registerSlashCommands registers all slash commands
func (b *Bot) registerSlashCommands(s *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Botの応答速度を確認します",
		},
		{
			Name:        "summon",
			Description: "Botをボイスチャンネルに召喚します",
		},
		{
			Name:        "bye",
			Description: "Botをボイスチャンネルから退出させます",
		},
		{
			Name:        "stop",
			Description: "現在再生中の音声を停止します",
		},
		{
			Name:        "speaker",
			Description: "話者を設定します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "話者ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "speaker_list",
			Description: "利用可能な話者の一覧を表示します",
		},
		{
			Name:        "help",
			Description: "ヘルプを表示します",
		},
		{
			Name:        "invite",
			Description: "Botの招待リンクを表示します",
		},
		{
			Name:        "reconnect",
			Description: "ボイスチャンネルに再接続します",
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

