package speaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JO3QMA/YourSaySan/internal/voicevox"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- モック定義 ---

type mockRedisClient struct {
	getVal string
	getErr error
	setErr error
}

func (m *mockRedisClient) Get(_ context.Context, _ string) *redis.StringCmd {
	cmd := redis.NewStringCmd(context.Background())
	if m.getErr != nil {
		cmd.SetErr(m.getErr)
	} else {
		cmd.SetVal(m.getVal)
	}
	return cmd
}

func (m *mockRedisClient) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(context.Background())
	if m.setErr != nil {
		cmd.SetErr(m.setErr)
	} else {
		cmd.SetVal("OK")
	}
	return cmd
}

func (m *mockRedisClient) Ping(_ context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(context.Background())
	cmd.SetVal("PONG")
	return cmd
}

type mockVoiceVoxAPI struct {
	speakers []voicevox.Speaker
	err      error
}

func (m *mockVoiceVoxAPI) GetSpeakers(_ context.Context) ([]voicevox.Speaker, error) {
	return m.speakers, m.err
}

// テスト用の話者リスト
var testSpeakers = []voicevox.Speaker{
	{
		Name:        "四国めたん",
		SpeakerUUID: "uuid-1",
		Styles: []voicevox.Style{
			{Name: "ノーマル", ID: 2},
			{Name: "あまあま", ID: 0},
		},
	},
	{
		Name:        "ずんだもん",
		SpeakerUUID: "uuid-2",
		Styles: []voicevox.Style{
			{Name: "ノーマル", ID: 3},
		},
	},
}

func newTestManager(t *testing.T, redisClient RedisClient, vvClient VoiceVoxAPI) *Manager {
	t.Helper()
	m, err := NewManager(redisClient, vvClient)
	require.NoError(t, err)
	return m
}

// --- GetSpeaker テスト ---

func TestManager_GetSpeaker_RedisMiss_ReturnsDefault(t *testing.T) {
	rc := &mockRedisClient{getErr: redis.Nil}
	m := newTestManager(t, rc, &mockVoiceVoxAPI{})

	id, err := m.GetSpeaker(context.Background(), "user1")
	require.NoError(t, err)
	assert.Equal(t, defaultSpeakerID, id)
}

func TestManager_GetSpeaker_RedisHasValue(t *testing.T) {
	rc := &mockRedisClient{getVal: "3"}
	m := newTestManager(t, rc, &mockVoiceVoxAPI{})

	id, err := m.GetSpeaker(context.Background(), "user1")
	require.NoError(t, err)
	assert.Equal(t, 3, id)
}

func TestManager_GetSpeaker_RedisError_ReturnsDefault(t *testing.T) {
	rc := &mockRedisClient{getErr: errors.New("connection refused")}
	m := newTestManager(t, rc, &mockVoiceVoxAPI{})

	// Redisエラー時はデフォルト値にフォールバック（エラーを伝播しない）
	id, err := m.GetSpeaker(context.Background(), "user1")
	require.NoError(t, err)
	assert.Equal(t, defaultSpeakerID, id)
}

func TestManager_GetSpeaker_InvalidValueInRedis_ReturnsDefault(t *testing.T) {
	rc := &mockRedisClient{getVal: "not-a-number"}
	m := newTestManager(t, rc, &mockVoiceVoxAPI{})

	id, err := m.GetSpeaker(context.Background(), "user1")
	require.NoError(t, err)
	assert.Equal(t, defaultSpeakerID, id)
}

func TestManager_GetSpeaker_LRUCacheHit(t *testing.T) {
	callCount := 0
	rc := &mockRedisClient{getVal: "5"}
	// Getの呼び出し回数を追跡するためのラッパー
	type countingRedis struct {
		RedisClient
		count *int
	}
	cr := countingRedis{RedisClient: rc, count: &callCount}

	m := newTestManager(t, cr.RedisClient, &mockVoiceVoxAPI{})

	ctx := context.Background()

	// 1回目: Redisから取得してキャッシュに保存
	id1, err := m.GetSpeaker(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 5, id1)

	// 2回目: キャッシュから取得（Redisへのアクセスなし）
	// モックのRedisClientを差し替えてキャッシュ検証
	m.redis = &mockRedisClient{getErr: errors.New("should not be called")}

	id2, err := m.GetSpeaker(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 5, id2, "2回目はキャッシュから返るべき")
}

// --- SetSpeaker テスト ---

func TestManager_SetSpeaker_Success(t *testing.T) {
	rc := &mockRedisClient{}
	m := newTestManager(t, rc, &mockVoiceVoxAPI{})

	err := m.SetSpeaker(context.Background(), "user1", 3)
	require.NoError(t, err)
}

