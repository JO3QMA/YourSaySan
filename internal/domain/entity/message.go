// internal/domain/entity/message.go
package entity

import "time"

// Message represents a text message to be converted to speech
type Message struct {
	ID        string
	Text      string
	UserID    string
	GuildID   string
	ChannelID string
	CreatedAt time.Time
}

// NewMessage creates a new Message instance
func NewMessage(id, text, userID, guildID, channelID string) *Message {
	return &Message{
		ID:        id,
		Text:      text,
		UserID:    userID,
		GuildID:   guildID,
		ChannelID: channelID,
		CreatedAt: time.Now(),
	}
}

// Validate checks if the message is valid
func (m *Message) Validate() error {
	if m.Text == "" {
		return ErrEmptyMessage
	}
	if m.UserID == "" {
		return ErrInvalidUserID
	}
	return nil
}

