package voicevox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	baseURL    string
	httpClient *http.Client

	// タイムアウト設定
	connectTimeout time.Duration // 接続タイムアウト: 3秒
	readTimeout    time.Duration // 読み込みタイムアウト: 10秒

	// リトライ設定
	maxRetries      int           // 最大リトライ回数: 3回
	retryBackoff    time.Duration // 初期バックオフ: 100ms
	retryBackoffMax time.Duration // 最大バックオフ: 2秒

	// レート制限
	rateLimiter *rate.Limiter // 1秒あたりのリクエスト数制限
}

func NewClient(baseURL string) *Client {
	connectTimeout := 3 * time.Second
	readTimeout := 10 * time.Second

	httpClient := &http.Client{
		Timeout: readTimeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: connectTimeout,
			}).DialContext,
			ResponseHeaderTimeout: readTimeout,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	return &Client{
		baseURL:         baseURL,
		httpClient:      httpClient,
		connectTimeout:  connectTimeout,
		readTimeout:     readTimeout,
		maxRetries:      3,
		retryBackoff:    100 * time.Millisecond,
		retryBackoffMax: 2 * time.Second,
		rateLimiter:     rate.NewLimiter(rate.Limit(10), 10), // 10 req/s
	}
}

func (c *Client) Speak(ctx context.Context, text string, speakerID int) ([]byte, error) {
	return c.speakWithRetry(ctx, text, speakerID)
}

func (c *Client) speakWithRetry(ctx context.Context, text string, speakerID int) ([]byte, error) {
	var audioData []byte
	err := c.withVoiceVoxRetry(ctx, func() error {
		var err error
		audioData, err = c.speakOnce(ctx, text, speakerID)
		return err
	})
	return audioData, err
}

// withVoiceVoxRetry は指数バックオフ・共有レート制限・4xx 即終了を Speak / CountMorae で共通化する。
func (c *Client) withVoiceVoxRetry(ctx context.Context, op func() error) error {
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.retryBackoff * time.Duration(1<<uint(attempt-1))
			if backoff > c.retryBackoffMax {
				backoff = c.retryBackoffMax
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		err := op()
		if err == nil {
			return nil
		}

		lastErr = err

		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			return err
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", c.maxRetries, lastErr)
}

// CountMorae は1行テキストのモーラ数を VoiceVox の分解に基づいて返す（Speak と同じリトライ・レート制限）。
func (c *Client) CountMorae(ctx context.Context, text string, speakerID int) (int, error) {
	var n int
	err := c.withVoiceVoxRetry(ctx, func() error {
		q, err := c.fetchAudioQuery(ctx, text, speakerID)
		if err != nil {
			return err
		}
		n = MoraeCountInQuery(q)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

// fetchAudioQuery は /audio_query を1回呼び出して結果を返す（レート制限・リトライは呼び出し側）。
func (c *Client) fetchAudioQuery(ctx context.Context, text string, speakerID int) (*AudioQuery, error) {
	encodedText := url.QueryEscape(text)
	queryURL := fmt.Sprintf("%s/audio_query?text=%s&speaker=%d", c.baseURL, encodedText, speakerID)
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request audio_query: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var audioQuery AudioQuery
	if err := json.NewDecoder(resp.Body).Decode(&audioQuery); err != nil {
		return nil, fmt.Errorf("failed to decode audio_query: %w", err)
	}

	return &audioQuery, nil
}

func (c *Client) speakOnce(ctx context.Context, text string, speakerID int) ([]byte, error) {
	audioQuery, err := c.fetchAudioQuery(ctx, text, speakerID)
	if err != nil {
		return nil, err
	}

	// DiscordのOpusエンコーダーは48kHzを要求するため、サンプルレートを48kHzに設定
	audioQuery.OutputSamplingRate = 48000

	queryJSON, err := json.Marshal(audioQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal audio_query: %w", err)
	}

	// /audio_query と /synthesis は別 HTTP リクエストのため、それぞれレート制限を取る
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	synthURL := fmt.Sprintf("%s/synthesis?speaker=%d", c.baseURL, speakerID)
	req, err := http.NewRequestWithContext(ctx, "POST", synthURL, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create synthesis request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request synthesis: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return audioData, nil
}

func (c *Client) GetSpeakers(ctx context.Context) ([]Speaker, error) {
	// レート制限
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	url := fmt.Sprintf("%s/speakers", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request speakers: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var speakers []Speaker
	if err := json.NewDecoder(resp.Body).Decode(&speakers); err != nil {
		return nil, fmt.Errorf("failed to decode speakers: %w", err)
	}

	return speakers, nil
}

// HTTPError はHTTPエラーを表す
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}
