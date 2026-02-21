package voice

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrQueueClosed   = errors.New("queue is closed")
	ErrAudioTooLarge = errors.New("audio data too large (max 1MB)")
)

const maxAudioItemSize = 1 * 1024 * 1024 // 1MB

// AudioItem はキューに積まれる音声データ単位。
type AudioItem struct {
	Data      []byte
	GuildID   string
	ChannelID string
	UserID    string
	Timestamp time.Time
}

// Queue はスレッドセーフな音声キュー（slice + sync.Cond ベース）。
// channel ベースの実装と異なり、Close/Clear をデッドロックなく安全に行える。
type Queue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []AudioItem
	max    int
	closed bool
}

// NewQueue は指定サイズの Queue を作成する。
func NewQueue(maxSize int) *Queue {
	q := &Queue{max: maxSize}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Push はアイテムをキューに追加する。
// キューが満杯の場合は最古のアイテムを捨てて新しいアイテムを追加する（ドロップポリシー）。
func (q *Queue) Push(item AudioItem) error {
	if len(item.Data) > maxAudioItemSize {
		return ErrAudioTooLarge
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrQueueClosed
	}

	if len(q.items) >= q.max {
		q.items = q.items[1:]
	}
	q.items = append(q.items, item)
	q.cond.Signal()
	return nil
}

// Pop はキューからアイテムを取り出す。
// キューが空の場合は次のアイテムが来るまでブロックする。
// done チャネルが閉じられると ErrQueueClosed を返す。
func (q *Queue) Pop(done <-chan struct{}) (AudioItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 && !q.closed {
		// done が既に閉じていれば即座に返す
		select {
		case <-done:
			return AudioItem{}, ErrQueueClosed
		default:
		}

		// done チャネルが閉じられたら Cond を Signal して Wait を解除する
		// goroutine は Wait が解除されるか wakeCh が閉じられるかで終了する
		wakeCh := make(chan struct{})
		go func() {
			select {
			case <-done:
				// done が来たら Cond を Signal して Wait を起こす
				q.mu.Lock()
				q.cond.Signal()
				q.mu.Unlock()
			case <-wakeCh:
				// Pop が先に復帰した（正常ケース）
			}
		}()

		q.cond.Wait()
		close(wakeCh) // goroutine を停止
	}

	// 起床後に done チャネルを確認
	select {
	case <-done:
		return AudioItem{}, ErrQueueClosed
	default:
	}

	if q.closed && len(q.items) == 0 {
		return AudioItem{}, ErrQueueClosed
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

// Clear はキューを空にする（Close しない）。
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = q.items[:0]
}

// Close はキューを閉じ、待機中のすべての Pop を解除する。
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.cond.Broadcast()
}

// Size は現在のキューの長さを返す。
func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
