// Package webrtc provides WebRTC signaling and peer management
package webrtc

import (
	"testing"
)


// TestNewSignalState handles the TestNewSignalState HTTP request.
func TestNewSignalState(t *testing.T) {
	s := NewSignalState()
	if s == nil {
		t.Fatal("NewSignalState returned nil")
	}
	state := s.GetState("nonexistent")
	if state == nil {
		t.Fatal("GetState returned nil for nonexistent room")
	}
	cands, ok := state["candidates"].([]string)
	if !ok {
		t.Fatal("expected candidates to be []string")
	}
	if len(cands) != 0 {
		t.Fatalf("expected empty candidates, got %d", len(cands))
	}
}


// TestPostSignalOffer handles the TestPostSignalOffer HTTP request.
func TestPostSignalOffer(t *testing.T) {
	s := NewSignalState()
	s.PostSignal("room1", map[string]any{"offer": "sdp-offer-123"})
	state := s.GetState("room1")
	offer, ok := state["offer"].(string)
	if !ok {
		t.Fatal("expected offer as string")
	}
	if offer != "sdp-offer-123" {
		t.Fatalf("expected 'sdp-offer-123', got '%s'", offer)
	}
}


// TestPostSignalAnswer handles the TestPostSignalAnswer HTTP request.
func TestPostSignalAnswer(t *testing.T) {
	s := NewSignalState()
	s.PostSignal("room1", map[string]any{"answer": "sdp-answer-456"})
	state := s.GetState("room1")
	ans, ok := state["answer"].(string)
	if !ok {
		t.Fatal("expected answer as string")
	}
	if ans != "sdp-answer-456" {
		t.Fatalf("expected 'sdp-answer-456', got '%s'", ans)
	}
}


// TestPostSignalCandidate handles the TestPostSignalCandidate HTTP request.
func TestPostSignalCandidate(t *testing.T) {
	s := NewSignalState()
	s.PostSignal("room2", map[string]any{"candidate": "cand-1"})
	state := s.GetState("room2")
	cands, ok := state["candidates"].([]string)
	if !ok {
		t.Fatal("expected candidates as []string")
	}
	if len(cands) != 1 || cands[0] != "cand-1" {
		t.Fatalf("expected [cand-1], got %v", cands)
	}
}


// TestPostSignalCandidatesBatch handles the TestPostSignalCandidatesBatch HTTP request.
func TestPostSignalCandidatesBatch(t *testing.T) {
	s := NewSignalState()
	s.PostSignal("room3", map[string]any{"candidates": []any{"cand-a", "cand-b", "cand-c"}})
	state := s.GetState("room3")
	cands, ok := state["candidates"].([]string)
	if !ok {
		t.Fatal("expected candidates as []string")
	}
	if len(cands) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(cands))
	}
}


// TestGetStateIsolation handles the TestGetStateIsolation HTTP request.
func TestGetStateIsolation(t *testing.T) {
	s := NewSignalState()
	s.PostSignal("roomA", map[string]any{"offer": "offer-A"})
	s.PostSignal("roomB", map[string]any{"offer": "offer-B"})

	stateA := s.GetState("roomA")
	stateB := s.GetState("roomB")

	offerA, _ := stateA["offer"].(string)
	offerB, _ := stateB["offer"].(string)
	if offerA != "offer-A" || offerB != "offer-B" {
		t.Fatalf("expected isolated rooms, got A=%s B=%s", offerA, offerB)
	}
}


// TestEmptyOfferNotStored handles the TestEmptyOfferNotStored HTTP request.
func TestEmptyOfferNotStored(t *testing.T) {
	s := NewSignalState()
	s.PostSignal("roomX", map[string]any{"offer": ""})
	state := s.GetState("roomX")
	_, hasOffer := state["offer"]
	if hasOffer {
		t.Fatal("empty offer should not be stored")
	}
}


// TestEmptyCandidateNotStored handles the TestEmptyCandidateNotStored HTTP request.
func TestEmptyCandidateNotStored(t *testing.T) {
	s := NewSignalState()
	s.PostSignal("roomY", map[string]any{"candidate": ""})
	state := s.GetState("roomY")
	cands, _ := state["candidates"].([]string)
	if len(cands) != 0 {
		t.Fatal("empty candidate should not be appended")
	}
}


