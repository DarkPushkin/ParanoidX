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


// TestCreateReleaseEscrow handles the TestCreateReleaseEscrow HTTP request.
func TestCreateReleaseEscrow(t *testing.T) {
	dir := newTestTreasuryDataDir(t)

	// Set up ledger with buyer funds
	ld := loadLedger(dir)
	ld.Mint("buyer1", 1000000)
	ld.Transfer("buyer1", "alice-holder", 500000) // give 500k to holder for good measure
	ld.Save(dir)

	createHandler := CreateEscrowHandler(dir)
	body := `{"buyer":"buyer1","seller":"seller1","item_id":"ITEM001","price_ng":200000}`
	w := httptest.NewRecorder()
	createHandler(w, localReq("POST", "/api/escrow/create", body))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Fatal("expected ok=true")
	}

	esc := resp["escrow"].(map[string]any)
	if esc["status"] != "active" {
		t.Fatalf("expected active escrow, got %v", esc["status"])
	}
	if esc["price_ng"] != float64(200000) {
		t.Fatalf("expected price_ng=200000, got %v", esc["price_ng"])
	}
	if esc["buyer"] != "buyer1" {
		t.Fatalf("expected buyer=buyer1, got %v", esc["buyer"])
	}
	escID := esc["id"].(string)

	// Verify buyer balance deducted (1M - 500K transfer - 200K escrow = 300K)
	ld2 := loadLedger(dir)
	if ld2.Balance("buyer1") != 300000 {
		t.Fatalf("expected buyer balance 800000, got %d", ld2.Balance("buyer1"))
	}

	// Release escrow
	releaseHandler := ReleaseEscrowHandler(dir)
	w2 := httptest.NewRecorder()
	releaseHandler(w2, localReq("POST", "/api/escrow/release?id="+escID, ""))

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on release, got %d: %s", w2.Code, w2.Body.String())
	}

	// Verify seller got funds
	ld3 := loadLedger(dir)
	if ld3.Balance("seller1") != 200000 {
		t.Fatalf("expected seller balance 200000, got %d", ld3.Balance("seller1"))
	}

	// Verify escrow status changed
	listHandler := ListEscrowHandler(dir)
	w3 := httptest.NewRecorder()
	listHandler(w3, localReq("GET", "/api/escrow/list?status=released", ""))

	var listResp map[string]any
	json.Unmarshal(w3.Body.Bytes(), &listResp)
	if listResp["count"] != float64(1) {
		t.Fatalf("expected 1 released escrow, got %v", listResp["count"])
	}
}


// TestCreateEscrowInsufficientBalance handles the TestCreateEscrowInsufficientBalance HTTP request.
func TestCreateEscrowInsufficientBalance(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	ld := loadLedger(dir)
	ld.Mint("poor-buyer", 100)
	ld.Save(dir)

	handler := CreateEscrowHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/api/escrow/create", `{"buyer":"poor-buyer","seller":"seller","item_id":"X","price_ng":1000}`))

	if w.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402 for insufficient balance, got %d: %s", w.Code, w.Body.String())
	}
}


// TestCreateEscrowMissingFields handles the TestCreateEscrowMissingFields HTTP request.
func TestCreateEscrowMissingFields(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	handler := CreateEscrowHandler(dir)

	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/api/escrow/create", `{}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}


// TestCancelEscrow handles the TestCancelEscrow HTTP request.
func TestCancelEscrow(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	ld := loadLedger(dir)
	ld.Mint("buyer2", 500000)
	ld.Save(dir)

	createHandler := CreateEscrowHandler(dir)
	w := httptest.NewRecorder()
	createHandler(w, localReq("POST", "/api/escrow/create", `{"buyer":"buyer2","seller":"seller2","item_id":"CANCEL_TEST","price_ng":100000}`))

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	escID := resp["escrow"].(map[string]any)["id"].(string)

	// Cancel
	cancelHandler := CancelEscrowHandler(dir)
	w2 := httptest.NewRecorder()
	cancelHandler(w2, localReq("POST", "/api/escrow/cancel?id="+escID, ""))

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on cancel, got %d: %s", w2.Code, w2.Body.String())
	}

	// Verify funds returned
	ld2 := loadLedger(dir)
	if ld2.Balance("buyer2") != 500000 {
		t.Fatalf("expected buyer balance restored to 500000, got %d", ld2.Balance("buyer2"))
	}

	// Verify escrow status
	listHandler := ListEscrowHandler(dir)
	w3 := httptest.NewRecorder()
	listHandler(w3, localReq("GET", "/api/escrow/list?status=cancelled", ""))

	var listResp map[string]any
	json.Unmarshal(w3.Body.Bytes(), &listResp)
	if listResp["count"] != float64(1) {
		t.Fatalf("expected 1 cancelled escrow, got %v", listResp["count"])
	}
}


