package voice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
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

	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Debug("Joining voice channel")

	// Discord VCに接続
	vc, err := c.session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"guild_id":   guildID,
			"channel_id": channelID,
		}).Error("Failed to join voice channel")
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	c.connection = vc

	// 音声接続がReadyになるまで待機（最大10秒）
	readyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	logrus.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Trace("Waiting for voice connection to be ready")

	for {
		select {
		case <-readyCtx.Done():
			logrus.WithError(readyCtx.Err()).WithFields(logrus.Fields{
				"guild_id":   guildID,
				"channel_id": channelID,
			}).Error("Timeout waiting for voice connection to be ready")
			return fmt.Errorf("timeout waiting for voice connection to be ready: %w", readyCtx.Err())
		case <-ticker.C:
			vc.RLock()
			ready := vc.Ready
			vc.RUnlock()
			if ready {
				logrus.WithFields(logrus.Fields{
					"guild_id":   guildID,
					"channel_id": channelID,
				}).Trace("Voice connection is ready")

				// Playerに接続を設定
				c.player.SetConnection(vc)

				// 再生ループを開始
				if err := c.player.Start(ctx); err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"guild_id":   guildID,
						"channel_id": channelID,
					}).Error("Failed to start player")
					return fmt.Errorf("failed to start player: %w", err)
				}

				logrus.WithFields(logrus.Fields{
					"guild_id":   guildID,
					"channel_id": channelID,
				}).Info("Successfully joined voice channel and started player")

				return nil
			}
		}
	}
}

func (c *Connection) Leave() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"guild_id":   c.guildID,
		"channel_id": c.channelID,
	}).Debug("Leaving voice channel")

	// 再生を停止
	if err := c.player.Stop(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"guild_id":   c.guildID,
			"channel_id": c.channelID,
		}).Error("Failed to stop player")
		return fmt.Errorf("failed to stop player: %w", err)
	}

	// VCから切断
	if c.connection != nil {
		if err := c.connection.Disconnect(); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"guild_id":   c.guildID,
				"channel_id": c.channelID,
			}).Error("Failed to disconnect from voice channel")
			return fmt.Errorf("failed to disconnect: %w", err)
		}
		c.connection = nil
	}

	logrus.WithFields(logrus.Fields{
		"guild_id":   c.guildID,
		"channel_id": c.channelID,
	}).Info("Successfully left voice channel")

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

	logrus.WithFields(logrus.Fields{
		"guild_id":   c.guildID,
		"channel_id": c.channelID,
		"audio_size": len(audioData),
	}).Trace("Queueing audio for playback")

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
