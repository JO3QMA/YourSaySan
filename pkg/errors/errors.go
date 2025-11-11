// pkg/errors/errors.go
package errors

import "errors"

var (
	// Repository errors
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input")
	
	// Voice errors
	ErrVoiceGenerationFailed = errors.New("voice generation failed")
	ErrSpeakerNotFound       = errors.New("speaker not found")
	
	// Discord errors
	ErrNotInVoiceChannel = errors.New("user not in voice channel")
	ErrBotNotInChannel   = errors.New("bot not in voice channel")
	ErrVoiceConnectionFailed = errors.New("failed to establish voice connection")
)

