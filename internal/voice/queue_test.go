package voice

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeItem(data []byte) AudioItem {
	return AudioItem{
		Data:      data,
		GuildID:   "guild1",
		ChannelID: "channel1",
		UserID:    "user1",
		Timestamp: time.Now(),
	}
}

func TestQueue_PushPop_Basic(t *testing.T) {
	q := NewQueue(10)
	item := makeItem([]byte("hello"))

	err := q.Push(item)
	require.NoError(t, err)
	assert.Equal(t, 1, q.Size())

	done := make(chan struct{})
	got, err := q.Pop(done)
	require.NoError(t, err)
	assert.Equal(t, item.Data, got.Data)
	assert.Equal(t, 0, q.Size())
}

func TestQueue_Push_AudioTooLarge(t *testing.T) {
	q := NewQueue(10)
	// 1MB + 1 byte
	largeData := make([]byte, maxAudioItemSize+1)
	err := q.Push(makeItem(largeData))
	assert.ErrorIs(t, err, ErrAudioTooLarge)
	assert.Equal(t, 0, q.Size())
}

func TestQueue_DropPolicy_WhenFull(t *testing.T) {
	q := NewQueue(3)

	// 3アイテムをプッシュ（満杯）
	for i := range 3 {
		data := []byte{byte(i)}
		require.NoError(t, q.Push(makeItem(data)))
	}
	assert.Equal(t, 3, q.Size())

	// 4つ目をプッシュすると最古（data=0x00）が捨てられる
	require.NoError(t, q.Push(makeItem([]byte{99})))
	assert.Equal(t, 3, q.Size())

	// 最初に取り出せるのは 0x01（最古の 0x00 が捨てられているはず）
	done := make(chan struct{})
	got, err := q.Pop(done)
	require.NoError(t, err)
	assert.Equal(t, []byte{1}, got.Data)
}

func TestQueue_Close_PushReturnsError(t *testing.T) {
	q := NewQueue(10)
	q.Close()

	err := q.Push(makeItem([]byte("data")))
	assert.ErrorIs(t, err, ErrQueueClosed)
}

func TestQueue_Close_PopReturnsError(t *testing.T) {
	q := NewQueue(10)

	// 別goroutineで少し後にClose
	go func() {
		time.Sleep(10 * time.Millisecond)
		q.Close()
	}()

	done := make(chan struct{})
	_, err := q.Pop(done)
	assert.ErrorIs(t, err, ErrQueueClosed)
}

func TestQueue_Pop_CancelledByDoneChannel(t *testing.T) {
	q := NewQueue(10)
	done := make(chan struct{})

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(done)
	}()

	_, err := q.Pop(done)
	assert.ErrorIs(t, err, ErrQueueClosed)
}

func TestQueue_Pop_AlreadyClosedDone(t *testing.T) {
	q := NewQueue(10)
	done := make(chan struct{})
	close(done) // 最初から閉じている

	_, err := q.Pop(done)
	assert.ErrorIs(t, err, ErrQueueClosed)
}

func TestQueue_Clear(t *testing.T) {
	q := NewQueue(10)
	for range 5 {
		require.NoError(t, q.Push(makeItem([]byte("x"))))
	}
	assert.Equal(t, 5, q.Size())

	q.Clear()
	assert.Equal(t, 0, q.Size())
}

func TestQueue_Clear_DoesNotClose(t *testing.T) {
	q := NewQueue(10)
	q.Clear()

	// Clear後でもPushできる
	err := q.Push(makeItem([]byte("data")))
	assert.NoError(t, err)
}

func TestQueue_MultipleItems_FIFO(t *testing.T) {
	q := NewQueue(10)
	for i := range 5 {
		require.NoError(t, q.Push(makeItem([]byte{byte(i)})))
	}

	done := make(chan struct{})
	for i := range 5 {
		got, err := q.Pop(done)
		require.NoError(t, err)
		assert.Equal(t, []byte{byte(i)}, got.Data, "FIFO order violation at index %d", i)
	}
}

func TestQueue_ConcurrentPushPop(t *testing.T) {
	q := NewQueue(100)
	const count = 50

	var wg sync.WaitGroup
	received := make([]bool, count)
	var mu sync.Mutex

	// コンシューマー
	wg.Add(1)
	go func() {
		defer wg.Done()
		done := make(chan struct{})
		for range count {
			item, err := q.Pop(done)
			if err != nil {
				return
			}
			mu.Lock()
			if len(item.Data) == 1 {
				received[int(item.Data[0])] = true
			}
			mu.Unlock()
		}
	}()

	// プロデューサー
	for i := range count {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(i%5) * time.Millisecond)
			_ = q.Push(makeItem([]byte{byte(i)}))
		}()
	}

	wg.Wait()

	// 全アイテムが受信できている
	mu.Lock()
	defer mu.Unlock()
	for i, got := range received {
		assert.True(t, got, "item %d was not received", i)
	}
}

func TestQueue_Size_ThreadSafe(t *testing.T) {
	q := NewQueue(1000)
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = q.Push(makeItem([]byte("x")))
			_ = q.Size()
		}()
	}

	wg.Wait()
}
