package speaker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/JO3QMA/YourSaySan/internal/voicevox"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	defaultSpeakerID = 2
)

type cacheEntry struct {
	speakerID int
	expires   time.Time
}

type Manager struct {
	redis    RedisClient
	voicevox VoiceVoxAPI

	// メモリキャッシュ（LRUキャッシュ）
	cache        *lru.Cache[string, *cacheEntry]
	cacheTTL     time.Duration // キャッシュTTL: 5分
	maxCacheSize int           // 最大キャッシュサイズ: 1000件

	// 話者一覧キャッシュ
	speakersCache     []voicevox.Speaker
	speakersCacheTime time.Time
	speakersCacheTTL  time.Duration // 話者一覧キャッシュTTL: 1時間
	speakersCacheMu   sync.RWMutex
}

func NewManager(redisClient RedisClient, voicevoxAPI VoiceVoxAPI) (*Manager, error) {
	cache, err := lru.New[string, *cacheEntry](1000)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}

	m := &Manager{
		redis:            redisClient,
		voicevox:         voicevoxAPI,
		cache:            cache,
		cacheTTL:         5 * time.Minute,
		maxCacheSize:     1000,
		speakersCacheTTL: 1 * time.Hour,
	}

	// Redis再接続ループを開始
	go m.reconnectLoop(context.Background())

	return m, nil
}

func (m *Manager) GetSpeaker(ctx context.Context, userID string) (int, error) {
	// キャッシュから取得を試みる
	if entry, ok := m.cache.Get(userID); ok {
		if time.Now().Before(entry.expires) {
			return entry.speakerID, nil
		}
		// 期限切れの場合はキャッシュから削除
		m.cache.Remove(userID)
	}

	// Redisから取得
	key := fmt.Sprintf("speaker:%s", userID)
	cmd := m.redis.Get(ctx, key)
	val, err := cmd.Result()
	if err == redis.Nil {
		// キーが存在しない場合はデフォルト値を返す
		defaultEntry := &cacheEntry{
			speakerID: defaultSpeakerID,
			expires:   time.Now().Add(m.cacheTTL),
		}
		m.cache.Add(userID, defaultEntry)
		return defaultSpeakerID, nil
	}
	if err != nil {
		// Redisエラー時はデフォルト値を使用
		logrus.WithError(err).WithField("user_id", userID).Warn("Failed to get speaker from Redis, using default")
		return defaultSpeakerID, nil
	}

	// 文字列を整数に変換
	speakerID, err := strconv.Atoi(val)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Warn("Invalid speaker ID in Redis, using default")
		return defaultSpeakerID, nil
	}

	// キャッシュに保存
	entry := &cacheEntry{
		speakerID: speakerID,
		expires:   time.Now().Add(m.cacheTTL),
	}
	m.cache.Add(userID, entry)

	return speakerID, nil
}

func (m *Manager) SetSpeaker(ctx context.Context, userID string, speakerID int) error {
	key := fmt.Sprintf("speaker:%s", userID)
	if err := m.redis.Set(ctx, key, speakerID, 0).Err(); err != nil {
		return fmt.Errorf("failed to set speaker in Redis: %w", err)
	}

	// キャッシュを無効化
	m.invalidateCache(userID)

	// 新しい値をキャッシュに保存
	entry := &cacheEntry{
		speakerID: speakerID,
		expires:   time.Now().Add(m.cacheTTL),
	}
	m.cache.Add(userID, entry)

	return nil
}

func (m *Manager) GetAvailableSpeakers(ctx context.Context) ([]voicevox.Speaker, error) {
	m.speakersCacheMu.RLock()
	if len(m.speakersCache) > 0 && time.Since(m.speakersCacheTime) < m.speakersCacheTTL {
		speakers := m.speakersCache
		m.speakersCacheMu.RUnlock()
		return speakers, nil
	}
	m.speakersCacheMu.RUnlock()

	// キャッシュが期限切れまたは空の場合はVoiceVoxから取得
	speakers, err := m.voicevox.GetSpeakers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get speakers from VoiceVox: %w", err)
	}

	// キャッシュを更新
	m.speakersCacheMu.Lock()
	m.speakersCache = speakers
	m.speakersCacheTime = time.Now()
	m.speakersCacheMu.Unlock()

	return speakers, nil
}

func (m *Manager) ValidSpeaker(ctx context.Context, speakerID int) (bool, error) {
	speakers, err := m.GetAvailableSpeakers(ctx)
	if err != nil {
		return false, err
	}

	for _, speaker := range speakers {
		for _, style := range speaker.Styles {
			if style.ID == speakerID {
				return true, nil
			}
		}
	}

	return false, nil
}

func (m *Manager) invalidateCache(userID string) {
	m.cache.Remove(userID)
}

func (m *Manager) reconnectLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.redis.Ping(ctx).Err(); err != nil {
				logrus.WithError(err).Warn("Redis still unavailable")
				continue
			}
			// 再接続成功
			logrus.Info("Redis reconnected, clearing cache")
			m.cache.Purge() // キャッシュをクリア
		}
	}
}
