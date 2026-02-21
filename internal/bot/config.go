package bot

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Bot struct {
		Token    string `yaml:"token" mapstructure:"token"`
		ClientID string `yaml:"client_id" mapstructure:"client_id"`
		Status   string `yaml:"status" mapstructure:"status"`
		OwnerID  string `yaml:"owner" mapstructure:"owner"`
	} `yaml:"bot" mapstructure:"bot"`

	VoiceVox struct {
		MaxChars         int    `yaml:"max_chars" mapstructure:"max_chars"`
		MaxMessageLength int    `yaml:"max_message_length" mapstructure:"max_message_length"`
		Host             string `yaml:"host" mapstructure:"host"`
	} `yaml:"voicevox" mapstructure:"voicevox"`

	Redis struct {
		Host string `yaml:"host" mapstructure:"host"`
		Port int    `yaml:"port" mapstructure:"port"`
		DB   int    `yaml:"db" mapstructure:"db"`
	} `yaml:"redis" mapstructure:"redis"`
}

// GetBotStatus はBotのステータスを返す
func (c *Config) GetBotStatus() string {
	return c.Bot.Status
}

// GetBotClientID はBotのクライアントIDを返す
func (c *Config) GetBotClientID() string {
	return c.Bot.ClientID
}

// GetBotOwnerID はBotのオーナーIDを返す
func (c *Config) GetBotOwnerID() string {
	return c.Bot.OwnerID
}

// GetVoiceVoxMaxMessageLength は読み上げメッセージの最大長を返す
func (c *Config) GetVoiceVoxMaxMessageLength() int {
	return c.VoiceVox.MaxMessageLength
}

func LoadConfig() (*Config, error) {
	// 1. .envファイル読み込み
	if err := godotenv.Load(); err != nil {
		// .envファイルがなくても続行（環境変数から直接読み込む）
	}

	// 2. 環境変数から直接設定を読み込み
	var config Config

	// Bot設定
	config.Bot.Token = os.Getenv("DISCORD_BOT_TOKEN")
	config.Bot.ClientID = os.Getenv("DISCORD_CLIENT_ID")
	config.Bot.Status = getEnvWithDefault("DISCORD_BOT_STATUS", "[TESTING] 読み上げBot")
	if ownerID := os.Getenv("DISCORD_OWNER_ID"); ownerID != "" {
		config.Bot.OwnerID = ownerID
	} else {
		config.Bot.OwnerID = "123456789012345678" // デフォルト値
	}

	// VoiceVox設定
	config.VoiceVox.Host = getEnvWithDefault("VOICEVOX_HOST", "http://voicevox:50021")
	config.VoiceVox.MaxChars = getEnvIntWithDefault("VOICEVOX_MAX_CHARS", 200)
	config.VoiceVox.MaxMessageLength = getEnvIntWithDefault("VOICEVOX_MAX_MESSAGE_LENGTH", 50)

	// Redis設定
	config.Redis.Host = getEnvWithDefault("REDIS_HOST", "redis")
	config.Redis.Port = getEnvIntWithDefault("REDIS_PORT", 6379)
	config.Redis.DB = getEnvIntWithDefault("REDIS_DB", 0)

	// 3. 設定バリデーション
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// getEnvWithDefault は環境変数を取得し、デフォルト値を返す
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntWithDefault は環境変数をintとして取得し、デフォルト値を返す
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func validateConfig(config *Config) error {
	if config.Bot.Token == "" {
		return errors.New("bot token is required")
	}
	if config.Bot.ClientID == "" {
		return errors.New("bot client ID is required")
	}
	if config.VoiceVox.Host == "" {
		return errors.New("voicevox host is required")
	}
	if config.VoiceVox.MaxChars <= 0 {
		return errors.New("voicevox max chars must be positive")
	}
	if config.VoiceVox.MaxMessageLength <= 0 {
		config.VoiceVox.MaxMessageLength = 50 // デフォルト値
	}
	if config.Redis.Host == "" {
		config.Redis.Host = "redis" // デフォルト値
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379 // デフォルト値
	}
	return nil
}
