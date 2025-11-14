package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func (b *Bot) startHTTPServer() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// ヘルスチェック
	mux.HandleFunc("/health", b.healthCheckHandler)

	// Readinessプローブ
	mux.HandleFunc("/health/ready", b.readinessCheckHandler)

	// メトリクス（将来の実装）
	mux.HandleFunc("/metrics", b.metricsHandler)

	b.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	logrus.WithField("port", port).Info("Starting HTTP server")
	if err := b.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Error("HTTP server error")
	}
}

func (b *Bot) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (b *Bot) readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := map[string]bool{
		"discord":  b.checkDiscordHealth(),
		"voicevox": b.checkVoiceVoxHealth(ctx),
		"redis":    b.checkRedisHealth(ctx),
	}

	allHealthy := true
	for _, healthy := range checks {
		if !healthy {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(checks)
}

func (b *Bot) checkDiscordHealth() bool {
	return b.session != nil && b.session.State != nil
}

func (b *Bot) checkVoiceVoxHealth(ctx context.Context) bool {
	if b.voicevox == nil {
		return false
	}
	// 簡単なチェック: GetSpeakersを呼び出してエラーがないか確認
	_, err := b.voicevox.GetSpeakers(ctx)
	return err == nil
}

func (b *Bot) checkRedisHealth(ctx context.Context) bool {
	// RedisクライアントはSpeakerManager内にあるため、直接チェックできない
	// 簡単なチェックとして、SpeakerManagerが存在するか確認
	return b.speakerManager != nil
}

func (b *Bot) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// 将来の実装: Prometheusメトリクスを返す
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# Metrics endpoint (not yet implemented)\n")
}
