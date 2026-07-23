// Package radio provides the radio streaming and scheduling system
package radio

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfflineCache manages pre-downloaded radio tracks for offline playback.
type OfflineCache struct {
	mu       sync.RWMutex
	DataDir  string
	baseURL  string
	cache    map[string]string // trackID -> local file path
}

type cachedTrack struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	RemoteURL  string    `json:"remote_url"`
	LocalPath  string    `json:"local_path"`
	Downloaded bool      `json:"downloaded"`
	AddedAt    time.Time `json:"added_at"`
}


// NewOfflineCache handles the NewOfflineCache HTTP request.
func NewOfflineCache(dataDir, baseURL string) *OfflineCache {
	oc := &OfflineCache{
		DataDir:  filepath.Join(dataDir, "radio", "offline"),
		baseURL:  baseURL,
		cache:    make(map[string]string),
	}
	os.MkdirAll(oc.DataDir, 0755)
	oc.loadIndex()
	return oc
}


// PreDownload handles the PreDownload HTTP request.
func (oc *OfflineCache) PreDownload(tracks []Track) int {
	var count int
	for _, t := range tracks {
		if t.IsAI {
			continue // AI tracks are already local
		}
		localPath := filepath.Join(oc.DataDir, t.ID+".mp3")
		if _, err := os.Stat(localPath); err == nil {
			oc.mu.Lock()
			oc.cache[t.ID] = localPath
			oc.mu.Unlock()
			count++
			continue
		}
		remoteURL := oc.baseURL + "/api/radio/track?path=" + t.FilePath
		if err := oc.download(remoteURL, localPath); err != nil {
			continue
		}
		oc.mu.Lock()
		oc.cache[t.ID] = localPath
		oc.mu.Unlock()
		count++
	}
	oc.saveIndex()
	return count
}

func (oc *OfflineCache) download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}


// GetLocalPath handles the GetLocalPath HTTP request.
func (oc *OfflineCache) GetLocalPath(trackID string) string {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return oc.cache[trackID]
}


// IsOffline handles the IsOffline HTTP request.
func (oc *OfflineCache) IsOffline(trackID string) bool {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	_, ok := oc.cache[trackID]
	return ok
}


// Clear handles the Clear HTTP request.
func (oc *OfflineCache) Clear() {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.cache = make(map[string]string)
	os.RemoveAll(oc.DataDir)
	os.MkdirAll(oc.DataDir, 0755)
}


// Stats handles the Stats HTTP request.
func (oc *OfflineCache) Stats() map[string]int {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return map[string]int{"cached": len(oc.cache)}
}

func (oc *OfflineCache) loadIndex() {
	p := filepath.Join(oc.DataDir, "index.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return
	}
	var entries []cachedTrack
	json.Unmarshal(b, &entries)
	for _, e := range entries {
		if e.Downloaded {
			oc.cache[e.ID] = e.LocalPath
		}
	}
}

func (oc *OfflineCache) saveIndex() {
	oc.mu.RLock()
	entries := make([]cachedTrack, 0, len(oc.cache))
	for id, path := range oc.cache {
		entries = append(entries, cachedTrack{
			ID: id, LocalPath: path, Downloaded: true, AddedAt: time.Now(),
		})
	}
	oc.mu.RUnlock()
	b, _ := json.MarshalIndent(entries, "", "  ")
	os.WriteFile(filepath.Join(oc.DataDir, "index.json"), b, 0644)
}
