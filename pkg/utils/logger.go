package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// InitLogger はロガーを初期化する
func InitLogger() {
	// 環境変数からログレベルを取得
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 環境変数からログ形式を取得
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

