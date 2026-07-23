// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)


// TestRoyalControlNotRoyal handles the TestRoyalControlNotRoyal HTTP request.
func TestRoyalControlNotRoyal(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	handler := RoyalControlHandler(dir)

	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/royal/control?action=heartbeat&sub_id=node1", ""))

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}


// TestRoyalControlRegisterSub handles the TestRoyalControlRegisterSub HTTP request.
func TestRoyalControlRegisterSub(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)

	// Register a sub-node
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/royal/control?action=register-sub", `{"id":"sub1","name":"Island North","address":"smp://north"}`))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	sub := resp["sub"].(map[string]any)
	if sub["id"] != "sub1" {
		t.Fatalf("expected sub id sub1, got %v", sub["id"])
	}
}


// TestRoyalListSubs handles the TestRoyalListSubs HTTP request.
func TestRoyalListSubs(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)
	handler(httptest.NewRecorder(), localReq("POST", "/royal/control?action=register-sub", `{"id":"sub_a","name":"Alpha","address":"a"}`))
	handler(httptest.NewRecorder(), localReq("POST", "/royal/control?action=register-sub", `{"id":"sub_b","name":"Beta","address":"b"}`))

	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/royal/control?action=list-subs", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"] != float64(2) {
		t.Fatalf("expected 2 subs, got %v", resp["count"])
	}
}


// TestRoyalRelayCommand handles the TestRoyalRelayCommand HTTP request.
func TestRoyalRelayCommand(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)

	// Relay a command
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/royal/control?action=relay&sub_id=sub1&cmd=restart", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	relay := resp["relay"].(map[string]any)
	if relay["command"] != "restart" {
		t.Fatalf("expected restart, got %v", relay["command"])
	}
	if relay["status"] != "pending" {
		t.Fatalf("expected pending, got %v", relay["status"])
	}
}


// TestRoyalPollAndAck handles the TestRoyalPollAndAck HTTP request.
func TestRoyalPollAndAck(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)

	// Send relay and poll
	handler(httptest.NewRecorder(), localReq("POST", "/royal/control?action=relay&sub_id=subX&cmd=status", ""))

	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/royal/control?action=poll&sub_id=subX", ""))

	var pollResp map[string]any
	json.Unmarshal(w.Body.Bytes(), &pollResp)
	if pollResp["count"] != float64(1) {
		t.Fatalf("expected 1 pending relay, got %v", pollResp["count"])
	}

	relays := pollResp["relays"].([]any)
	relay := relays[0].(map[string]any)
	relayID := relay["id"].(string)

	// Ack
	w2 := httptest.NewRecorder()
	handler(w2, localReq("POST", "/royal/control?action=ack&relay_id="+relayID+"&result=done", ""))

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on ack, got %d", w2.Code)
	}

	// Verify no longer pending
	w3 := httptest.NewRecorder()
	handler(w3, localReq("GET", "/royal/control?action=poll&sub_id=subX", ""))

	json.Unmarshal(w3.Body.Bytes(), &pollResp)
	if pollResp["count"] != float64(0) {
		t.Fatalf("expected 0 pending after ack, got %v", pollResp["count"])
	}
}


// TestRoyalHeartbeat handles the TestRoyalHeartbeat HTTP request.
func TestRoyalHeartbeat(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)
	handler(httptest.NewRecorder(), localReq("POST", "/royal/control?action=register-sub", `{"id":"node1","name":"Heartbeat Node"}`))

	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/royal/control?action=heartbeat&sub_id=node1", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Check status changed
	w2 := httptest.NewRecorder()
	handler(w2, localReq("GET", "/royal/control?action=list-subs", ""))

	var resp map[string]any
	json.Unmarshal(w2.Body.Bytes(), &resp)
	subs := resp["subs"].([]any)
	sub := subs[0].(map[string]any)
	if sub["status"] != "online" {
		t.Fatalf("expected online status, got %v", sub["status"])
	}
}


// TestRoyalRelayMissingFields handles the TestRoyalRelayMissingFields HTTP request.
func TestRoyalRelayMissingFields(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)

	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/royal/control?action=relay", ""))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 missing fields, got %d", w.Code)
	}
}


// TestRoyalSync handles the TestRoyalSync HTTP request.
func TestRoyalSync(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/royal/sync", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}


// TestRoyalUnknownAction handles the TestRoyalUnknownAction HTTP request.
func TestRoyalUnknownAction(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := RoyalControlHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/royal/control?action=nonexistent", ""))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown action, got %d", w.Code)
	}
}