// TestReleaseAlreadyCancelledEscrow handles the TestReleaseAlreadyCancelledEscrow HTTP request.
func TestReleaseAlreadyCancelledEscrow(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	ld := loadLedger(dir)
	ld.Mint("buyer3", 500000)
	ld.Save(dir)

	createHandler := CreateEscrowHandler(dir)
	w := httptest.NewRecorder()
	createHandler(w, localReq("POST", "/api/escrow/create", `{"buyer":"buyer3","seller":"seller3","item_id":"DUP_TEST","price_ng":100000}`))

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	escID := resp["escrow"].(map[string]any)["id"].(string)

	// Cancel first
	cancelHandler := CancelEscrowHandler(dir)
	w2 := httptest.NewRecorder()
	cancelHandler(w2, localReq("POST", "/api/escrow/cancel?id="+escID, ""))

	// Try to release after cancel
	releaseHandler := ReleaseEscrowHandler(dir)
	w3 := httptest.NewRecorder()
	releaseHandler(w3, localReq("POST", "/api/escrow/release?id="+escID, ""))

	if w3.Code != http.StatusConflict {
		t.Fatalf("expected 409 on releasing cancelled escrow, got %d", w3.Code)
	}
}


// TestListEscrowNoFilter handles the TestListEscrowNoFilter HTTP request.
func TestListEscrowNoFilter(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	ld := loadLedger(dir)
	ld.Mint("buyer-list", 1000000)
	ld.Save(dir)

	createHandler := CreateEscrowHandler(dir)
	createHandler(httptest.NewRecorder(), localReq("POST", "/api/escrow/create", `{"buyer":"buyer-list","seller":"slr","item_id":"L1","price_ng":100000}`))
	createHandler(httptest.NewRecorder(), localReq("POST", "/api/escrow/create", `{"buyer":"buyer-list","seller":"slr","item_id":"L2","price_ng":200000}`))

	listHandler := ListEscrowHandler(dir)
	w := httptest.NewRecorder()
	listHandler(w, localReq("GET", "/api/escrow/list", ""))

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"] != float64(2) {
		t.Fatalf("expected 2 escrows, got %v", resp["count"])
	}
}


// TestEscrowBuyWithRelease handles the TestEscrowBuyWithRelease HTTP request.
func TestEscrowBuyWithRelease(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	ld := loadLedger(dir)
	ld.Mint("esc-buyer", 1000000)
	ld.Save(dir)

	// Create an RWA item for sale
	rwaItems := []map[string]any{
		{"id": "RWA001", "holder": "esc-seller", "name": "Test Asset", "for_sale": true, "price_ng": float64(300000)},
	}
	b, _ := json.Marshal(rwaItems)
	os.WriteFile(filepath.Join(dir, "rwa_registry.json"), b, 0600)

	// Escrow buy
	buyHandler := NewEscrowBuyHandler(dir)
	w := httptest.NewRecorder()
	buyHandler(w, localReq("POST", "/api/escrow/buy", `{"buyer":"esc-buyer","seller":"esc-seller","item_id":"RWA001"}`))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ok"] != true {
		t.Fatal("expected ok=true")
	}

	esc := resp["escrow"].(map[string]any)
	if esc["status"] != "active" {
		t.Fatalf("expected active escrow, got %v", esc["status"])
	}
	escID := esc["id"].(string)

	// Release
	releaseHandler := ReleaseEscrowHandler(dir)
	w2 := httptest.NewRecorder()
	releaseHandler(w2, localReq("POST", "/api/escrow/release?id="+escID, ""))

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on release, got %d: %s", w2.Code, w2.Body.String())
	}

	// Check seller got paid
	ld2 := loadLedger(dir)
	if ld2.Balance("esc-seller") != 300000 {
		t.Fatalf("expected seller 300000, got %d", ld2.Balance("esc-seller"))
	}
	// Check buyer balance deducted
	if ld2.Balance("esc-buyer") != 700000 {
		t.Fatalf("expected buyer 700000, got %d", ld2.Balance("esc-buyer"))
	}
}


// TestMarketSellHandler handles the TestMarketSellHandler HTTP request.
func TestMarketSellHandler(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	billRecorder := func(price int64, action, itemID string) {}

	handler := MarketSellHandler(dir, billRecorder)
	w := httptest.NewRecorder()
	handler(w, localReq("POST", "/api/market/sell", `{"id":"MKT001","price_ng":50000}`))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}


// TestMarketListHandler handles the TestMarketListHandler HTTP request.
func TestMarketListHandler(t *testing.T) {
	dir := newTestTreasuryDataDir(t)
	// Add a listing
	billRecorder := func(price int64, action, itemID string) {}
	MarketSellHandler(dir, billRecorder)(httptest.NewRecorder(), localReq("POST", "/api/market/sell", `{"id":"MKT002","price_ng":99999}`))

	handler := MarketListHandler(dir)
	w := httptest.NewRecorder()
	handler(w, localReq("GET", "/api/market/list", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"] != float64(1) {
		t.Fatalf("expected 1 listing, got %v", resp["count"])
	}
}
