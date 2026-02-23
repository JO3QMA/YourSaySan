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
	opusSampleRate   = 48000
	opusChannels     = 2
	opusBitrate      = 64000
	opusFrameMs      = 20
	opusFrameSamples = opusSampleRate * opusFrameMs / 1000 // 960
	opusMaxPacket    = 4000
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

	// Discord はステレオ (2ch) 48kHz を要求するため、常にステレオでエンコードする
	const outChannels = 2
	enc, err := opus.NewEncoder(sampleRate, outChannels, opus.AppVoIP)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder: %w", err)
	}
	if err := enc.SetBitrate(opusBitrate); err != nil {
		return nil, fmt.Errorf("failed to set bitrate: %w", err)
	}

	if err := dec.FwdToPCM(); err != nil {
		return nil, fmt.Errorf("failed to seek to PCM chunk: %w", err)
	}

	// 1フレームあたりの出力サンプル数 (20ms * 48kHz * 2ch = 1920)
	outFrameSamples := opusFrameSamples * outChannels
	// 入力バッファ (1フレーム分の入力を読み込む)
	buf := &audio.IntBuffer{
		Format: &audio.Format{NumChannels: channels, SampleRate: sampleRate},
		Data:   make([]int, opusFrameSamples*channels),
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

		n, err := dec.PCMBuffer(buf)
		if err == io.EOF || (err == nil && n == 0) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read PCM: %w", err)
		}

		for i := 0; i < n; i++ {
			s := buf.Data[i]
			var s16 int16
			switch {
			case s > 32767:
				s16 = 32767
			case s < -32768:
				s16 = -32768
			default:
				s16 = int16(s)
			}

			if channels == 1 {
				// モノラルをステレオに変換（同じ値を2回入れる）
				pending = append(pending, s16, s16)
			} else {
				// ステレオの場合はそのまま
				pending = append(pending, s16)
			}
		}

		for len(pending) >= outFrameSamples {
			if err := flush(pending[:outFrameSamples]); err != nil {
				return nil, err
			}
			pending = pending[outFrameSamples:]
		}
	}

	// 最後の端数フレームをゼロパディングしてエンコード
	if len(pending) > 0 {
		padded := make([]int16, outFrameSamples)
		copy(padded, pending)
		if err := flush(padded); err != nil {
			return nil, err
		}
	}

	return frames, nil
}
