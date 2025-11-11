// internal/interface/redis/repository.go
package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/yoursaysan/discord-voicevox-bot/internal/domain/entity"
	"github.com/yoursaysan/discord-voicevox-bot/internal/domain/repository"
	pkgerrors "github.com/yoursaysan/discord-voicevox-bot/pkg/errors"
)

// SpeakerRepository implements the repository.SpeakerRepository interface using Redis
type SpeakerRepository struct {
	client *redis.Client
	prefix string
}

// NewSpeakerRepository creates a new Redis-based speaker repository
func NewSpeakerRepository(client *redis.Client, prefix string) repository.SpeakerRepository {
	if prefix == "" {
		prefix = "speaker"
	}
	return &SpeakerRepository{
		client: client,
		prefix: prefix,
	}
}

// key generates the Redis key for a user
func (r *SpeakerRepository) key(userID string) string {
	return fmt.Sprintf("%s:%s", r.prefix, userID)
}

// Get retrieves the speaker configuration for a user
func (r *SpeakerRepository) Get(ctx context.Context, userID string) (*entity.Speaker, error) {
	key := r.key(userID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get speaker from redis: %w", err)
	}

	var speaker entity.Speaker
	if err := json.Unmarshal(data, &speaker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal speaker data: %w", err)
	}

	return &speaker, nil
}

// Set saves the speaker configuration for a user
func (r *SpeakerRepository) Set(ctx context.Context, speaker *entity.Speaker) error {
	if err := speaker.Validate(); err != nil {
		return fmt.Errorf("invalid speaker: %w", err)
	}

	key := r.key(speaker.UserID)
	data, err := json.Marshal(speaker)
	if err != nil {
		return fmt.Errorf("failed to marshal speaker data: %w", err)
	}

	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set speaker in redis: %w", err)
	}

	return nil
}

// Delete removes the speaker configuration for a user
func (r *SpeakerRepository) Delete(ctx context.Context, userID string) error {
	key := r.key(userID)
	result, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete speaker from redis: %w", err)
	}

	if result == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}

// Exists checks if a speaker configuration exists for a user
func (r *SpeakerRepository) Exists(ctx context.Context, userID string) (bool, error) {
	key := r.key(userID)
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check speaker existence in redis: %w", err)
	}

	return result > 0, nil
}

// NewRedisClient creates a new Redis client
func NewRedisClient(host string, port int, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
		DB:   db,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}

