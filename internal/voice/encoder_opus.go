package voice

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hraban/opus"
)

const (
	opusSampleRate  = 48000
	opusChannels    = 2
	opusBitrate     = 64000
	opusFrameMs     = 20
	opusFrameSamples = opusSampleRate * opusFrameMs / 1000 // 960
	opusMaxPacket   = 4000
)

// OpusEncoder は CGO opus ライブラリで WAV → Opus に変換するエンコーダー。
// 呼び出しごとに opus.Encoder を新規作成するため共有状態なし。
type OpusEncoder struct{}

// NewOpusEncoder は新しい OpusEncoder を作成する。
func NewOpusEncoder() (*OpusEncoder, error) {
	return &OpusEncoder{}, nil
}

// Encode は WAV データを Opus フレームのスライスに変換して返す。
func (e *OpusEncoder) Encode(ctx context.Context, wavData []byte) ([][]byte, error) {
	dec := wav.NewDecoder(bytes.NewReader(wavData))
	dec.ReadInfo()
	if !dec.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV data")
	}

	sampleRate := int(dec.SampleRate)
	channels := int(dec.NumChans)

	if sampleRate != opusSampleRate {
		return nil, fmt.Errorf("unsupported sample rate %dHz (require 48kHz)", sampleRate)
	}
	if channels != 1 && channels != 2 {
		return nil, fmt.Errorf("unsupported channel count %d (require 1 or 2)", channels)
	}

	enc, err := opus.NewEncoder(sampleRate, channels, opus.AppVoIP)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder: %w", err)
	}
	if err := enc.SetBitrate(opusBitrate); err != nil {
		return nil, fmt.Errorf("failed to set bitrate: %w", err)
	}

	if err := dec.FwdToPCM(); err != nil {
		return nil, fmt.Errorf("failed to seek to PCM chunk: %w", err)
	}

	frameSamples := opusFrameSamples * channels
	buf := &audio.IntBuffer{
		Format: &audio.Format{NumChannels: channels, SampleRate: sampleRate},
		Data:   make([]int, frameSamples),
	}
	opusBuf := make([]byte, opusMaxPacket)

	// 端数サンプルを保持するリングバッファ
	var pending []int16
	var frames [][]byte

	flush := func(pcm []int16) error {
		n, err := enc.Encode(pcm, opusBuf)
		if err != nil {
			return fmt.Errorf("opus encode: %w", err)
		}
		if n > 0 {
			cp := make([]byte, n)
			copy(cp, opusBuf[:n])
			frames = append(frames, cp)
		}
		return nil
	}

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		_, err := dec.PCMBuffer(buf)
		if err == io.EOF || (err == nil && len(buf.Data) == 0) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read PCM: %w", err)
		}

		for _, s := range buf.Data {
			var s16 int16
			switch {
			case s > 32767:
				s16 = 32767
			case s < -32768:
				s16 = -32768
			default:
				s16 = int16(s)
			}
			pending = append(pending, s16)
		}

		for len(pending) >= frameSamples {
			if err := flush(pending[:frameSamples]); err != nil {
				return nil, err
			}
			pending = pending[frameSamples:]
		}
	}

	// 最後の端数フレームをゼロパディングしてエンコード
	if len(pending) > 0 {
		padded := make([]int16, frameSamples)
		copy(padded, pending)
		if err := flush(padded); err != nil {
			return nil, err
		}
	}

	return frames, nil
}
