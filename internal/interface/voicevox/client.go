package voicevox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents VoiceVox API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// Speaker represents VoiceVox speaker information
type Speaker struct {
	Name        string `json:"name"`
	SpeakerUUID string `json:"speaker_uuid"`
	Styles      []Style `json:"styles"`
}

// Style represents VoiceVox speaker style
type Style struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// QueryResponse represents audio query response
type QueryResponse struct {
	AccentPhrases []interface{} `json:"accent_phrases"`
	SpeedScale    float64       `json:"speedScale"`
	PitchScale    float64       `json:"pitchScale"`
	IntonationScale float64     `json:"intonationScale"`
	VolumeScale   float64       `json:"volumeScale"`
	PrePhonemeLength float64    `json:"prePhonemeLength"`
	PostPhonemeLength float64   `json:"postPhonemeLength"`
	OutputSamplingRate int      `json:"outputSamplingRate"`
	OutputStereo       bool     `json:"outputStereo"`
	Kana               string   `json:"kana"`
}

// NewClient creates a new VoiceVox API client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// GetSpeakers retrieves available speakers
func (c *Client) GetSpeakers(ctx context.Context) ([]Speaker, error) {
	url := fmt.Sprintf("%s/speakers", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var speakers []Speaker
	if err := json.NewDecoder(resp.Body).Decode(&speakers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return speakers, nil
}

// GenerateAudio generates audio from text using specified speaker
func (c *Client) GenerateAudio(ctx context.Context, text string, speakerID int) ([]byte, error) {
	// First, create audio query
	query, err := c.createAudioQuery(ctx, text, speakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio query: %w", err)
	}

	// Then, synthesize audio
	audio, err := c.synthesizeAudio(ctx, query, speakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize audio: %w", err)
	}

	return audio, nil
}

// createAudioQuery creates an audio query for text synthesis
func (c *Client) createAudioQuery(ctx context.Context, text string, speakerID int) ([]byte, error) {
	url := fmt.Sprintf("%s/audio_query?text=%s&speaker=%d", c.baseURL, text, speakerID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	query, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return query, nil
}

// synthesizeAudio synthesizes audio from query
func (c *Client) synthesizeAudio(ctx context.Context, query []byte, speakerID int) ([]byte, error) {
	url := fmt.Sprintf("%s/synthesis?speaker=%d", c.baseURL, speakerID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return audio, nil
}
