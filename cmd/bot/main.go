package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/JO3QMA/YourSaySan/internal/bot"
	"github.com/JO3QMA/YourSaySan/pkg/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	// ログ設定
	utils.InitLogger()
	logrus.Info("Starting bot application")

	// Bot初期化
	logrus.Debug("Initializing bot")
	b, err := bot.NewBot(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create bot")
	}
	logrus.Debug("Bot instance created successfully")

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Bot起動
	logrus.Info("Starting bot...")
	go func() {
		if err := b.Start(); err != nil {
			logrus.WithError(err).Fatal("Failed to start bot")
		}
	}()

	// シグナル待機
	<-sigChan
	logrus.Info("Shutdown signal received, shutting down...")

	// Bot停止
	if err := b.Stop(); err != nil {
		logrus.WithError(err).Error("Error during shutdown")
	}
	logrus.Info("Bot shutdown completed")
}