// TestICEConfigGenerateCredentials handles the TestICEConfigGenerateCredentials HTTP request.
func TestICEConfigGenerateCredentials(t *testing.T) {
	c := &ICEConfig{Secret: "test-secret", Onion: "test.onion"}
	user, cred, exp := c.GenerateCredentials()
	if user == "" {
		t.Fatal("username should not be empty")
	}
	if cred == "" {
		t.Fatal("credential should not be empty")
	}
	if exp == 0 {
		t.Fatal("expiry should not be zero")
	}
}


// TestICEConfigGenerateConfig handles the TestICEConfigGenerateConfig HTTP request.
func TestICEConfigGenerateConfig(t *testing.T) {
	c := &ICEConfig{Secret: "test-secret", Onion: "test.onion"}
	cfg := c.GenerateConfig()
	servers, ok := cfg["iceServers"].([]map[string]any)
	if !ok {
		t.Fatal("expected iceServers as []map[string]any")
	}
	if len(servers) == 0 {
		t.Fatal("expected at least one ICE server")
	}
	urls, ok := servers[0]["urls"].([]string)
	if !ok {
		t.Fatal("expected urls as []string")
	}
	if len(urls) < 2 {
		t.Fatal("expected at least 2 URLs")
	}
	if urls[0] != "turn:test.onion:3478?transport=tcp" {
		t.Fatalf("unexpected turn URL: %s", urls[0])
	}
}


// TestICEConfigGenerateConfigNoOnion handles the TestICEConfigGenerateConfigNoOnion HTTP request.
func TestICEConfigGenerateConfigNoOnion(t *testing.T) {
	c := &ICEConfig{Secret: "", Onion: ""}
	cfg := c.GenerateConfig()
	servers, ok := cfg["iceServers"].([]any)
	if !ok {
		t.Fatal("expected iceServers as []any")
	}
	if len(servers) != 0 {
		t.Fatal("expected empty iceServers when no onion")
	}
}


// TestICEConfigGenerateConfigDefaultSecret handles the TestICEConfigGenerateConfigDefaultSecret HTTP request.
func TestICEConfigGenerateConfigDefaultSecret(t *testing.T) {
	c := &ICEConfig{Secret: "", Onion: "default-test.onion"}
	cfg := c.GenerateConfig()
	servers, ok := cfg["iceServers"].([]map[string]any)
	if !ok {
		t.Fatal("expected iceServers as []map[string]any")
	}
	if len(servers) == 0 {
		t.Fatal("expected ICE servers even with empty secret")
	}
	_, hasUser := servers[0]["username"]
	_, hasCred := servers[0]["credential"]
	if !hasUser || !hasCred {
		t.Fatal("expected username and credential even with default secret")
	}
}


// TestPasteLines handles the TestPasteLines HTTP request.
func TestPasteLines(t *testing.T) {
	c := &ICEConfig{Secret: "s", Onion: "paste-test.onion"}
	cfg := c.GenerateConfig()
	lines, ok := cfg["pasteLines"].([]string)
	if !ok {
		t.Fatal("expected pasteLines as []string")
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 paste lines, got %d", len(lines))
	}
	if len(lines[0]) == 0 || len(lines[1]) == 0 {
		t.Fatal("paste lines should not be empty")
	}
}


// TestSignalStateConcurrency handles the TestSignalStateConcurrency HTTP request.
func TestSignalStateConcurrency(t *testing.T) {
	s := NewSignalState()
	done := make(chan bool)
	n := 50

	for i := 0; i < n; i++ {
		go func(id int) {
			s.PostSignal("concurrent", map[string]any{
				"candidate": "cand",
				"offer":     "offer",
			})
			done <- true
		}(i)
	}

	for i := 0; i < n; i++ {
		<-done
	}

	state := s.GetState("concurrent")
	_, hasOffer := state["offer"].(string)
	if !hasOffer {
		t.Fatal("expected offer after concurrent writes")
	}
}
