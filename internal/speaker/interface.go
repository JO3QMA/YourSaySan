package speaker

import (
	"context"
	"time"

	"github.com/JO3QMA/YourSaySan/internal/voicevox"
	"github.com/redis/go-redis/v9"
)

// RedisClient はRedisクライアントのインターフェース
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

// VoiceVoxAPI はVoiceVoxクライアントのインターフェース
type VoiceVoxAPI interface {
	GetSpeakers(ctx context.Context) ([]voicevox.Speaker, error)
}
