package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

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

	// #region agent log C/E
	var firstFrameBytes []byte
	var totalSize int
	if len(frames) > 0 {
		firstFrameBytes = frames[0]
		if len(firstFrameBytes) > 8 { firstFrameBytes = firstFrameBytes[:8] }
		for _, fr := range frames { totalSize += len(fr) }
	}
	func() {
		note := ""
		if len(frames) == 0 { note = "zero frames" }
		p := map[string]interface{}{"sessionId": "148c43", "runId": "run4", "hypothesisId": "C_E", "location": "encoder_dca.go:Encode-exit", "message": "dca encode finished", "data": map[string]interface{}{"frame_count": len(frames), "wav_bytes": len(wavData), "total_opus_bytes": totalSize, "avg_frame_bytes": func() int { if len(frames) > 0 { return totalSize / len(frames) }; return 0 }(), "first_frame_hex": fmt.Sprintf("%x", firstFrameBytes), "note": note}, "timestamp": time.Now().UnixMilli()}
		b, _ := json.Marshal(p)
		f, _ := os.OpenFile("/app/.cursor/debug-148c43.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil { f.Write(append(b, '\n')); f.Close() }
	}()
	// #endregion
	return frames, nil
}
