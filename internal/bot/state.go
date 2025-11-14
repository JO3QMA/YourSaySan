package bot

import (
	"sync"
	"sync/atomic"
	"time"
)

type GuildState struct {
	// VC接続情報
	VoiceConn      interface{} // *voice.Connection（循環参照を避けるためinterface{}）
	VoiceChannelID string

	// 読み上げ対象チャンネル（現在は単一チャンネルのみ対応）
	// 将来的な拡張（複数チャンネル対応）を見越してmapで実装
	TextChannelIDs map[string]bool // channelID -> bool

	// 読み上げ状態
	IsReading    atomic.Bool // Go 1.19+で導入
	LastActivity time.Time

	// ロック
	mu sync.RWMutex
}

type State struct {
	// ギルドごとの状態管理
	Guilds map[string]*GuildState // guildID -> state
	mu     sync.RWMutex
}

func NewState() *State {
	return &State{
		Guilds: make(map[string]*GuildState),
	}
}

func (s *State) GetGuildState(guildID string) *GuildState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.Guilds[guildID]
	if !ok {
		return nil
	}
	return state
}

func (s *State) SetGuildState(guildID string, state *GuildState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Guilds[guildID] = state
}

func (s *State) RemoveGuildState(guildID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Guilds, guildID)
}

func (s *State) IsTextChannelActive(guildID, channelID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.Guilds[guildID]
	if !ok {
		return false
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	return state.TextChannelIDs[channelID]
}

func (s *State) AddTextChannel(guildID, channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.Guilds[guildID]
	if !ok {
		state = &GuildState{
			TextChannelIDs: make(map[string]bool),
		}
		s.Guilds[guildID] = state
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.TextChannelIDs == nil {
		state.TextChannelIDs = make(map[string]bool)
	}
	state.TextChannelIDs[channelID] = true
}

func (s *State) RemoveTextChannel(guildID, channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.Guilds[guildID]
	if !ok {
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	delete(state.TextChannelIDs, channelID)
}

func (s *State) GetGuildCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Guilds)
}