func TestManager_SetSpeaker_InvalidatesCache(t *testing.T) {
	// キャッシュに値を入れる（Redisから取得）
	rc := &mockRedisClient{getVal: "2"}
	m := newTestManager(t, rc, &mockVoiceVoxAPI{})

	ctx := context.Background()

	// キャッシュウォームアップ
	id, _ := m.GetSpeaker(ctx, "user1")
	assert.Equal(t, 2, id)

	// SetSpeakerでキャッシュを無効化
	err := m.SetSpeaker(ctx, "user1", 5)
	require.NoError(t, err)

	// キャッシュ無効化後は新しいRedis値で返る（モックは5を返すよう設定）
	m.redis = &mockRedisClient{getVal: "5"}
	id, _ = m.GetSpeaker(ctx, "user1")
	assert.Equal(t, 5, id)
}

func TestManager_SetSpeaker_RedisError(t *testing.T) {
	rc := &mockRedisClient{setErr: errors.New("redis write failed")}
	m := newTestManager(t, rc, &mockVoiceVoxAPI{})

	err := m.SetSpeaker(context.Background(), "user1", 3)
	assert.Error(t, err)
}

// --- GetAvailableSpeakers テスト ---

func TestManager_GetAvailableSpeakers_FetchesFromVoiceVox(t *testing.T) {
	vv := &mockVoiceVoxAPI{speakers: testSpeakers}
	m := newTestManager(t, &mockRedisClient{}, vv)

	speakers, err := m.GetAvailableSpeakers(context.Background())
	require.NoError(t, err)
	assert.Len(t, speakers, 2)
	assert.Equal(t, "四国めたん", speakers[0].Name)
}

func TestManager_GetAvailableSpeakers_CacheHit(t *testing.T) {
	callCount := 0
	vv := &mockVoiceVoxAPI{speakers: testSpeakers}

	m := newTestManager(t, &mockRedisClient{}, vv)

	ctx := context.Background()

	// 1回目: VoiceVoxから取得してキャッシュに保存
	_, err := m.GetAvailableSpeakers(ctx)
	require.NoError(t, err)
	callCount++

	// キャッシュをVoiceVox APIエラーに差し替え（呼ばれないことを確認）
	m.voicevox = &mockVoiceVoxAPI{err: errors.New("should not be called")}

	// 2回目: キャッシュから返る
	speakers, err := m.GetAvailableSpeakers(ctx)
	require.NoError(t, err)
	assert.Len(t, speakers, 2, "キャッシュから返るべき")
	assert.Equal(t, 1, callCount, "VoiceVoxは1回しか呼ばれないべき")
}

func TestManager_GetAvailableSpeakers_VoiceVoxError(t *testing.T) {
	vv := &mockVoiceVoxAPI{err: errors.New("voicevox unavailable")}
	m := newTestManager(t, &mockRedisClient{}, vv)

	_, err := m.GetAvailableSpeakers(context.Background())
	assert.Error(t, err)
}

// --- ValidSpeaker テスト ---

func TestManager_ValidSpeaker_ValidID(t *testing.T) {
	vv := &mockVoiceVoxAPI{speakers: testSpeakers}
	m := newTestManager(t, &mockRedisClient{}, vv)

	// スタイルID 2 は存在する（四国めたんのノーマル）
	valid, err := m.ValidSpeaker(context.Background(), 2)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestManager_ValidSpeaker_InvalidID(t *testing.T) {
	vv := &mockVoiceVoxAPI{speakers: testSpeakers}
	m := newTestManager(t, &mockRedisClient{}, vv)

	// スタイルID 999 は存在しない
	valid, err := m.ValidSpeaker(context.Background(), 999)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestManager_ValidSpeaker_VoiceVoxError(t *testing.T) {
	vv := &mockVoiceVoxAPI{err: errors.New("voicevox unavailable")}
	m := newTestManager(t, &mockRedisClient{}, vv)

	_, err := m.ValidSpeaker(context.Background(), 2)
	assert.Error(t, err)
}

func TestManager_GetAvailableSpeakers_CacheExpiry(t *testing.T) {
	vv := &mockVoiceVoxAPI{speakers: testSpeakers}
	m := newTestManager(t, &mockRedisClient{}, vv)

	// キャッシュを手動で期限切れにする
	m.speakersCacheTTL = 1 * time.Millisecond

	ctx := context.Background()
	_, err := m.GetAvailableSpeakers(ctx)
	require.NoError(t, err)

	// TTL が切れるまで待つ
	time.Sleep(5 * time.Millisecond)

	// 新しい話者リストを返すよう差し替え
	newSpeakers := []voicevox.Speaker{{Name: "新話者", SpeakerUUID: "uuid-new", Styles: []voicevox.Style{{Name: "ノーマル", ID: 100}}}}
	m.voicevox = &mockVoiceVoxAPI{speakers: newSpeakers}

	speakers, err := m.GetAvailableSpeakers(ctx)
	require.NoError(t, err)
	assert.Equal(t, "新話者", speakers[0].Name, "キャッシュ期限切れ後は再取得されるべき")
}
