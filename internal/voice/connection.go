package voice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Connection struct {
	session    *discordgo.Session
	guildID    string
	channelID  string
	connection *discordgo.VoiceConnection
	player     *Player
	mu         sync.RWMutex
}

func NewConnection(session *discordgo.Session, maxQueueSize int) (*Connection, error) {
	queue := NewQueue(maxQueueSize)
	encoder, err := NewEncoder()
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}
	player := NewPlayer(queue, encoder)

	return &Connection{
		session: session,
		player:  player,
	}, nil
}

func (c *Connection) Join(ctx context.Context, guildID, channelID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.guildID = guildID
	c.channelID = channelID

	// Discord VCに接続
	vc, err := c.session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	c.connection = vc

	// 音声接続がReadyになるまで待機（最大10秒）
	readyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-readyCtx.Done():
			return fmt.Errorf("timeout waiting for voice connection to be ready: %w", readyCtx.Err())
		case <-ticker.C:
			vc.RLock()
			ready := vc.Ready
			vc.RUnlock()
			if ready {
				// Playerに接続を設定
				c.player.SetConnection(vc)

				// 再生ループを開始
				if err := c.player.Start(ctx); err != nil {
					return fmt.Errorf("failed to start player: %w", err)
				}

				return nil
			}
		}
	}
}

func (c *Connection) Leave() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 再生を停止
	if err := c.player.Stop(); err != nil {
		return fmt.Errorf("failed to stop player: %w", err)
	}

	// VCから切断
	if c.connection != nil {
		if err := c.connection.Disconnect(); err != nil {
			return fmt.Errorf("failed to disconnect: %w", err)
		}
		c.connection = nil
	}

	return nil
}

func (c *Connection) Play(ctx context.Context, audioData []byte) error {
	c.mu.RLock()
	queue := c.player.queue
	c.mu.RUnlock()

	item := AudioItem{
		Data:      audioData,
		GuildID:   c.guildID,
		ChannelID: c.channelID,
		Timestamp: time.Now(),
	}

	return queue.Push(ctx, item)
}

func (c *Connection) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.player.Stop()
}

func (c *Connection) QueueSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.player.queue.Size()
}

func (c *Connection) GetChannelID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channelID
}
