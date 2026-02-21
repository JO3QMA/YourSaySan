package voice

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/jonas747/dca"
	"github.com/sirupsen/logrus"
)

// DCAEncoder は ffmpeg/dca を使って WAV → Opus に変換するエンコーダー。
type DCAEncoder struct {
	options *dca.EncodeOptions
}

// NewDCAEncoder は新しい DCAEncoder を作成する。
func NewDCAEncoder() *DCAEncoder {
	return &DCAEncoder{
		options: &dca.EncodeOptions{
			Volume:        256,
			FrameDuration: 20,
			Bitrate:       64,
			FrameRate:     48000,
			Channels:      2,
			Application:   dca.AudioApplicationVoip,
		},
	}
}

// Encode は WAV データを Opus フレームのスライスに変換して返す。
// 一時ファイルを使用し、defer で確実に削除する。
func (e *DCAEncoder) Encode(ctx context.Context, wavData []byte) ([][]byte, error) {
	tmpFile, err := os.CreateTemp("", "yoursay-*.wav")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(wavData); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	session, err := dca.EncodeFile(tmpPath, e.options)
	if err != nil {
		return nil, fmt.Errorf("failed to start dca encode: %w", err)
	}
	defer session.Cleanup()

	var frames [][]byte
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		frame, err := session.OpusFrame()
		if err != nil {
			if err == io.EOF {
				break
			}
			logrus.WithError(err).Warn("dca: error reading opus frame")
			break
		}

		// コピーして保持（session.Cleanup() 後も参照できるよう）
		cp := make([]byte, len(frame))
		copy(cp, frame)
		frames = append(frames, cp)
	}

	return frames, nil
}
