package bot

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	s := NewState()
	assert.NotNil(t, s)
	assert.NotNil(t, s.Guilds)
	assert.Equal(t, 0, s.GetGuildCount())
}

func TestState_AddTextChannel_IsTextChannelActive(t *testing.T) {
	s := NewState()

	// 追加前はfalse
	assert.False(t, s.IsTextChannelActive("guild1", "ch1"))

	s.AddTextChannel("guild1", "ch1")

	// 追加後はtrue
	assert.True(t, s.IsTextChannelActive("guild1", "ch1"))
	// 別チャンネルはfalse
	assert.False(t, s.IsTextChannelActive("guild1", "ch2"))
	// 別guildはfalse
	assert.False(t, s.IsTextChannelActive("guild2", "ch1"))
}

func TestState_RemoveTextChannel(t *testing.T) {
	s := NewState()
	s.AddTextChannel("guild1", "ch1")
	s.AddTextChannel("guild1", "ch2")

	s.RemoveTextChannel("guild1", "ch1")

	assert.False(t, s.IsTextChannelActive("guild1", "ch1"))
	assert.True(t, s.IsTextChannelActive("guild1", "ch2"))
}

func TestState_RemoveTextChannel_NonExistentGuild(t *testing.T) {
	s := NewState()

	// 存在しないguildIDでもpanicしない
	assert.NotPanics(t, func() {
		s.RemoveTextChannel("nonexistent", "ch1")
	})
}

func TestState_IsTextChannelActive_NonExistentGuild(t *testing.T) {
	s := NewState()

	// 存在しないguildIDでもpanicせずfalseを返す
	assert.False(t, s.IsTextChannelActive("nonexistent", "ch1"))
}

func TestState_GetGuildCount(t *testing.T) {
	s := NewState()
	assert.Equal(t, 0, s.GetGuildCount())

	s.AddTextChannel("guild1", "ch1")
	assert.Equal(t, 1, s.GetGuildCount())

	s.AddTextChannel("guild2", "ch1")
	assert.Equal(t, 2, s.GetGuildCount())

	// 同じguildに複数チャンネル追加してもguild数は増えない
	s.AddTextChannel("guild1", "ch2")
	assert.Equal(t, 2, s.GetGuildCount())
}

func TestState_MultipleChannelsPerGuild(t *testing.T) {
	s := NewState()

	channels := []string{"ch1", "ch2", "ch3"}
	for _, ch := range channels {
		s.AddTextChannel("guild1", ch)
	}

	for _, ch := range channels {
		assert.True(t, s.IsTextChannelActive("guild1", ch))
	}

	// ch2を削除してもch1, ch3は残る
	s.RemoveTextChannel("guild1", "ch2")
	assert.True(t, s.IsTextChannelActive("guild1", "ch1"))
	assert.False(t, s.IsTextChannelActive("guild1", "ch2"))
	assert.True(t, s.IsTextChannelActive("guild1", "ch3"))
}

func TestState_ConcurrentAccess(t *testing.T) {
	s := NewState()
	const goroutines = 50

	var wg sync.WaitGroup

	// 並行でAddTextChannel
	for i := range goroutines {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			guildID := "guild1"
			channelID := string(rune('A' + i%26))
			s.AddTextChannel(guildID, channelID)
		}()
	}

	// 並行でIsTextChannelActive
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.IsTextChannelActive("guild1", "ch1")
		}()
	}

	// 並行でGetGuildCount
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.GetGuildCount()
		}()
	}

	wg.Wait()
}

func TestState_ConcurrentAddRemove(t *testing.T) {
	s := NewState()
	var wg sync.WaitGroup

	for i := range 100 {
		i := i
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.AddTextChannel("guild1", string(rune('a'+i%26)))
		}()
		go func() {
			defer wg.Done()
			s.RemoveTextChannel("guild1", string(rune('a'+i%26)))
		}()
	}

	wg.Wait()
}
