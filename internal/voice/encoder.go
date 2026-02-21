package voice

import (
	"context"
	"os"
)

// Encoder は WAV データを Opus フレームのスライスに変換するインターフェース。
// ストリーミングではなく一括変換とすることで goroutine リークを防ぐ。
type Encoder interface {
	Encode(ctx context.Context, wavData []byte) ([][]byte, error)
}

// NewEncoder は環境変数に基づいてエンコーダーを作成する。
// USE_PION_OPUS=true の場合 Opus エンコーダー、それ以外は DCA エンコーダー。
func NewEncoder() (Encoder, error) {
	if os.Getenv("USE_PION_OPUS") == "true" {
		return NewOpusEncoder()
	}
	return NewDCAEncoder(), nil
}
