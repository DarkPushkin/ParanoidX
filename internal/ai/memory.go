// Package ai provides AI integration with Ollama, including chat, generation, and monitoring
package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"log/slog"
)

type MemoryEntry struct {
	UserID    string    `json:"user_id"`
	Messages  []Message `json:"messages"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MemoryStore struct {
	mu      sync.RWMutex
	path    string
	entries map[string]*MemoryEntry
	maxMsgs int
}


// NewMemoryStore handles the NewMemoryStore HTTP request.
func NewMemoryStore(dataDir string, maxMsgs int) *MemoryStore {
	if maxMsgs <= 0 {
		maxMsgs = 20
	}
	store := &MemoryStore{
		path:    filepath.Join(dataDir, "steward_memory.json"),
		entries: map[string]*MemoryEntry{},
		maxMsgs: maxMsgs,
	}
	store.load()
	return store
}

func (s *MemoryStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var entries []*MemoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		slog.Warn("steward memory load", "error", err)
		return
	}
	for _, e := range entries {
		s.entries[e.UserID] = e
	}
	slog.Info("steward memory loaded", "users", len(entries))
}

func (s *MemoryStore) save() {
	entries := make([]*MemoryEntry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		slog.Error("steward memory save", "error", err)
		return
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		slog.Error("steward memory save write", "error", err)
	}
}


// Add handles the Add HTTP request.
func (s *MemoryStore) Add(userID, role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[userID]
	if !ok {
		entry = &MemoryEntry{
			UserID:   userID,
			Messages: []Message{},
		}
		s.entries[userID] = entry
	}

	entry.Messages = append(entry.Messages, Message{Role: role, Content: content})
	// Trim to max
	if len(entry.Messages) > s.maxMsgs {
		entry.Messages = entry.Messages[len(entry.Messages)-s.maxMsgs:]
	}
	entry.UpdatedAt = time.Now()

	s.save()
}


// GetContext handles the GetContext HTTP request.
func (s *MemoryStore) GetContext(userID string, maxMessages int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[userID]
	if !ok {
		return nil
	}
	if maxMessages <= 0 || maxMessages >= len(entry.Messages) {
		return entry.Messages
	}
	return entry.Messages[len(entry.Messages)-maxMessages:]
}


// Clear handles the Clear HTTP request.
func (s *MemoryStore) Clear(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, userID)
	s.save()
}


// Stats handles the Stats HTTP request.
func (s *MemoryStore) Stats() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	users := make([]string, 0, len(s.entries))
	for uid, e := range s.entries {
		users = append(users, uid)
		total += len(e.Messages)
	}
	return map[string]any{
		"users":       len(s.entries),
		"total_msgs":  total,
		"user_ids":    users,
		"updated_at":  time.Now().Format(time.RFC3339),
	}
}
