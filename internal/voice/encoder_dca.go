package voice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jonas747/dca"
	"github.com/sirupsen/logrus"
)

var (
	ErrDiskFull         = errors.New("disk full")
	ErrPermissionDenied = errors.New("permission denied")
	ErrFFmpegNotFound   = errors.New("ffmpeg not found")
)

// DCAEncoder はDCAライブラリを使用するエンコーダー（後方互換性のため）
type DCAEncoder struct {
	options *dca.EncodeOptions
}

// NewDCAEncoder は新しいDCAエンコーダーを作成します
func NewDCAEncoder() *DCAEncoder {
	return &DCAEncoder{
		options: &dca.EncodeOptions{
			FrameDuration: 20,    // 20ms
			Bitrate:       64,    // 64kbps
			FrameRate:     48000, // 48000Hz
			Channels:      2,     // ステレオ
			Application:   dca.AudioApplicationVoip,
		},
	}
}

func (e *DCAEncoder) EncodeFile(ctx context.Context, wavPath string) (<-chan []byte, error) {
	encodeSession, err := dca.EncodeFile(wavPath, e.options)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFFmpegNotFound
		}
		if os.IsPermission(err) {
			return nil, ErrPermissionDenied
		}
		return nil, fmt.Errorf("failed to encode file: %w", err)
	}

	// チャンネルを作成してエンコード結果をストリーミング
	output := make(chan []byte, 2)
	go func() {
		defer close(output)
		defer encodeSession.Cleanup()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				frame, err := encodeSession.ReadFrame()
				if err != nil {
					if err == io.EOF {
						return
					}
					logrus.WithError(err).Error("Failed to read frame")
					return
				}
				select {
				case output <- frame:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return output, nil
}

func (e *DCAEncoder) EncodeBytes(ctx context.Context, wavData []byte) (<-chan []byte, error) {
	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "yoursay-*.wav")
	if err != nil {
		if os.IsPermission(err) {
			return nil, ErrPermissionDenied
		}
		// ディスク容量不足の可能性
		if _, ok := err.(*os.PathError); ok {
			return nil, ErrDiskFull
		}
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	// WAVデータを一時ファイルに書き込み
	if _, err := tmpFile.Write(wavData); err != nil {
		if os.IsPermission(err) {
			return nil, ErrPermissionDenied
		}
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// ファイルを閉じる（EncodeFileが開き直す）
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// EncodeFileを呼び出し
	return e.EncodeFile(ctx, tmpPath)
}
