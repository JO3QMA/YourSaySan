// internal/infrastructure/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Bot      BotConfig      `yaml:"bot"`
	VoiceVox VoiceVoxConfig `yaml:"voicevox"`
	Redis    RedisConfig    `yaml:"redis"`
}

// BotConfig represents Discord bot configuration
type BotConfig struct {
	Token    string `yaml:"token"`
	ClientID string `yaml:"client_id"`
	Prefix   string `yaml:"prefix"`
	Status   string `yaml:"status"`
	OwnerID  string `yaml:"owner"`
}

// VoiceVoxConfig represents VoiceVox engine configuration
type VoiceVoxConfig struct {
	Max  int    `yaml:"max"`
	Host string `yaml:"host"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	DB   int    `yaml:"db"`
}

// Load reads and parses the configuration file with environment variable substitution
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	cfg.applyEnvironmentVariables()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// applyEnvironmentVariables overrides config values with environment variables
func (c *Config) applyEnvironmentVariables() {
	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		c.Bot.Token = token
	}
	if clientID := os.Getenv("DISCORD_CLIENT_ID"); clientID != "" {
		c.Bot.ClientID = clientID
	}
	if ownerID := os.Getenv("DISCORD_OWNER_ID"); ownerID != "" {
		c.Bot.OwnerID = ownerID
	}
	if host := os.Getenv("VOICEVOX_HOST"); host != "" {
		c.VoiceVox.Host = host
	}
	if host := os.Getenv("REDIS_HOST"); host != "" {
		c.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Redis.Port = p
		}
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			c.Redis.DB = d
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Bot.Token == "" {
		return fmt.Errorf("bot token is required")
	}
	if c.Bot.ClientID == "" {
		return fmt.Errorf("bot client ID is required")
	}
	if c.VoiceVox.Host == "" {
		return fmt.Errorf("voicevox host is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port <= 0 {
		return fmt.Errorf("redis port must be positive")
	}
	return nil
}

