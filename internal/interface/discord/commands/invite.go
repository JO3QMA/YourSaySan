// internal/interface/discord/commands/invite.go
package commands

import (
	"fmt"
)

// HandleInvite handles the invite command
func HandleInvite(clientID string) string {
	inviteURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&permissions=3146752&scope=bot%%20applications.commands",
		clientID,
	)
	return fmt.Sprintf("🔗 Botを招待するにはこちらのリンクをクリックしてください:\n%s", inviteURL)
}

