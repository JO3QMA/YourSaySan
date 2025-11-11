// internal/domain/entity/speaker.go
package entity

// Speaker represents a voice speaker configuration for a user
type Speaker struct {
	UserID    string
	SpeakerID int
	Name      string
}

// NewSpeaker creates a new Speaker instance
func NewSpeaker(userID string, speakerID int, name string) *Speaker {
	return &Speaker{
		UserID:    userID,
		SpeakerID: speakerID,
		Name:      name,
	}
}

// Validate checks if the speaker configuration is valid
func (s *Speaker) Validate() error {
	if s.UserID == "" {
		return ErrInvalidUserID
	}
	if s.SpeakerID < 0 {
		return ErrInvalidSpeakerID
	}
	return nil
}

