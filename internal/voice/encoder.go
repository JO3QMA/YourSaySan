package voice

import (
	"context"
	"os"
)

// Encoder は音声エンコーダーのインターフェース
type Encoder interface {
	EncodeBytes(ctx context.Context, wavData []byte) (<-chan []byte, error)
	EncodeFile(ctx context.Context, wavPath string) (<-chan []byte, error)
}

// NewEncoder は環境変数に基づいてエンコーダーを作成します
// USE_PION_OPUS=true の場合、Opusエンコーダーを使用
// それ以外の場合、DCAエンコーダーを使用（後方互換性のため）
func NewEncoder() (Encoder, error) {
	if os.Getenv("USE_PION_OPUS") == "true" {
		return NewOpusEncoder()
	}
	return NewDCAEncoder(), nil
}
