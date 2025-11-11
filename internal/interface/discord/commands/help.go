package commands

import (
	"github.com/bwmarrin/discordgo"
)

// HandleHelp handles the help command
func HandleHelp(b BotInterface, s *discordgo.Session, i *discordgo.InteractionCreate) {
	prefix := b.GetPrefix()

	content := "**Voice Bot Help**\n\n" +
		"**Text-to-Speech:**\n" +
		"Send messages starting with `" + prefix + "` in a text channel while in a voice channel.\n\n" +
		"**Commands:**\n" +
		"`/ping` - Check bot latency\n" +
		"`/summon` - Join your voice channel\n" +
		"`/bye` - Leave voice channel\n" +
		"`/stop` - Stop current audio playback\n" +
		"`/speaker <id>` - Set your speaker voice\n" +
		"`/speaker_list` - Show available speakers\n" +
		"`/help` - Show this help message\n" +
		"`/invite` - Get bot invite link\n" +
		"`/reconnect` - Reconnect to voice channel\n\n" +
		"**Setup:**\n" +
		"1. Use `/summon` to bring the bot to your voice channel\n" +
		"2. Use `/speaker_list` to see available voices\n" +
		"3. Use `/speaker <id>` to choose your voice\n" +
		"4. Start sending messages with the prefix!"

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})

	if err != nil {
		b.GetLogger().Error("Failed to respond to help command: %v", err)
	}
}
