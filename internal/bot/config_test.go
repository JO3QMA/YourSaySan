package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setEnv はテスト用に環境変数をセットし、テスト終了後に復元するヘルパー。
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

// mustLoadConfig は必須の環境変数を設定した上で LoadConfig を呼ぶヘルパー。
func mustSetRequiredEnvs(t *testing.T) {
	t.Helper()
	setEnv(t, "DISCORD_BOT_TOKEN", "test-token")
	setEnv(t, "DISCORD_CLIENT_ID", "test-client-id")
}

func TestLoadConfig_Success(t *testing.T) {
	mustSetRequiredEnvs(t)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "test-token", cfg.Bot.Token)
	assert.Equal(t, "test-client-id", cfg.Bot.ClientID)
}

func TestLoadConfig_MissingToken(t *testing.T) {
	// DISCORD_BOT_TOKEN を未設定にする
	t.Setenv("DISCORD_BOT_TOKEN", "")
	t.Setenv("DISCORD_CLIENT_ID", "test-client-id")

	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bot token is required")
}

func TestLoadConfig_MissingClientID(t *testing.T) {
	t.Setenv("DISCORD_BOT_TOKEN", "test-token")
	t.Setenv("DISCORD_CLIENT_ID", "")

	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bot client ID is required")
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	mustSetRequiredEnvs(t)
	// オプション環境変数をクリア
	t.Setenv("DISCORD_BOT_STATUS", "")
	t.Setenv("DISCORD_OWNER_ID", "")
	t.Setenv("VOICEVOX_HOST", "")
	t.Setenv("VOICEVOX_MAX_CHARS", "")
	t.Setenv("VOICEVOX_MAX_MESSAGE_LENGTH", "")
	t.Setenv("REDIS_HOST", "")
	t.Setenv("REDIS_PORT", "")
	t.Setenv("REDIS_DB", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "[TESTING] 読み上げBot", cfg.Bot.Status)
	assert.Equal(t, "123456789012345678", cfg.Bot.OwnerID)
	assert.Equal(t, "http://voicevox:50021", cfg.VoiceVox.Host)
	assert.Equal(t, 200, cfg.VoiceVox.MaxChars)
	assert.Equal(t, 50, cfg.VoiceVox.MaxMessageLength)
	assert.Equal(t, "redis", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, 0, cfg.Redis.DB)
}

func TestLoadConfig_CustomValues(t *testing.T) {
	mustSetRequiredEnvs(t)
	setEnv(t, "DISCORD_BOT_STATUS", "Custom Status")
	setEnv(t, "DISCORD_OWNER_ID", "999888777666555444")
	setEnv(t, "VOICEVOX_HOST", "http://custom-voicevox:50021")
	setEnv(t, "VOICEVOX_MAX_CHARS", "100")
	setEnv(t, "VOICEVOX_MAX_MESSAGE_LENGTH", "30")
	setEnv(t, "REDIS_HOST", "custom-redis")
	setEnv(t, "REDIS_PORT", "6380")
	setEnv(t, "REDIS_DB", "1")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "Custom Status", cfg.Bot.Status)
	assert.Equal(t, "999888777666555444", cfg.Bot.OwnerID)
	assert.Equal(t, "http://custom-voicevox:50021", cfg.VoiceVox.Host)
	assert.Equal(t, 100, cfg.VoiceVox.MaxChars)
	assert.Equal(t, 30, cfg.VoiceVox.MaxMessageLength)
	assert.Equal(t, "custom-redis", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, 1, cfg.Redis.DB)
}

func TestLoadConfig_InvalidIntFallsBackToDefault(t *testing.T) {
	mustSetRequiredEnvs(t)
	// 整数として不正な値を設定
	setEnv(t, "VOICEVOX_MAX_CHARS", "not-a-number")
	setEnv(t, "REDIS_PORT", "abc")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// デフォルト値にフォールバック
	assert.Equal(t, 200, cfg.VoiceVox.MaxChars)
	assert.Equal(t, 6379, cfg.Redis.Port)
}

func TestConfig_GetBotStatus(t *testing.T) {
	cfg := &Config{}
	cfg.Bot.Status = "テストステータス"
	assert.Equal(t, "テストステータス", cfg.GetBotStatus())
}

func TestConfig_GetBotClientID(t *testing.T) {
	cfg := &Config{}
	cfg.Bot.ClientID = "client-123"
	assert.Equal(t, "client-123", cfg.GetBotClientID())
}

func TestConfig_GetBotOwnerID(t *testing.T) {
	cfg := &Config{}
	cfg.Bot.OwnerID = "owner-456"
	assert.Equal(t, "owner-456", cfg.GetBotOwnerID())
}

func TestConfig_GetVoiceVoxMaxMessageLength(t *testing.T) {
	cfg := &Config{}
	cfg.VoiceVox.MaxMessageLength = 75
	assert.Equal(t, 75, cfg.GetVoiceVoxMaxMessageLength())
}
