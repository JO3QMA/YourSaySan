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
		logrus.WithFields(logrus.Fields{
			"invalid_level": logLevel,
			"default_level": "info",
		}).Warn("Invalid LOG_LEVEL, using default")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Traceレベルの場合はReportCallerを有効化
	if level == logrus.TraceLevel {
		logrus.SetReportCaller(true)
	}

	// 環境変数からログ形式を取得
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// 起動時に現在のログ設定を表示
	logrus.WithFields(logrus.Fields{
		"level":  level.String(),
		"format": logFormat,
	}).Info("Logger initialized")
}

