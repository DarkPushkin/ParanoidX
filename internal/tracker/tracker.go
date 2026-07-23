// Package tracker implements a BitTorrent-style tracker for radio track distribution.
// Peers announce which .isle pieces they have and discover other seeders.
package tracker

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var validInfohash = regexp.MustCompile(`^[a-fA-F0-9]{40,}$`)

type PeerInfo struct {
	Addr    string `json:"addr"`
	ID      string `json:"id"`
	Pieces  string `json:"pieces"` // bitfield hex
	Updated int64  `json:"updated"`
}

type Tracker struct {
	mu        sync.RWMutex
	swarms    map[string]map[string]*PeerInfo // trackID -> peerID -> PeerInfo
}


// New handles the New HTTP request.
func New() *Tracker {
	t := &Tracker{swarms: make(map[string]map[string]*PeerInfo)}
	go t.cleanLoop()
	return t
}


// Announce handles the Announce HTTP request.
func (t *Tracker) Announce(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TrackID string `json:"track_id"`
		PeerID  string `json:"peer_id"`
		Addr    string `json:"addr"`
		Pieces  string `json:"pieces"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if req.TrackID == "" || req.PeerID == "" {
		http.Error(w, "track_id and peer_id required", http.StatusBadRequest)
		return
	}

	t.mu.Lock()
	if t.swarms[req.TrackID] == nil {
		t.swarms[req.TrackID] = make(map[string]*PeerInfo)
	}
	t.swarms[req.TrackID][req.PeerID] = &PeerInfo{
		Addr:    req.Addr,
		ID:      req.PeerID,
		Pieces:  req.Pieces,
		Updated: time.Now().Unix(),
	}
	t.mu.Unlock()

	// Return other peers in the swarm
	t.mu.RLock()
	swarm := t.swarms[req.TrackID]
	var peers []PeerInfo
	for id, p := range swarm {
		if id != req.PeerID {
			peers = append(peers, *p)
		}
	}
	t.mu.RUnlock()

	writeJSON(w, map[string]any{
		"ok":      true,
		"track":   req.TrackID,
		"peers":   peers,
		"count":   len(peers),
	})
}


// Scrape handles the Scrape HTTP request.
func (t *Tracker) Scrape(w http.ResponseWriter, r *http.Request) {
	trackID := r.URL.Query().Get("track")

	if trackID != "" {
		if !validInfohash.MatchString(trackID) {
			http.Error(w, `{"error":"invalid infohash: must be 40+ hex characters"}`, http.StatusBadRequest)
			return
		}
		t.mu.RLock()
		swarm := t.swarms[trackID]
		count := len(swarm)
		t.mu.RUnlock()
		writeJSON(w, map[string]any{
			"track":    trackID,
			"seeders":  count,
			"complete": count,
		})
		return
	}

	t.mu.RLock()
	result := make(map[string]int)
	for id, swarm := range t.swarms {
		result[id] = len(swarm)
	}
	t.mu.RUnlock()
	writeJSON(w, map[string]any{"swarms": result, "count": len(result)})
}

// Nodes returns all unique peers across all swarms.
func (t *Tracker) Nodes(w http.ResponseWriter, r *http.Request) {
	t.mu.RLock()
	seen := make(map[string]bool)
	var nodes []PeerInfo
	for _, swarm := range t.swarms {
		for _, p := range swarm {
			if !seen[p.ID] {
				seen[p.ID] = true
				nodes = append(nodes, *p)
			}
		}
	}
	t.mu.RUnlock()
	writeJSON(w, map[string]any{"count": len(nodes), "nodes": nodes})
}

func (t *Tracker) cleanLoop() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		now := time.Now().Unix()
		t.mu.Lock()
		for trackID, swarm := range t.swarms {
			for peerID, p := range swarm {
				if now-p.Updated > 120 { // 2 min timeout
					delete(swarm, peerID)
				}
			}
			if len(swarm) == 0 {
				delete(t.swarms, trackID)
			}
		}
		t.mu.Unlock()
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
