package voice

import (
	"context"
	"os"
	"testing"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func createMonoWavFile(t *testing.T, sampleRate int, samples []int) string {
	f, err := os.CreateTemp("", "test*.wav")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f.Close()

	enc := wav.NewEncoder(f, sampleRate, 16, 1, 1)
	intBuf := &audio.IntBuffer{
		Format: &audio.Format{NumChannels: 1, SampleRate: sampleRate},
		Data:   samples,
	}
	if err := enc.Write(intBuf); err != nil {
		t.Fatalf("failed to write wav: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("failed to close wav: %v", err)
	}
	return f.Name()
}

func TestOpusEncoder_Reproduction(t *testing.T) {
	// 48kHz mono WAV, 20ms = 960 samples
	samples := make([]int, 960)
	for i := range samples {
		samples[i] = i % 1000
	}
	wavPath := createMonoWavFile(t, 48000, samples)
	defer os.Remove(wavPath)

	wavData, err := os.ReadFile(wavPath)
	if err != nil {
		t.Fatalf("failed to read wav file: %v", err)
	}

	encoder, err := NewOpusEncoder()
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	frames, err := encoder.Encode(context.Background(), wavData)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(frames) != 1 {
		t.Errorf("Expected 1 frame, got %d", len(frames))
	}
}

func TestOpusEncoder_Padding(t *testing.T) {
	// 48kHz mono WAV, 30ms = 1440 samples
	samples := make([]int, 1440)
	for i := range samples {
		samples[i] = i % 1000
	}
	wavPath := createMonoWavFile(t, 48000, samples)
	defer os.Remove(wavPath)

	wavData, err := os.ReadFile(wavPath)
	if err != nil {
		t.Fatalf("failed to read wav file: %v", err)
	}

	encoder, err := NewOpusEncoder()
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	frames, err := encoder.Encode(context.Background(), wavData)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(frames) != 2 {
		t.Errorf("Expected 2 frames, got %d", len(frames))
	}
}
