package voice

import (
	"context"
	"os"
	"testing"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/stretchr/testify/require"
)

// createMonoWavFile は 48kHz / 16bit / mono の WAV バイト列を組み立てる。
func createMonoWavFile(t *testing.T, samples []int16) []byte {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "mono-*.wav")
	require.NoError(t, err)

	enc := wav.NewEncoder(f, opusSampleRate, 16, 1, 1)
	buf := &audio.IntBuffer{
		Format: &audio.Format{NumChannels: 1, SampleRate: opusSampleRate},
		Data:   make([]int, len(samples)),
	}
	for i, s := range samples {
		buf.Data[i] = int(s)
	}
	require.NoError(t, enc.Write(buf))
	require.NoError(t, enc.Close())
	require.NoError(t, f.Close())

	data, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	return data
}

// createStereoWavFile は 48kHz / 16bit / stereo（LRLR インターリーブ）の WAV バイト列を組み立てる。
func createStereoWavFile(t *testing.T, interleavedLR []int16) []byte {
	t.Helper()

	if len(interleavedLR)%2 != 0 {
		t.Fatalf("stereo samples must be L,R pairs (even length), got %d", len(interleavedLR))
	}

	f, err := os.CreateTemp(t.TempDir(), "stereo-*.wav")
	require.NoError(t, err)

	enc := wav.NewEncoder(f, opusSampleRate, 16, 2, 1)
	buf := &audio.IntBuffer{
		Format: &audio.Format{NumChannels: 2, SampleRate: opusSampleRate},
		Data:   make([]int, len(interleavedLR)),
	}
	for i, s := range interleavedLR {
		buf.Data[i] = int(s)
	}
	require.NoError(t, enc.Write(buf))
	require.NoError(t, enc.Close())
	require.NoError(t, f.Close())

	data, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	return data
}

func TestOpusEncoder_Encode_Mono(t *testing.T) {
	// 1 フレーム分（20ms @ 48kHz mono = 960 サンプル）
	samples := make([]int16, opusFrameSamples)
	for i := range samples {
		samples[i] = int16(i % 3000)
	}
	wavData := createMonoWavFile(t, samples)

	enc, err := NewOpusEncoder()
	require.NoError(t, err)
	frames, err := enc.Encode(context.Background(), wavData)
	require.NoError(t, err)
	require.NotEmpty(t, frames, "expected at least one Opus frame")
}

func TestOpusEncoder_Encode_Stereo(t *testing.T) {
	// 1 フレーム分（20ms @ 48kHz stereo = 960 * 2 インターリーブサンプル）
	n := opusFrameSamples * 2
	interleaved := make([]int16, n)
	for i := 0; i < n; i += 2 {
		interleaved[i] = int16(i)
		interleaved[i+1] = int16(i + 1)
	}
	wavData := createStereoWavFile(t, interleaved)

	enc, err := NewOpusEncoder()
	require.NoError(t, err)
	frames, err := enc.Encode(context.Background(), wavData)
	require.NoError(t, err)
	require.NotEmpty(t, frames, "expected at least one Opus frame")
}
