package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"discord-voice-bot/internal/domain/entity"
	"discord-voice-bot/internal/domain/repository"

	"github.com/redis/go-redis/v9"
)

// Repository implements SpeakerRepository using Redis
type Repository struct {
	client *redis.Client
}

// NewRepository creates a new Redis repository
func NewRepository(client *redis.Client) repository.SpeakerRepository {
	return &Repository{
		client: client,
	}
}

// GetSpeaker retrieves the speaker setting for a user
func (r *Repository) GetSpeaker(ctx context.Context, userID string) (*entity.Speaker, error) {
	key := r.speakerKey(userID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Speaker not found, return default
			return &entity.Speaker{ID: 1, Name: "Default"}, nil
		}
		return nil, fmt.Errorf("failed to get speaker from redis: %w", err)
	}

	var speaker entity.Speaker
	if err := json.Unmarshal([]byte(val), &speaker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal speaker: %w", err)
	}

	return &speaker, nil
}

// SetSpeaker saves the speaker setting for a user
func (r *Repository) SetSpeaker(ctx context.Context, userID string, speaker *entity.Speaker) error {
	if speaker == nil {
		return fmt.Errorf("speaker cannot be nil")
	}

	if !speaker.IsValid() {
		return fmt.Errorf("invalid speaker: %+v", speaker)
	}

	key := r.speakerKey(userID)

	data, err := json.Marshal(speaker)
	if err != nil {
		return fmt.Errorf("failed to marshal speaker: %w", err)
	}

	// Set with expiration (optional, can be configured)
	err = r.client.Set(ctx, key, data, 30*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set speaker in redis: %w", err)
	}

	return nil
}

// DeleteSpeaker removes the speaker setting for a user
func (r *Repository) DeleteSpeaker(ctx context.Context, userID string) error {
	key := r.speakerKey(userID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete speaker from redis: %w", err)
	}

	return nil
}

// speakerKey generates the Redis key for a user's speaker setting
func (r *Repository) speakerKey(userID string) string {
	return fmt.Sprintf("speaker:%s", userID)
}
