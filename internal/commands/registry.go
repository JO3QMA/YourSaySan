package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type CommandInfo struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
}

type CommandHandler func(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) error

type Registry struct {
	bot      BotInterface
	commands map[string]CommandHandler
	infos    map[string]CommandInfo
}

func NewRegistry(b BotInterface) *Registry {
	return &Registry{
		bot:      b,
		commands: make(map[string]CommandHandler),
		infos:    make(map[string]CommandInfo),
	}
}

func (r *Registry) Register(name string, info CommandInfo, handler CommandHandler) {
	r.commands[name] = handler
	r.infos[name] = info
}

func (r *Registry) RegisterAll(s *discordgo.Session) error {
	commands := make([]*discordgo.ApplicationCommand, 0, len(r.infos))

	for name, info := range r.infos {
		cmd := &discordgo.ApplicationCommand{
			Name:        name,
			Description: info.Description,
			Options:     info.Options,
		}
		commands = append(commands, cmd)
	}

	// グローバルコマンドとして登録
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("cannot create '%s' command: %w", cmd.Name, err)
		}
	}

	return nil
}

func (r *Registry) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name
	if commandName == "" {
		return
	}

	// コマンド情報をログに記録
	logrus.WithFields(logrus.Fields{
		"command":    commandName,
		"guild_id":   i.GuildID,
		"user_id":    i.Member.User.ID,
		"channel_id": i.ChannelID,
	}).Debug("Command received")

	handler, exists := r.commands[commandName]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"command":  commandName,
			"guild_id": i.GuildID,
		}).Warn("Unknown command received")
		return
	}

	if err := handler(r.bot, s, i); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"command":  commandName,
			"guild_id": i.GuildID,
			"user_id":  i.Member.User.ID,
		}).Error("Command handler error")
		// エラーハンドリング
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("エラーが発生しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	} else {
		logrus.WithFields(logrus.Fields{
			"command":  commandName,
			"guild_id": i.GuildID,
			"user_id":  i.Member.User.ID,
		}).Debug("Command completed successfully")
	}
}
