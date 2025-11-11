// internal/interface/discord/handler_voice_state.go
package discord

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// handleVoiceStateUpdate handles voice state updates
func (b *Bot) handleVoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	// Get the voice connection for this guild
	vc, ok := b.GetVoiceConnection(vsu.GuildID)
	if !ok {
		return
	}

	// Get the channel the bot is in
	botChannelID := vc.ChannelID

	// Count users in the bot's voice channel
	guild, err := s.State.Guild(vsu.GuildID)
	if err != nil {
		b.logger.Errorf("Failed to get guild: %v", err)
		return
	}

	userCount := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == botChannelID && !vs.Member.User.Bot {
			userCount++
		}
	}

	// If no users left, schedule auto-disconnect
	if userCount == 0 {
		b.logger.Infof("No users left in voice channel %s, scheduling auto-disconnect", botChannelID)
		go b.scheduleAutoDisconnect(s, vsu.GuildID, botChannelID, 5*time.Minute)
	}
}

// scheduleAutoDisconnect disconnects the bot after a delay if no users join
func (b *Bot) scheduleAutoDisconnect(s *discordgo.Session, guildID, channelID string, delay time.Duration) {
	time.Sleep(delay)

	// Check if voice connection still exists
	vc, ok := b.GetVoiceConnection(guildID)
	if !ok {
		return
	}

	// Check if the channel is still the same
	if vc.ChannelID != channelID {
		return
	}

	// Count users again
	guild, err := s.State.Guild(guildID)
	if err != nil {
		b.logger.Errorf("Failed to get guild: %v", err)
		return
	}

	userCount := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID && !vs.Member.User.Bot {
			userCount++
		}
	}

	// If still no users, disconnect
	if userCount == 0 {
		b.logger.Infof("Auto-disconnecting from voice channel %s due to inactivity", channelID)
		vc.Disconnect()
		b.RemoveVoiceConnection(guildID)
	}
}

