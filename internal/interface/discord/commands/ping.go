// internal/interface/discord/commands/ping.go
package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandlePing handles the ping command
func HandlePing(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	// Calculate the latency
	latency := s.HeartbeatLatency().Milliseconds()
	
	return fmt.Sprintf("🏓 Pong! レイテンシ: %dms", latency)
}

