package voice

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/jonas747/ogg"
	"github.com/sirupsen/logrus"
)

// DCAEncoder は ffmpeg を直接呼び出して WAV → Opus に変換するエンコーダー。
// 旧 jonas747/dca ライブラリの -vol オプションが ffmpeg 7.x で削除されたため、
// ffmpeg を直接起動し、ogg コンテナから Opus フレームを抽出する方式に変更。
type DCAEncoder struct {
	frameDuration int
	bitrate       int
	frameRate     int
	channels      int
	application   string
}

// NewDCAEncoder は新しい DCAEncoder を作成する。
func NewDCAEncoder() *DCAEncoder {
	return &DCAEncoder{
		frameDuration: 20,
		bitrate:       64,
		frameRate:     48000,
		channels:      2,
		application:   "voip",
	}
}

// Encode は WAV データを Opus フレームのスライスに変換して返す。
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

	args := []string{
		"-i", tmpPath,
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "ogg",
		"-ar", fmt.Sprintf("%d", e.frameRate),
		"-ac", fmt.Sprintf("%d", e.channels),
		"-b:a", fmt.Sprintf("%d", e.bitrate*1000),
		"-application", e.application,
		"-frame_duration", fmt.Sprintf("%d", e.frameDuration),
		"pipe:1",
	}

	ffmpeg := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr bytes.Buffer
	ffmpeg.Stderr = &stderr

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := ffmpeg.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	frames, readErr := readOpusFrames(stdout)

	waitErr := ffmpeg.Wait()

	if readErr != nil {
		logrus.WithError(readErr).Warn("error reading opus frames from ffmpeg")
	}

	if waitErr != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("ffmpeg failed: %w (stderr: %s)", waitErr, stderr.String())
	}

	return frames, nil
}

// readOpusFrames は ogg コンテナから Opus フレームを抽出する。
// 先頭2パケット（Opus ヘッダー + コメントヘッダー）をスキップする。
func readOpusFrames(r io.Reader) ([][]byte, error) {
	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(r))

	var frames [][]byte
	packetIndex := 0

	for {
		packet, _, err := decoder.Decode()
		if err != nil {
			if err == io.EOF {
				break
			}
			return frames, fmt.Errorf("ogg decode error: %w", err)
		}

		packetIndex++
		if packetIndex <= 2 {
			continue
		}

		cp := make([]byte, len(packet))
		copy(cp, packet)
		frames = append(frames, cp)
	}

	return frames, nil
}
