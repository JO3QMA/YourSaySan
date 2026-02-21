package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Player はキューから音声を取り出して Discord VC に送信するコンポーネント。
//
// ライフサイクル:
//   - Start(ctx): playLoop goroutine を起動する（Connection.Join から呼ばれる）
//   - ClearAndInterrupt(): 現在再生中のアイテムをキャンセルしキューをクリアする。
//     goroutine は終了しないため次のアイテムが来たら再開できる。
//   - Shutdown(): goroutine を完全に停止する（Connection.Leave から呼ばれる）
type Player struct {
	queue   *Queue
	encoder Encoder
	conn    *discordgo.VoiceConnection

	mu         sync.Mutex
	cancelPlay context.CancelFunc // 現在再生中のアイテムのキャンセル
	shutdownCh chan struct{}       // Shutdown() で閉じる
	doneCh     chan struct{}       // playLoop 終了通知
}

// NewPlayer は Player を作成する。Start を呼ぶまで再生は始まらない。
func NewPlayer(queue *Queue, encoder Encoder, conn *discordgo.VoiceConnection) *Player {
	return &Player{
		queue:      queue,
		encoder:    encoder,
		conn:       conn,
		cancelPlay: func() {}, // no-op
		shutdownCh: make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

// Start は playLoop goroutine を起動する。ctx は Bot 全体のコンテキスト。
func (p *Player) Start(ctx context.Context) {
	go p.playLoop(ctx)
}

// ClearAndInterrupt は現在再生中のアイテムをキャンセルしキューをクリアする。
// goroutine は生き続けるため、次のアイテムが Push されれば再生が再開する。
func (p *Player) ClearAndInterrupt() {
	p.mu.Lock()
	cancel := p.cancelPlay
	p.mu.Unlock()

	cancel()
	p.queue.Clear()
}

// Shutdown は playLoop goroutine を停止し、完了を待つ。
// Connection.Leave から呼ばれる。
func (p *Player) Shutdown() {
	p.mu.Lock()
	cancel := p.cancelPlay
	p.mu.Unlock()

	cancel()
	close(p.shutdownCh)
	<-p.doneCh
}

func (p *Player) playLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"panic": r,
				"stack": string(debug.Stack()),
			}).Error("panic in player loop")
		}
		p.conn.Speaking(false)
		close(p.doneCh)
	}()

	for {
		// Shutdown または Bot context キャンセルを確認
		select {
		case <-p.shutdownCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		item, err := p.queue.Pop(p.shutdownCh)
		if err != nil {
			// shutdownCh が閉じられた
			return
		}

		p.playItem(ctx, item)
	}
}

func (p *Player) playItem(ctx context.Context, item AudioItem) {
	// このアイテム専用のキャンセル可能な context を作成
	playCtx, cancel := context.WithCancel(ctx)
	p.mu.Lock()
	p.cancelPlay = cancel
	p.mu.Unlock()
	defer func() {
		cancel()
		p.mu.Lock()
		p.cancelPlay = func() {}
		p.mu.Unlock()
	}()

	logrus.WithFields(logrus.Fields{
		"guild_id":   item.GuildID,
		"audio_size": len(item.Data),
	}).Trace("encoding audio")

	// #region agent log A/B/C
	dbgLog("player.go:playItem-entry", "playItem called", "A", map[string]interface{}{"guild_id": item.GuildID, "audio_bytes": len(item.Data)})
	// #endregion

	frames, err := p.encoder.Encode(playCtx, item.Data)

	// #region agent log A/C
	encErr := ""
	if err != nil {
		encErr = err.Error()
	}
	ctxErr := ""
	if playCtx.Err() != nil {
		ctxErr = playCtx.Err().Error()
	}
	dbgLog("player.go:after-encode", "encode result", "A", map[string]interface{}{"frame_count": len(frames), "encode_err": encErr, "ctx_err": ctxErr})
	// #endregion

	if err != nil {
		if playCtx.Err() != nil {
			return // キャンセルされた
		}
		logrus.WithError(err).WithField("guild_id", item.GuildID).Error("failed to encode audio")
		return
	}

	if len(frames) == 0 {
		// #region agent log A
		dbgLog("player.go:zero-frames", "encoder returned 0 frames", "A", map[string]interface{}{"guild_id": item.GuildID})
		// #endregion
		return
	}

	logrus.WithFields(logrus.Fields{
		"guild_id":    item.GuildID,
		"frame_count": len(frames),
	}).Trace("sending opus frames")

	p.conn.Speaking(true)
	defer p.conn.Speaking(false)

	// 20ms ごとにフレームを送信（Discord Opus の標準フレーム長）
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	sentCount := 0
	for _, frame := range frames {
		select {
		case <-playCtx.Done():
			// #region agent log B/D
			dbgLog("player.go:playCtx-cancelled", "playCtx cancelled during send", "B", map[string]interface{}{"sent": sentCount, "total": len(frames)})
			// #endregion
			return
		case <-p.shutdownCh:
			// #region agent log B
			dbgLog("player.go:shutdown-during-send", "shutdownCh closed during send", "B", map[string]interface{}{"sent": sentCount, "total": len(frames)})
			// #endregion
			return
		case <-ticker.C:
			select {
			case p.conn.OpusSend <- frame:
				sentCount++
			case <-playCtx.Done():
				// #region agent log B/D
				dbgLog("player.go:playCtx-inner", "playCtx cancelled on OpusSend", "B", map[string]interface{}{"sent": sentCount, "total": len(frames)})
				// #endregion
				return
			case <-p.shutdownCh:
				return
			}
		}
	}

	// #region agent log B/D
	dbgLog("player.go:send-complete", "all frames sent", "B", map[string]interface{}{"sent": sentCount, "total": len(frames)})
	// #endregion

	logrus.WithField("guild_id", item.GuildID).Trace("audio playback completed")
}

// QueueSize は現在のキューサイズを返す。
func (p *Player) QueueSize() int {
	return p.queue.Size()
}

// SetConnection はボイス接続を差し替える（再接続時に使用）。
func (p *Player) SetConnection(conn *discordgo.VoiceConnection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
}

// IsActive は playLoop goroutine がまだ動いているか確認する。
func (p *Player) IsActive() bool {
	select {
	case <-p.doneCh:
		return false
	default:
		return true
	}
}

// formatDuration はデバッグ用のフレーム時間計算ヘルパー。
func formatDuration(frameCount int) string {
	ms := frameCount * 20
	return fmt.Sprintf("%dms", ms)
}

// #region agent log helper
func dbgLog(location, msg, hyp string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"sessionId":    "148c43",
		"runId":        "run4",
		"hypothesisId": hyp,
		"location":     location,
		"message":      msg,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	b, _ := json.Marshal(payload)
	f, _ := os.OpenFile("/app/.cursor/debug-148c43.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		f.Write(append(b, '\n'))
		f.Close()
	}
}

// #endregion
