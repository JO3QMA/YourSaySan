package voice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hraban/opus"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidWAVFormat      = errors.New("invalid WAV format")
	ErrUnsupportedSampleRate = errors.New("unsupported sample rate (require 48kHz)")
	ErrUnsupportedChannels   = errors.New("unsupported channels (require 1 or 2)")
	ErrOpusEncodeFailed      = errors.New("opus encode failed")
)

// OpusEncoder はWAVデータをOpus形式にエンコードするエンコーダー
type OpusEncoder struct {
	// Opusエンコーダー（ステートフル）
	encoder *opus.Encoder
	mu      sync.Mutex

	// エンコーディングパラメータ
	sampleRate int // 48000Hz
	channels   int // 1 (mono) or 2 (stereo)
	bitrate    int // 64kbps
	frameSize  int // 960サンプル (20ms @ 48kHz)
}

// NewOpusEncoder は新しいOpusエンコーダーを作成します
func NewOpusEncoder() (*OpusEncoder, error) {
	sampleRate := 48000
	channels := 2 // ステレオ
	bitrate := 64000

	// AppVoIPはApplication型の定数（CGO依存のためlintが正しく解析できない場合がある）
	app := opus.AppVoIP
	enc, err := opus.NewEncoder(sampleRate, channels, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder: %w", err)
	}

	if err := enc.SetBitrate(bitrate); err != nil {
		return nil, fmt.Errorf("failed to set bitrate: %w", err)
	}

	// フレームサイズ: 20ms @ 48kHz = 960サンプル
	frameSize := sampleRate * 20 / 1000 // 960

	return &OpusEncoder{
		encoder:    enc,
		sampleRate: sampleRate,
		channels:   channels,
		bitrate:    bitrate,
		frameSize:  frameSize,
	}, nil
}

// EncodeBytes はWAVデータをOpus形式にエンコードします
func (e *OpusEncoder) EncodeBytes(ctx context.Context, wavData []byte) (<-chan []byte, error) {
	// WAVデコーダーを作成
	reader := bytes.NewReader(wavData)
	decoder := wav.NewDecoder(reader)

	// WAVフォーマットを読み取り
	decoder.ReadInfo()
	if !decoder.IsValidFile() {
		return nil, ErrInvalidWAVFormat
	}

	// サンプルレートとチャンネル数を確認
	sampleRate := int(decoder.SampleRate)
	channels := int(decoder.NumChans)

	// サンプルレートが48kHzでない場合はエラー
	if sampleRate != 48000 {
		return nil, fmt.Errorf("%w: got %dHz", ErrUnsupportedSampleRate, sampleRate)
	}

	// チャンネル数が1または2でない場合はエラー
	if channels != 1 && channels != 2 {
		return nil, fmt.Errorf("%w: got %d", ErrUnsupportedChannels, channels)
	}

	// エンコーダーをチャンネル数に合わせて再初期化（必要に応じて）
	if channels != e.channels {
		// エンコーダーを再作成
		e.mu.Lock()
		app := opus.AppVoIP
		enc, err := opus.NewEncoder(sampleRate, channels, app)
		if err != nil {
			e.mu.Unlock()
			return nil, fmt.Errorf("failed to create opus encoder: %w", err)
		}
		if err := enc.SetBitrate(e.bitrate); err != nil {
			e.mu.Unlock()
			return nil, fmt.Errorf("failed to set bitrate: %w", err)
		}
		e.encoder = enc
		e.channels = channels
		e.mu.Unlock()
	}

	// 出力チャンネルを作成
	output := make(chan []byte, 2)

	go func() {
		defer close(output)

		// PCMデータチャンクに移動
		if err := decoder.FwdToPCM(); err != nil {
			logrus.WithError(err).Error("Failed to forward to PCM chunk")
			return
		}

		// フレームサイズ（サンプル数）
		frameSamples := e.frameSize * channels

		// PCMバッファを作成（フレームサイズ分）
		pcmBuffer := &audio.IntBuffer{
			Format: &audio.Format{
				NumChannels: channels,
				SampleRate:  sampleRate,
			},
			Data: make([]int, frameSamples),
		}

		// Opusエンコード用バッファ（最大フレームサイズ）
		opusBuffer := make([]byte, 4000) // Opusフレームの最大サイズ

		// 残りのサンプルを保持するバッファ（int16に変換して保持）
		remainingSamples := make([]int16, 0, frameSamples)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 残りのサンプルがある場合はそれを使用
				if len(remainingSamples) >= frameSamples {
					// フレームサイズ分のサンプルを取得
					frameData := remainingSamples[:frameSamples]
					remainingSamples = remainingSamples[frameSamples:]

					// Opusエンコード
					e.mu.Lock()
					encodedLen, err := e.encoder.Encode(frameData, opusBuffer)
					e.mu.Unlock()

					if err != nil {
						logrus.WithError(err).Error("Failed to encode Opus frame")
						return
					}

					if encodedLen > 0 {
						// エンコードされたフレームを送信
						frame := make([]byte, encodedLen)
						copy(frame, opusBuffer[:encodedLen])

						select {
						case output <- frame:
						case <-ctx.Done():
							return
						}
					}
					continue
				}

				// PCMデータを読み取り
				_, err := decoder.PCMBuffer(pcmBuffer)
				if err != nil {
					if err == io.EOF {
						// 残りのサンプルがある場合はパディングしてエンコード
						if len(remainingSamples) > 0 {
							// フレームサイズに満たない場合はゼロで埋める
							padding := make([]int16, frameSamples-len(remainingSamples))
							remainingSamples = append(remainingSamples, padding...)

							e.mu.Lock()
							encodedLen, err := e.encoder.Encode(remainingSamples, opusBuffer)
							e.mu.Unlock()

							if err == nil && encodedLen > 0 {
								frame := make([]byte, encodedLen)
								copy(frame, opusBuffer[:encodedLen])
								select {
								case output <- frame:
								case <-ctx.Done():
									return
								}
							}
						}
						return
					}
					logrus.WithError(err).Error("Failed to read PCM data")
					return
				}

				// 読み取ったデータをint16に変換して残りのサンプルに追加
				for _, sample := range pcmBuffer.Data {
					// intをint16に変換（クリッピング）
					var s16 int16
					if sample > 32767 {
						s16 = 32767
					} else if sample < -32768 {
						s16 = -32768
					} else {
						s16 = int16(sample)
					}
					remainingSamples = append(remainingSamples, s16)
				}
			}
		}
	}()

	return output, nil
}

// EncodeFile はWAVファイルをOpus形式にエンコードします
func (e *OpusEncoder) EncodeFile(ctx context.Context, wavPath string) (<-chan []byte, error) {
	// ファイルを読み込んでEncodeBytesを呼び出す
	wavData, err := os.ReadFile(wavPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV file: %w", err)
	}
	return e.EncodeBytes(ctx, wavData)
}

// Close はエンコーダーを閉じます
func (e *OpusEncoder) Close() error {
	// opus.EncoderはGCで管理されるため、明示的なクローズは不要
	return nil
}
