// Package channels provides SimpleX channel management with persistence.
//
// A SimpleX channel is a broadcast medium where one creator sends messages
// to subscribers. This package manages channel metadata, local subscriptions,
// and message caching.
package channels

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Channel represents a SimpleX channel with local metadata.
type Channel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Link        string    `json:"link,omitempty"`
	Role        string    `json:"role"` // creator, subscriber
	CreatedAt   time.Time `json:"created_at"`
	LastMessage string    `json:"last_message,omitempty"`
	Unread      int       `json:"unread"`
}

// ChannelMessage represents a cached message from a channel.
type ChannelMessage struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Manager manages channels and their messages with JSON persistence.
type Manager struct {
	mu       sync.RWMutex
	DataDir  string
	Channels map[string]*Channel       `json:"channels"`
	Messages map[string][]ChannelMessage `json:"messages"` // keyed by channel ID
}


// NewManager handles the NewManager HTTP request.
func NewManager(dataDir string) *Manager {
	m := &Manager{
		DataDir:  filepath.Join(dataDir, "channels"),
		Channels: make(map[string]*Channel),
		Messages: make(map[string][]ChannelMessage),
	}
	m.load()
	return m
}

func (m *Manager) filePath() string {
	os.MkdirAll(m.DataDir, 0755)
	return filepath.Join(m.DataDir, "channels.json")
}

func (m *Manager) load() {
	p := m.filePath()
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	var stored struct {
		Channels map[string]*Channel       `json:"channels"`
		Messages map[string][]ChannelMessage `json:"messages"`
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if stored.Channels != nil {
		m.Channels = stored.Channels
	}
	if stored.Messages != nil {
		m.Messages = stored.Messages
	}
}

// saveLocked persists state to disk. Caller must hold at least RLock.
func (m *Manager) saveLocked() {
	os.MkdirAll(m.DataDir, 0755)
	data, _ := json.MarshalIndent(map[string]any{
		"channels": m.Channels,
		"messages": m.Messages,
	}, "", "  ")
	os.WriteFile(m.filePath(), data, 0644)
}

// AddChannel stores a new channel locally.
func (m *Manager) AddChannel(id, name, link, role string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Channels[id] = &Channel{
		ID:        id,
		Name:      name,
		Link:      link,
		Role:      role,
		CreatedAt: time.Now(),
	}
	m.saveLocked()
}

// GetChannel returns a channel by ID.
func (m *Manager) GetChannel(id string) *Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Channels[id]
}

// ListChannels returns all channels sorted by creation time.
func (m *Manager) ListChannels() []*Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Channel, 0, len(m.Channels))
	for _, ch := range m.Channels {
		out = append(out, ch)
	}
	return out
}

// RemoveChannel deletes a channel and its messages.
func (m *Manager) RemoveChannel(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Channels, id)
	delete(m.Messages, id)
	m.saveLocked()
}

// AddMessage caches a message for a channel.
func (m *Manager) AddMessage(channelID, text, sender string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msg := ChannelMessage{
		ID:        fmt.Sprintf("chmsg-%d", time.Now().UnixNano()),
		ChannelID: channelID,
		Text:      text,
		Sender:    sender,
		Timestamp: time.Now(),
	}
	m.Messages[channelID] = append(m.Messages[channelID], msg)
	if ch, ok := m.Channels[channelID]; ok {
		ch.LastMessage = text
		ch.Unread++
	}
	m.saveLocked()
}

// MarkRead resets the unread count for a channel.
func (m *Manager) MarkRead(channelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ch, ok := m.Channels[channelID]; ok {
		ch.Unread = 0
	}
	m.saveLocked()
}

// GetMessages returns cached messages for a channel with optional limit.
func (m *Manager) GetMessages(channelID string, limit int) []ChannelMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	msgs := m.Messages[channelID]
	if len(msgs) == 0 {
		return []ChannelMessage{}
	}
	if limit <= 0 || limit >= len(msgs) {
		out := make([]ChannelMessage, len(msgs))
		copy(out, msgs)
		return out
	}
	out := make([]ChannelMessage, limit)
	copy(out, msgs[len(msgs)-limit:])
	return out
}
