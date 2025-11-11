package entity

// Speaker represents a voice speaker entity
type Speaker struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// IsValid checks if the speaker is valid
func (s *Speaker) IsValid() bool {
	return s.ID > 0 && s.Name != ""
}
