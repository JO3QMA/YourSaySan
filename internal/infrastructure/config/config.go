package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Bot     BotConfig     `yaml:"bot"`
	VoiceVox VoiceVoxConfig `yaml:"voicevox"`
	Redis   RedisConfig   `yaml:"redis"`
}

// BotConfig represents Discord bot configuration
type BotConfig struct {
	Token    string `yaml:"token"`
	ClientID string `yaml:"client_id"`
	Prefix   string `yaml:"prefix"`
	Status   string `yaml:"status"`
	Owner    string `yaml:"owner"`
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

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Load from YAML file if it exists
	config := &Config{}
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	config.overrideWithEnv()

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// overrideWithEnv overrides configuration with environment variables
func (c *Config) overrideWithEnv() {
	// Bot configuration
	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		c.Bot.Token = token
	}
	if clientID := os.Getenv("DISCORD_CLIENT_ID"); clientID != "" {
		c.Bot.ClientID = clientID
	}
	if prefix := os.Getenv("DISCORD_PREFIX"); prefix != "" {
		c.Bot.Prefix = prefix
	}
	if status := os.Getenv("DISCORD_STATUS"); status != "" {
		c.Bot.Status = status
	}
	if owner := os.Getenv("DISCORD_OWNER_ID"); owner != "" {
		c.Bot.Owner = owner
	}

	// VoiceVox configuration
	if maxStr := os.Getenv("VOICEVOX_MAX"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil {
			c.VoiceVox.Max = max
		}
	}
	if host := os.Getenv("VOICEVOX_HOST"); host != "" {
		c.VoiceVox.Host = host
	}

	// Redis configuration
	if host := os.Getenv("REDIS_HOST"); host != "" {
		c.Redis.Host = host
	}
	if portStr := os.Getenv("REDIS_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			c.Redis.Port = port
		}
	}
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			c.Redis.DB = db
		}
	}
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.Bot.Token == "" {
		return fmt.Errorf("bot token is required")
	}
	if c.Bot.ClientID == "" {
		return fmt.Errorf("bot client_id is required")
	}
	if c.VoiceVox.Host == "" {
		return fmt.Errorf("voicevox host is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	return nil
}
