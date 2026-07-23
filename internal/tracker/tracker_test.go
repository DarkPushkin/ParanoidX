// Package tracker implements BitTorrent-style tracker for P2P discovery
package tracker

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)


// TestNew handles the TestNew HTTP request.
func TestNew(t *testing.T) {
	tr := New()
	if tr == nil {
		t.Fatal("expected non-nil tracker")
	}
	if tr.swarms == nil {
		t.Fatal("expected swarms map")
	}
}


// TestAnnounce handles the TestAnnounce HTTP request.
func TestAnnounce(t *testing.T) {
	tr := New()
	body := `{"track_id":"abc123","peer_id":"peer1","addr":"1.2.3.4:9000","pieces":"aabb"}`
	req := httptest.NewRequest("POST", "/announce", strings.NewReader(body))
	w := httptest.NewRecorder()
	tr.Announce(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatal("expected ok=true")
	}
	if resp["track"] != "abc123" {
		t.Fatal("wrong track id")
	}
	if resp["count"].(float64) != 0 {
		t.Fatal("expected 0 peers")
	}
}


// TestAnnounceTwoPeers handles the TestAnnounceTwoPeers HTTP request.
func TestAnnounceTwoPeers(t *testing.T) {
	tr := New()
	body1 := `{"track_id":"t1","peer_id":"p1","addr":"1.2.3.4:9000","pieces":"ff"}`
	w1 := httptest.NewRecorder()
	tr.Announce(w1, httptest.NewRequest("POST", "/announce", strings.NewReader(body1)))

	body2 := `{"track_id":"t1","peer_id":"p2","addr":"5.6.7.8:9000","pieces":"00"}`
	w2 := httptest.NewRecorder()
	tr.Announce(w2, httptest.NewRequest("POST", "/announce", strings.NewReader(body2)))

	var resp map[string]any
	json.NewDecoder(w2.Body).Decode(&resp)
	count := resp["count"].(float64)
	if count != 1 {
		t.Fatalf("expected 1 peer, got %v", count)
	}
	peers := resp["peers"].([]any)
	if len(peers) != 1 {
		t.Fatal("expected 1 peer in list")
	}
	peer := peers[0].(map[string]any)
	if peer["id"] != "p1" {
		t.Fatal("expected peer p1")
	}
}


