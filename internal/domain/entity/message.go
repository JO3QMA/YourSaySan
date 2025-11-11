package entity

// Message represents a text message entity
type Message struct {
	Text     string `json:"text"`
	UserID   string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	GuildID  string `json:"guild_id"`
}

// IsValid checks if the message is valid
func (m *Message) IsValid() bool {
	return m.Text != "" && m.UserID != "" && m.ChannelID != "" && m.GuildID != ""
}
