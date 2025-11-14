package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func DisconnectHandler(s *discordgo.Session, event *discordgo.Disconnect) {
	logrus.Info("WebSocket disconnected, reconnecting...")
	// discordgoの自動再接続機能に任せる
}

