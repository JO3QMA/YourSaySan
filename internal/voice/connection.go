package voice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Connection はギルドごとの VC 接続・再生を管理するコンポーネント。
//
// ライフサイクル:
//   - Join: VC に接続し Player を起動する
//   - Play: WAV データをキューに積む
//   - Stop: 現在の再生を中断しキューをクリアする（Player は継続）
//   - Leave: Player を停止し VC から切断する
type Connection struct {
	session      *discordgo.Session
	guildID      string
	channelID    string
	maxQueueSize int

	mu     sync.RWMutex
	vc     *discordgo.VoiceConnection
	player *Player
	queue  *Queue
	enc    Encoder
}

// NewConnection は Connection を作成する。Join を呼ぶまで VC には接続しない。
func NewConnection(session *discordgo.Session, maxQueueSize int) (*Connection, error) {
	encoder, err := NewEncoder()
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	return &Connection{
		session:      session,
		maxQueueSize: maxQueueSize,
		enc:          encoder,
	}, nil
}

// Join は指定 VC チャンネルに接続し Player を起動する。
func (c *Connection) Join(ctx context.Context, guildID, channelID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 既存の Player と Queue があれば先に停止
	if c.player != nil {
		c.player.Shutdown()
	}
	if c.queue != nil {
		c.queue.Close()
	}

	c.guildID = guildID
	c.channelID = channelID

	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Debug("joining voice channel")

	vc, err := c.session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}
	c.vc = vc

	if err := c.waitReady(ctx, vc); err != nil {
		vc.Disconnect()
		c.vc = nil
		return err
	}

	// 新しい Queue と Player を作成して起動
	q := NewQueue(c.maxQueueSize)
	p := NewPlayer(q, c.enc, vc)
	c.queue = q
	c.player = p
	p.Start(ctx)

	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Info("joined voice channel")

	return nil
}

func (c *Connection) waitReady(ctx context.Context, vc *discordgo.VoiceConnection) error {
	readyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-readyCtx.Done():
			return fmt.Errorf("timeout waiting for voice connection ready: %w", readyCtx.Err())
		case <-ticker.C:
			vc.RLock()
			ready := vc.Ready
			vc.RUnlock()
			if ready {
				return nil
			}
		}
	}
}

// Play は WAV データをキューに積む。
// エンコードは Player の goroutine 内で行うため、この関数はすぐに返る。
func (c *Connection) Play(_ context.Context, audioData []byte) error {
	c.mu.RLock()
	q := c.queue
	c.mu.RUnlock()

	if q == nil {
		return fmt.Errorf("voice connection is not ready")
	}

	item := AudioItem{
		Data:      audioData,
		GuildID:   c.guildID,
		ChannelID: c.channelID,
		Timestamp: time.Now(),
	}

	return q.Push(item)
}

// Stop は現在の再生を中断しキューをクリアする。
// Player goroutine は継続するため、次のアイテムが来れば再開できる。
func (c *Connection) Stop() error {
	c.mu.RLock()
	player := c.player
	c.mu.RUnlock()

	if player != nil {
		player.ClearAndInterrupt()
	}
	return nil
}

// Leave は Player を停止し VC から切断する。
func (c *Connection) Leave() error {
	c.mu.Lock()
	player := c.player
	q := c.queue
	vc := c.vc
	c.player = nil
	c.queue = nil
	c.vc = nil
	c.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"guild_id":   c.guildID,
		"channel_id": c.channelID,
	}).Debug("leaving voice channel")

	if player != nil {
		player.Shutdown()
	}
	if q != nil {
		q.Close()
	}

	var leaveErr error
	if vc != nil {
		if err := vc.Disconnect(); err != nil {
			logrus.WithError(err).WithField("guild_id", c.guildID).Error("failed to disconnect voice")
			leaveErr = fmt.Errorf("failed to disconnect: %w", err)
		}
	}

	logrus.WithField("guild_id", c.guildID).Info("left voice channel")
	return leaveErr
}

// GetChannelID は現在接続中の VC チャンネル ID を返す。
func (c *Connection) GetChannelID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channelID
}

// QueueSize は現在のキューのサイズを返す。
func (c *Connection) QueueSize() int {
	c.mu.RLock()
	q := c.queue
	c.mu.RUnlock()

	if q == nil {
		return 0
	}
	return q.Size()
}
