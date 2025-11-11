package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"discord-voice-bot/internal/infrastructure/config"
	"discord-voice-bot/internal/interface/discord"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start bot
	bot, err := discord.NewBot(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create context for graceful shutdown
	ctx := context.Background()

	// Start bot
	if err := bot.Start(ctx); err != nil {
		log.Fatalf("Bot exited with error: %v", err)
	}

	fmt.Println("Bot stopped gracefully")
}
