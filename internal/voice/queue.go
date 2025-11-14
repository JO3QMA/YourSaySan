package voice

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrQueueFull    = errors.New("audio queue is full")
	ErrQueueTimeout = errors.New("audio queue timeout")
	ErrAudioTooLarge = errors.New("audio data too large (max 1MB)")
)

const (
	maxAudioItemSize = 1 * 1024 * 1024 // 1MB
)

type AudioItem struct {
	Data      []byte
	GuildID   string
	ChannelID string
	UserID    string
	Timestamp time.Time
}

type Queue struct {
	items   chan AudioItem
	maxSize int
	timeout time.Duration
	mu      sync.Mutex
	closed  atomic.Bool
}

func NewQueue(maxSize int) *Queue {
	return &Queue{
		items:   make(chan AudioItem, maxSize),
		maxSize: maxSize,
		timeout: 30 * time.Second,
	}
}

func (q *Queue) Push(ctx context.Context, item AudioItem) error {
	if q.closed.Load() {
		return errors.New("queue is closed")
	}

	// サイズチェック
	if len(item.Data) > maxAudioItemSize {
		return ErrAudioTooLarge
	}

	// キューが満杯の場合、古いアイテムを破棄（FIFO方式）
	select {
	case q.items <- item:
		return nil
	default:
		// キューが満杯の場合、1つ古いアイテムを破棄してから追加
		select {
		case <-q.items: // 古いアイテムを破棄
		default:
		}
		select {
		case q.items <- item:
			return nil
		default:
			return ErrQueueFull
		}
	}
}

func (q *Queue) Pop(ctx context.Context) (AudioItem, error) {
	if q.closed.Load() {
		return AudioItem{}, errors.New("queue is closed")
	}

	select {
	case item := <-q.items:
		return item, nil
	case <-ctx.Done():
		return AudioItem{}, ctx.Err()
	case <-time.After(q.timeout):
		return AudioItem{}, ErrQueueTimeout
	}
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for {
		select {
		case <-q.items:
		default:
			return
		}
	}
}

func (q *Queue) Size() int {
	return len(q.items)
}

func (q *Queue) Close() {
	q.closed.Store(true)
	close(q.items)
}

