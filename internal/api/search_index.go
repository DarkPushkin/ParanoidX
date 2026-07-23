// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// InvertedIndex maps keywords to a set of message IDs for fast full-text search.
type InvertedIndex struct {
	mu       sync.RWMutex
	index    map[string]map[string]struct{} // keyword → messageID set
	msgCount int
	filePath string
	hub      *ChatHub
}

// GlobalInvertedIndex is the package-level singleton search index instance.
var GlobalInvertedIndex *InvertedIndex


// NewInvertedIndex creates and initializes a new search index backed by JSON persistence.
func NewInvertedIndex(hub *ChatHub, dataDir string) *InvertedIndex {
	idx := &InvertedIndex{
		index:    make(map[string]map[string]struct{}),
		hub:      hub,
		filePath: filepath.Join(dataDir, "search_index.json"),
	}
	idx.load()
	idx.rebuild()
	slog.Info("inverted search index initialized", "path", idx.filePath)
	return idx
}

func (idx *InvertedIndex) tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '@' || r == '_')
	})
	uniq := make(map[string]struct{}, len(words))
	for _, w := range words {
		if len(w) >= 2 {
			uniq[w] = struct{}{}
		}
	}
	result := make([]string, 0, len(uniq))
	for w := range uniq {
		result = append(result, w)
	}
	return result
}


// AddMessage indexes a chat message's text and sender for future searches.
func (idx *InvertedIndex) AddMessage(msg ChatMessage) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	tokens := idx.tokenize(msg.Text)
	tokens = append(tokens, idx.tokenize(msg.From)...)
	for _, token := range tokens {
		if idx.index[token] == nil {
			idx.index[token] = make(map[string]struct{})
		}
		idx.index[token][msg.ID] = struct{}{}
	}
	idx.msgCount++
}

func (idx *InvertedIndex) rebuild() {
	msgs := idx.hub.GetMessages()
	idx.mu.Lock()
	idx.index = make(map[string]map[string]struct{})
	for _, m := range msgs {
		tokens := idx.tokenize(m.Text)
		tokens = append(tokens, idx.tokenize(m.From)...)
		for _, token := range tokens {
			if idx.index[token] == nil {
				idx.index[token] = make(map[string]struct{})
			}
			idx.index[token][m.ID] = struct{}{}
		}
	}
	idx.msgCount = len(msgs)
	idx.mu.Unlock()
}

func (idx *InvertedIndex) save() {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	data := map[string]any{
		"msg_count": idx.msgCount,
		"index":     idx.index,
	}
	b, err := json.Marshal(data)
	if err != nil {
		slog.Error("search index save marshal", "error", err)
		return
	}
	if err := os.WriteFile(idx.filePath, b, 0644); err != nil {
		slog.Error("search index save write", "error", err)
	}
}

func (idx *InvertedIndex) load() {
	b, err := os.ReadFile(idx.filePath)
	if err != nil {
		return
	}
	var data struct {
		MsgCount int                                 `json:"msg_count"`
		Index    map[string]map[string]struct{}       `json:"index"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		slog.Warn("search index load corrupt", "error", err)
		return
	}
	idx.mu.Lock()
	idx.index = data.Index
	idx.msgCount = data.MsgCount
	idx.mu.Unlock()
	slog.Info("search index loaded", "msg_count", idx.msgCount, "keywords", len(idx.index))
}


// Search returns chat messages matching the given query keyword, up to the limit.
func (idx *InvertedIndex) Search(q string, limit int) []ChatMessage {
	if q == "" {
		return nil
	}
	q = strings.ToLower(strings.TrimSpace(q))
	idx.mu.RLock()
	msgIDs, ok := idx.index[q]
	idx.mu.RUnlock()
	if !ok {
		return nil
	}
	all := idx.hub.GetMessages()
	out := make([]ChatMessage, 0, limit)
	for i := len(all) - 1; i >= 0 && len(out) < limit; i-- {
		if _, found := msgIDs[all[i].ID]; found {
			out = append(out, all[i])
		}
	}
	return out
}


// Stats returns index statistics (message count and keyword count).
func (idx *InvertedIndex) Stats() map[string]any {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return map[string]any{
		"msg_count": idx.msgCount,
		"keywords":  len(idx.index),
	}
}

// SearchIndexStatusHandler returns index status.
func SearchIndexStatusHandler(idx *InvertedIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"ok":       true,
			"status":   idx.Stats(),
		})
	}
}

// SearchIndexRebuildHandler manually rebuilds the index.
func SearchIndexRebuildHandler(idx *InvertedIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		start := time.Now()
		idx.rebuild()
		idx.save()
		elapsed := time.Since(start).Milliseconds()
		writeJSON(w, map[string]any{
			"ok":       true,
			"message":  fmt.Sprintf("index rebuilt with %d messages in %dms", idx.msgCount, elapsed),
		})
	}
}
