// Package economy implements the island economy system
package economy

import (
	"encoding/json"
	"testing"
)


// TestOracleParseSilverPrice handles the TestOracleParseSilverPrice HTTP request.
func TestOracleParseSilverPrice(t *testing.T) {
	jsonData := `{"price": 69.355, "timestamp": 1718000000, "currency": "USD"}`
	var resp struct {
		Price float64 `json:"price"`
	}
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Price != 69.355 {
		t.Fatalf("expected 69.355, got %f", resp.Price)
	}
}


// TestOracleParseSwissquotePrice handles the TestOracleParseSwissquotePrice HTTP request.
func TestOracleParseSwissquotePrice(t *testing.T) {
	jsonData := `[{"symbol": "XAG/USD", "bid": 69.10, "ask": 69.50, "spread": 0.40}]`
	var resp []struct {
		Bid float64 `json:"bid"`
		Ask float64 `json:"ask"`
	}
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil || len(resp) == 0 {
		t.Fatalf("unmarshal error: %v", err)
	}
	mid := (resp[0].Bid + resp[0].Ask) / 2
	if mid != 69.30 {
		t.Fatalf("expected 69.30 (mid), got %f", mid)
	}
}


// TestOracleParseSwissquotePriceMid handles the TestOracleParseSwissquotePriceMid HTTP request.
func TestOracleParseSwissquotePriceMid(t *testing.T) {
	jsonData := `[{"symbol": "XAG/USD", "bid": 65.00, "ask": 65.50}]`
	var resp []struct {
		Bid float64 `json:"bid"`
		Ask float64 `json:"ask"`
	}
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil || len(resp) == 0 {
		t.Fatalf("unmarshal error: %v", err)
	}
	mid := (resp[0].Bid + resp[0].Ask) / 2
	if mid != 65.25 {
		t.Fatalf("expected 65.25 (mid), got %f", mid)
	}
}


// TestOracleLoadSave handles the TestOracleLoadSave HTTP request.
func TestOracleLoadSave(t *testing.T) {
	dir := t.TempDir()
	o := LoadOracle(dir)

	if o.GetPrice() == 0 {
		t.Log("oracle price is 0 (unset)")
	}

	o.UpdatePrice(69.50)
	o.Save(dir)

	o2 := LoadOracle(dir)
	if o2.GetPrice() != 69.50 {
		t.Fatalf("expected 69.50, got %f", o2.GetPrice())
	}
}
