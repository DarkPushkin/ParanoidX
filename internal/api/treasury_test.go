// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestTreasuryDataDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "api-treasury-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}


// TestProofOfReserveHandler handles the TestProofOfReserveHandler HTTP request.
func TestProofOfReserveHandler(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "silver_reserve_ng.txt"), []byte("5000000000"), 0600)

	handler := ProofOfReserveHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/api/treasury/proof-of-reserve", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["reserve_ng"] != float64(5000000000) {
		t.Fatalf("expected reserve 5000000000, got %v", resp["reserve_ng"])
	}
	if resp["solvent"] != true {
		t.Fatal("expected solvent=true with no liabilities")
	}
}


// TestProofOfReserveHandlerRemoteForbidden handles the TestProofOfReserveHandlerRemoteForbidden HTTP request.
func TestProofOfReserveHandlerRemoteForbidden(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	handler := ProofOfReserveHandler(dir)

	req := httptest.NewRequest("GET", "/api/treasury/proof-of-reserve", nil)
	req.RemoteAddr = "8.8.8.8:12345"
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}


// TestProofOfReserveWithLiabilities handles the TestProofOfReserveWithLiabilities HTTP request.
func TestProofOfReserveWithLiabilities(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "silver_reserve_ng.txt"), []byte("1000000000"), 0600)

	banknotes := []map[string]any{
		{"serial": "SN001", "holder": "alice", "status": "active", "frozen_ng": float64(500000000)},
		{"serial": "SN002", "holder": "bob", "status": "active", "frozen_ng": float64(500000000)},
	}
	b, _ := json.Marshal(banknotes)
	os.WriteFile(filepath.Join(dir, "banknotes_registry_v2.json"), b, 0600)

	handler := ProofOfReserveHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/api/treasury/proof-of-reserve", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["solvent"] != true {
		t.Fatal("expected solvent=true (1B reserve / 1B liabilities)")
	}
	if resp["active_banknotes"] != float64(2) {
		t.Fatalf("expected 2 active banknotes, got %v", resp["active_banknotes"])
	}
}


// TestTreasuryStateHandler handles the TestTreasuryStateHandler HTTP request.
func TestTreasuryStateHandler(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "silver_reserve_ng.txt"), []byte("3000000000"), 0600)
	os.WriteFile(filepath.Join(dir, "silver_rounds.log"), []byte("round 100: usdt=500 new_ng=500000000\n"), 0644)
	os.WriteFile(filepath.Join(dir, "royal.enabled"), []byte("1"), 0644)

	handler := TreasuryStateHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/api/treasury/state", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["is_royal"] != true {
		t.Fatal("expected is_royal=true")
	}
	if resp["current_reserve_ng"] != float64(3000000000) {
		t.Fatalf("expected 3000000000, got %v", resp["current_reserve_ng"])
	}
}


// TestTreasuryStateHandlerNotRoyal handles the TestTreasuryStateHandlerNotRoyal HTTP request.
func TestTreasuryStateHandlerNotRoyal(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "silver_reserve_ng.txt"), []byte("1000000000"), 0600)

	handler := TreasuryStateHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/api/treasury/state", ""))

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["is_royal"] != false {
		t.Fatal("expected is_royal=false")
	}
}


// TestTreasuryStateNoReserveFile handles the TestTreasuryStateNoReserveFile HTTP request.
func TestTreasuryStateNoReserveFile(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	handler := TreasuryStateHandler(dir)
	// create empty banknotes_registry.json to avoid read errors
	os.WriteFile(filepath.Join(dir, "banknotes_registry.json"), []byte("[]"), 0600)

	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/api/treasury/state", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}


// TestClaimDividendsHandler handles the TestClaimDividendsHandler HTTP request.
func TestClaimDividendsHandler(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	banknotes := []map[string]any{
		{
			"serial":  "BN001",
			"holder":  "alice",
			"status":  "active",
			"frozen_ng": float64(1000000000),
			"dividend_history": []map[string]any{
				{"round": float64(1), "ng": float64(50000)},
			},
		},
	}
	b, _ := json.Marshal(banknotes)
	os.WriteFile(filepath.Join(dir, "banknotes_registry_v2.json"), b, 0600)

	handler := ClaimDividendsHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/api/treasury/claim-dividends?serial=BN001&holder=alice", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["claimed"] != float64(50000) {
		t.Fatalf("expected 50000 claimed, got %v", resp["claimed"])
	}

	// Verify dividend_history was cleared
	b, _ = os.ReadFile(filepath.Join(dir, "banknotes_registry_v2.json"))
	var updated []map[string]any
	json.Unmarshal(b, &updated)
	dh := updated[0]["dividend_history"].([]any)
	if len(dh) != 0 {
		t.Fatal("expected dividend_history to be cleared after claim")
	}
}


// TestClaimDividendsHandlerNotFound handles the TestClaimDividendsHandlerNotFound HTTP request.
func TestClaimDividendsHandlerNotFound(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	os.WriteFile(filepath.Join(dir, "banknotes_registry.json"), []byte("[]"), 0600)

	handler := ClaimDividendsHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/api/treasury/claim-dividends?serial=NONE&holder=nobody", ""))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}


// TestClaimDividendsHandlerMissingParams handles the TestClaimDividendsHandlerMissingParams HTTP request.
func TestClaimDividendsHandlerMissingParams(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	handler := ClaimDividendsHandler(dir)

	// missing serial
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/api/treasury/claim-dividends?holder=alice", ""))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 missing serial, got %d", w.Code)
	}

	// missing holder
	w = httptest.NewRecorder()
	handler(w, localReq("POST", "/api/treasury/claim-dividends?serial=BN001", ""))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 missing holder, got %d", w.Code)
	}
}


// TestClaimDividendsNoDividends handles the TestClaimDividendsNoDividends HTTP request.
func TestClaimDividendsNoDividends(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	banknotes := []map[string]any{
		{
			"serial":  "BN001",
			"holder":  "alice",
			"status":  "active",
			"frozen_ng": float64(1000000000),
			"dividend_history": []map[string]any{},
		},
	}
	b, _ := json.Marshal(banknotes)
	os.WriteFile(filepath.Join(dir, "banknotes_registry_v2.json"), b, 0600)

	handler := ClaimDividendsHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/api/treasury/claim-dividends?serial=BN001&holder=alice", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "no dividends") {
		t.Fatalf("expected 'no dividends' note, got: %s", w.Body.String())
	}
}
