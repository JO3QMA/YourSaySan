// internal/domain/entity/errors.go
package entity

import "errors"

var (
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrInvalidSpeakerID = errors.New("invalid speaker ID")
	ErrEmptyMessage     = errors.New("message text cannot be empty")
	ErrMessageTooLong   = errors.New("message text is too long")
)