// TestAnnounceInvalidMethod handles the TestAnnounceInvalidMethod HTTP request.
func TestAnnounceInvalidMethod(t *testing.T) {
	tr := New()
	w := httptest.NewRecorder()
	tr.Announce(w, httptest.NewRequest("GET", "/announce", nil))
	if w.Code != 405 {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}


// TestAnnounceMissingFields handles the TestAnnounceMissingFields HTTP request.
func TestAnnounceMissingFields(t *testing.T) {
	tr := New()
	body := `{"track_id":"t1","peer_id":""}`
	w := httptest.NewRecorder()
	tr.Announce(w, httptest.NewRequest("POST", "/announce", strings.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}


// TestScrapeAll handles the TestScrapeAll HTTP request.
func TestScrapeAll(t *testing.T) {
	tr := New()
	post := func(track, peer, addr string) {
		body, _ := json.Marshal(map[string]string{"track_id": track, "peer_id": peer, "addr": addr, "pieces": "ff"})
		tr.Announce(httptest.NewRecorder(), httptest.NewRequest("POST", "/announce", strings.NewReader(string(body))))
	}
	post("t1", "p1", "1.2.3.4:9000")
	post("t1", "p2", "5.6.7.8:9000")
	post("t2", "p3", "9.10.11.12:9000")

	req := httptest.NewRequest("GET", "/scrape", nil)
	w := httptest.NewRecorder()
	tr.Scrape(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	swarms := resp["swarms"].(map[string]any)
	if len(swarms) != 2 {
		t.Fatalf("expected 2 swarms, got %d", len(swarms))
	}
	if resp["count"].(float64) != 2 {
		t.Fatalf("expected count 2, got %v", resp["count"])
	}
}


// TestScrapeSingleTrack handles the TestScrapeSingleTrack HTTP request.
func TestScrapeSingleTrack(t *testing.T) {
	tr := New()
	post := func(track, peer, addr string) {
		body, _ := json.Marshal(map[string]string{"track_id": track, "peer_id": peer, "addr": addr, "pieces": "ff"})
		tr.Announce(httptest.NewRecorder(), httptest.NewRequest("POST", "/announce", strings.NewReader(string(body))))
	}
	ih1 := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0"
	ih2 := "f0e9d8c7b6a5e4f3c2b1a0f9e8d7c6b5a4f3e2d1"
	post(ih1, "p1", "1.2.3.4:9000")
	post(ih2, "p2", "5.6.7.8:9000")

	req := httptest.NewRequest("GET", "/scrape?track="+ih1, nil)
	w := httptest.NewRecorder()
	tr.Scrape(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["track"] != ih1 {
		t.Fatal("expected track " + ih1)
	}
	if resp["seeders"].(float64) != 1 {
		t.Fatal("expected 1 seeder")
	}
}


// TestScrapeMissing handles the TestScrapeMissing HTTP request.
func TestScrapeMissing(t *testing.T) {
	tr := New()
	req := httptest.NewRequest("GET", "/scrape", nil)
	w := httptest.NewRecorder()
	tr.Scrape(w, req)
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["count"].(float64) != 0 {
		t.Fatal("expected count 0 for empty swarms")
	}
}


// TestCleanLoopStaleLogic handles the TestCleanLoopStaleLogic HTTP request.
func TestCleanLoopStaleLogic(t *testing.T) {
	tr := New()
	tr.mu.Lock()
	tr.swarms["stale"] = map[string]*PeerInfo{
		"p1": {ID: "p1", Addr: "1.2.3.4:9000", Updated: time.Now().Add(-2 * time.Hour).Unix()},
		"p2": {ID: "p2", Addr: "5.6.7.8:9000", Updated: time.Now().Unix()},
	}
	tr.mu.Unlock()

	// Simulate cleanLoop logic manually
	now := time.Now().Unix()
	tr.mu.Lock()
	for trackID, swarm := range tr.swarms {
		for peerID, p := range swarm {
			if now-p.Updated > 120 {
				delete(swarm, peerID)
			}
		}
		if len(swarm) == 0 {
			delete(tr.swarms, trackID)
		}
	}
	tr.mu.Unlock()

	tr.mu.RLock()
	swarm, exists := tr.swarms["stale"]
	tr.mu.RUnlock()
	if !exists {
		t.Fatal("expected track to still exist (p2 is fresh)")
	}
	if len(swarm) != 1 || swarm["p2"] == nil {
		t.Fatal("expected only p2 to remain, stale p1 removed")
	}
	if swarm["p1"] != nil {
		t.Fatal("expected stale peer p1 removed")
	}
}


// TestNodes handles the TestNodes HTTP request.
func TestNodes(t *testing.T) {
	tr := New()
	post := func(track, peer, addr string) {
		body, _ := json.Marshal(map[string]string{"track_id": track, "peer_id": peer, "addr": addr, "pieces": "ff"})
		tr.Announce(httptest.NewRecorder(), httptest.NewRequest("POST", "/announce", strings.NewReader(string(body))))
	}
	post("t1", "p1", "1.2.3.4:9000")
	post("t1", "p2", "5.6.7.8:9000")
	post("t2", "p1", "1.2.3.4:9000") // same peer in different swarm — should dedupe

	req := httptest.NewRequest("GET", "/nodes", nil)
	w := httptest.NewRecorder()
	tr.Nodes(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["count"].(float64) != 2 {
		t.Fatalf("expected 2 unique nodes, got %v", resp["count"])
	}
}
