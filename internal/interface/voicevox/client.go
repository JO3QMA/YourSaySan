// internal/interface/voicevox/client.go
package voicevox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a VoiceVox Engine API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	maxLength  int
}

// Speaker represents a VoiceVox speaker
type Speaker struct {
	Name     string   `json:"name"`
	SpeakerUUID string `json:"speaker_uuid"`
	Styles   []Style  `json:"styles"`
	Version  string   `json:"version"`
}

// Style represents a speaker style
type Style struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// AudioQuery represents the audio query parameters
type AudioQuery struct {
	AccentPhrases      []AccentPhrase `json:"accent_phrases"`
	SpeedScale         float64        `json:"speedScale"`
	PitchScale         float64        `json:"pitchScale"`
	IntonationScale    float64        `json:"intonationScale"`
	VolumeScale        float64        `json:"volumeScale"`
	PrePhonemeLength   float64        `json:"prePhonemeLength"`
	PostPhonemeLength  float64        `json:"postPhonemeLength"`
	OutputSamplingRate int            `json:"outputSamplingRate"`
	OutputStereo       bool           `json:"outputStereo"`
	Kana               string         `json:"kana"`
}

// AccentPhrase represents an accent phrase in the audio query
type AccentPhrase struct {
	Moras           []Mora  `json:"moras"`
	Accent          int     `json:"accent"`
	PauseMora       *Mora   `json:"pause_mora"`
	IsInterrogative bool    `json:"is_interrogative"`
}

// Mora represents a mora in the accent phrase
type Mora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant"`
	ConsonantLength *float64 `json:"consonant_length"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowel_length"`
	Pitch           float64  `json:"pitch"`
}

// NewClient creates a new VoiceVox client
func NewClient(baseURL string, maxLength int) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxLength: maxLength,
	}
}

// GetSpeakers retrieves the list of available speakers
func (c *Client) GetSpeakers(ctx context.Context) ([]Speaker, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/speakers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get speakers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var speakers []Speaker
	if err := json.NewDecoder(resp.Body).Decode(&speakers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return speakers, nil
}

// CreateAudioQuery creates an audio query from text
func (c *Client) CreateAudioQuery(ctx context.Context, text string, speakerID int) (*AudioQuery, error) {
	if len(text) > c.maxLength {
		text = text[:c.maxLength]
	}

	params := url.Values{}
	params.Set("text", text)
	params.Set("speaker", fmt.Sprintf("%d", speakerID))

	reqURL := fmt.Sprintf("%s/audio_query?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var query AudioQuery
	if err := json.NewDecoder(resp.Body).Decode(&query); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &query, nil
}

// Synthesize generates audio from an audio query
func (c *Client) Synthesize(ctx context.Context, query *AudioQuery, speakerID int) ([]byte, error) {
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	params := url.Values{}
	params.Set("speaker", fmt.Sprintf("%d", speakerID))

	reqURL := fmt.Sprintf("%s/synthesis?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize audio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return audioData, nil
}

// GenerateVoice is a convenience method that combines CreateAudioQuery and Synthesize
func (c *Client) GenerateVoice(ctx context.Context, text string, speakerID int) ([]byte, error) {
	query, err := c.CreateAudioQuery(ctx, text, speakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio query: %w", err)
	}

	audioData, err := c.Synthesize(ctx, query, speakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize audio: %w", err)
	}

	return audioData, nil
}

// GetVersion retrieves the VoiceVox engine version
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/version", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

