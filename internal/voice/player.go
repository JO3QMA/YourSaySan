package voice

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Player struct {
	queue    *Queue
	encoder  Encoder
	conn     *discordgo.VoiceConnection
	playing  atomic.Bool
	stopChan chan struct{}
	doneChan chan struct{}
	mu       sync.Mutex
	wg       sync.WaitGroup
}

func NewPlayer(queue *Queue, encoder Encoder) *Player {
	return &Player{
		queue:    queue,
		encoder:  encoder,
		conn:     nil,
		playing:  atomic.Bool{},
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

func (p *Player) SetConnection(conn *discordgo.VoiceConnection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
}

func (p *Player) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.playing.Load() {
		return fmt.Errorf("player is already playing")
	}

	if p.conn == nil {
		return fmt.Errorf("voice connection is not set")
	}

	p.playing.Store(true)
	p.stopChan = make(chan struct{})
	p.doneChan = make(chan struct{})

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.playLoop(ctx)
	}()

	return nil
}

func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.playing.Load() {
		return nil
	}

	close(p.stopChan)
	p.playing.Store(false)

	// 完了を待つ
	<-p.doneChan

	return nil
}

func (p *Player) IsPlaying() bool {
	return p.playing.Load()
}

func (p *Player) playLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"panic": r,
				"stack": string(debug.Stack()),
			}).Error("Panic in play loop")
		}
		close(p.doneChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		default:
			// キューからアイテムを取得
			item, err := p.queue.Pop(ctx)
			if err != nil {
				if err == ErrQueueTimeout {
					continue
				}
				return
			}

			// 音声を再生
			logrus.WithFields(logrus.Fields{
				"guild_id":   item.GuildID,
				"channel_id": item.ChannelID,
				"audio_size": len(item.Data),
			}).Trace("Starting audio playback")
			if err := p.playAudio(ctx, item); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"guild_id":   item.GuildID,
					"channel_id": item.ChannelID,
				}).Error("Failed to play audio")
				continue
			}
			logrus.WithFields(logrus.Fields{
				"guild_id":   item.GuildID,
				"channel_id": item.ChannelID,
			}).Trace("Audio playback completed")
		}
	}
}

func (p *Player) playAudio(ctx context.Context, item AudioItem) error {
	p.mu.Lock()
	conn := p.conn
	p.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("voice connection is not set")
	}

	// WAVデータをOpus形式にエンコード
	logrus.WithFields(logrus.Fields{
		"guild_id":   item.GuildID,
		"channel_id": item.ChannelID,
		"audio_size": len(item.Data),
	}).Trace("Encoding audio to Opus")
	opusChan, err := p.encoder.EncodeBytes(ctx, item.Data)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"guild_id":   item.GuildID,
			"channel_id": item.ChannelID,
		}).Error("Failed to encode audio")
		return fmt.Errorf("failed to encode audio: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"guild_id":   item.GuildID,
		"channel_id": item.ChannelID,
	}).Trace("Audio encoded successfully")

	// 再生開始
	conn.Speaking(true)
	defer conn.Speaking(false)

	// OpusストリームをDiscordに送信
	// エンコーダーのチャンネルから読み取ったフレームを送信する
	// ただし、conn.OpusSendがブロックしている場合は、フレームを破棄して続行する
	// これにより、エンコーダーのチャンネルが閉じられた後、確実にconn.Speaking(false)が呼ばれる
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.stopChan:
			return nil
		case opusFrame, ok := <-opusChan:
			if !ok {
				// エンコード完了 - エンコーダーのチャンネルが閉じられた
				// この時点で、残りのフレームはすべて送信されたか、破棄された
				// conn.Speaking(false)はdeferで呼ばれる
				return nil
			}
			// conn.OpusSendへの送信を試みる
			// ブロックしている場合は、タイムアウトしてフレームを破棄する
			select {
			case conn.OpusSend <- opusFrame:
				// 送信成功
			case <-ctx.Done():
				return ctx.Err()
			case <-p.stopChan:
				return nil
			case <-time.After(50 * time.Millisecond):
				// タイムアウト: conn.OpusSendがブロックしている
				// フレームを破棄して続行（音声が少し途切れる可能性があるが、
				// 次の音声は再生できる）
				logrus.Debug("OpusSend channel is blocked, dropping frame")
				// フレームを破棄して続行
			}
		}
	}
}
