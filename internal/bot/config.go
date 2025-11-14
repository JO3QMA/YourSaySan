package bot

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
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

func LoadConfig(path string) (*Config, error) {
	// 1. .envファイル読み込み
	if err := godotenv.Load(); err != nil {
		// .envファイルがなくても続行（環境変数から直接読み込む）
	}

	// 2. Viperで設定ファイル読み込み
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 3. 環境変数の自動読み込み
	viper.AutomaticEnv()
	viper.BindEnv("bot.token", "DISCORD_BOT_TOKEN")
	viper.BindEnv("bot.client_id", "DISCORD_CLIENT_ID")
	viper.BindEnv("bot.owner", "DISCORD_OWNER_ID")
	viper.BindEnv("voicevox.host", "VOICEVOX_HOST")
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.db", "REDIS_DB")

	// 4. 設定ファイル読み込み
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 4-1. 環境変数展開（YAML内の${VAR}形式を展開）
	for _, key := range viper.AllKeys() {
		val := viper.GetString(key)
		if strings.Contains(val, "${") {
			expanded := os.ExpandEnv(val)
			viper.Set(key, expanded)
		}
	}

	// 5. Config構造体にマッピング
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 6. 設定バリデーション
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
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
