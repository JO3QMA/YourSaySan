package voice

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Player struct {
	queue    *Queue
	encoder  *Encoder
	conn     *discordgo.VoiceConnection
	playing  atomic.Bool
	stopChan chan struct{}
	doneChan chan struct{}
	mu       sync.Mutex
	wg       sync.WaitGroup
}

func NewPlayer(queue *Queue, encoder *Encoder) *Player {
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
			if err := p.playAudio(ctx, item); err != nil {
				logrus.WithError(err).Error("Failed to play audio")
				continue
			}
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
	opusChan, err := p.encoder.EncodeBytes(ctx, item.Data)
	if err != nil {
		return fmt.Errorf("failed to encode audio: %w", err)
	}

	// 再生開始
	conn.Speaking(true)
	defer conn.Speaking(false)

	// OpusストリームをDiscordに送信
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.stopChan:
			return nil
		case opusFrame, ok := <-opusChan:
			if !ok {
				// エンコード完了
				return nil
			}
			select {
			case conn.OpusSend <- opusFrame:
			case <-ctx.Done():
				return ctx.Err()
			case <-p.stopChan:
				return nil
			}
		}
	}
}

