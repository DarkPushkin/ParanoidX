// Package economy implements the island economy system
package economy

import (
	"math"
	"testing"
)


// TestNewOracleHasDefaultPrice handles the TestNewOracleHasDefaultPrice HTTP request.
func TestNewOracleHasDefaultPrice(t *testing.T) {
	o := NewSilverSpotOracle()
	if o.CurrentPrice != DefaultSilverSpotUSDperOZ {
		t.Fatalf("expected %f, got %f", DefaultSilverSpotUSDperOZ, o.CurrentPrice)
	}
}


// TestGetPrice handles the TestGetPrice HTTP request.
func TestGetPrice(t *testing.T) {
	o := NewSilverSpotOracle()
	if o.GetPrice() != DefaultSilverSpotUSDperOZ {
		t.Fatal("get price mismatch")
	}
}


// TestUpdatePriceValid handles the TestUpdatePriceValid HTTP request.
func TestUpdatePriceValid(t *testing.T) {
	o := NewSilverSpotOracle()
	err := o.UpdatePrice(80.0)
	if err != nil {
		t.Fatal(err)
	}
	if o.CurrentPrice != 80.0 {
		t.Fatalf("expected 80, got %f", o.CurrentPrice)
	}
}


// TestUpdatePriceInvalid handles the TestUpdatePriceInvalid HTTP request.
func TestUpdatePriceInvalid(t *testing.T) {
	o := NewSilverSpotOracle()
	if err := o.UpdatePrice(-1); err == nil {
		t.Fatal("expected error for negative price")
	}
	if err := o.UpdatePrice(0); err == nil {
		t.Fatal("expected error for zero price")
	}
}


// TestUpdatePriceNaN handles the TestUpdatePriceNaN HTTP request.
func TestUpdatePriceNaN(t *testing.T) {
	o := NewSilverSpotOracle()
	nan := math.NaN()
	if err := o.UpdatePrice(nan); err == nil {
		t.Fatal("expected error for NaN")
	}
}


// TestUpdatePriceInf handles the TestUpdatePriceInf HTTP request.
func TestUpdatePriceInf(t *testing.T) {
	o := NewSilverSpotOracle()
	if err := o.UpdatePrice(math.Inf(1)); err == nil {
		t.Fatal("expected error for Inf")
	}
}


// TestUSDTtoNGWithOracle handles the TestUSDTtoNGWithOracle HTTP request.
func TestUSDTtoNGWithOracle(t *testing.T) {
	o := NewSilverSpotOracle()
	ng := o.USDTtoNG(1.0)
	ratio := float64(NGPerTLR) / DefaultSilverSpotUSDperOZ
	expected := int64(ratio)
	if ng != expected {
		t.Fatalf("got %d, want %d", ng, expected)
	}
}


// TestUSDTtoNGWithCustomPrice handles the TestUSDTtoNGWithCustomPrice HTTP request.
func TestUSDTtoNGWithCustomPrice(t *testing.T) {
	o := NewSilverSpotOracle()
	o.UpdatePrice(100.0)
	ng := o.USDTtoNG(1.0)
	ratio := float64(NGPerTLR) / 100.0
	expected := int64(ratio)
	if ng != expected {
		t.Fatalf("got %d, want %d", ng, expected)
	}
}


// TestSaveAndLoadOracle handles the TestSaveAndLoadOracle HTTP request.
func TestSaveAndLoadOracle(t *testing.T) {
	dir := t.TempDir()
	o := NewSilverSpotOracle()
	o.UpdatePrice(82.5)
	o.Save(dir)

	loaded := LoadOracle(dir)
	if loaded.CurrentPrice != 82.5 {
		t.Fatalf("expected 82.5, got %f", loaded.CurrentPrice)
	}
}


// TestDeflationManagerNew handles the TestDeflationManagerNew HTTP request.
func TestDeflationManagerNew(t *testing.T) {
	d := NewDeflationManager()
	if d.ThresholdBasis != 12 {
		t.Fatalf("expected threshold 12, got %d", d.ThresholdBasis)
	}
}


// TestDeflationNoBurnBelowThreshold handles the TestDeflationNoBurnBelowThreshold HTTP request.
func TestDeflationNoBurnBelowThreshold(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	d := NewDeflationManager()
	amt, err := d.Burn("", cfg.MonthlyOpsNg*10, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if amt != 0 {
		t.Fatalf("expected 0 burn below threshold, got %d", amt)
	}
}


// TestDeflationBurnAtVeryFat handles the TestDeflationBurnAtVeryFat HTTP request.
func TestDeflationBurnAtVeryFat(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultTreasuryConfig()
	d := NewDeflationManager()

	// Create a ledger with initial supply
	ledger := LoadLedger(dir)
	ledger.TotalSupply = cfg.MonthlyOpsNg * 100
	ledger.Save(dir)

	amt, err := d.Burn(dir, cfg.MonthlyOpsNg*12, cfg)
	if err != nil {
		t.Fatal(err)
	}
	expected := cfg.MonthlyOpsNg * 12 * 40 / 100
	if amt != expected {
		t.Fatalf("expected %d burn, got %d", expected, amt)
	}

	// Check total supply reduced
	ledger2 := LoadLedger(dir)
	if ledger2.TotalSupply >= ledger.TotalSupply {
		t.Fatal("total supply should have decreased")
	}
}


// TestDeflationTracksTotal handles the TestDeflationTracksTotal HTTP request.
func TestDeflationTracksTotal(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultTreasuryConfig()
	d := NewDeflationManager()

	d.Burn(dir, cfg.MonthlyOpsNg*12, cfg)
	d.Burn(dir, cfg.MonthlyOpsNg*15, cfg)

	if d.TotalBurnedNg <= 0 {
		t.Fatal("total burned should be > 0")
	}
}


// TestDeflationSaveAndLoad handles the TestDeflationSaveAndLoad HTTP request.
func TestDeflationSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	d := NewDeflationManager()
	d.TotalBurnedNg = 5000000
	d.Save(dir)

	loaded := LoadDeflation(dir)
	if loaded.TotalBurnedNg != 5000000 {
		t.Fatalf("expected 5000000, got %d", loaded.TotalBurnedNg)
	}
}


// TestOracleHistoryLen handles the TestOracleHistoryLen HTTP request.
func TestOracleHistoryLen(t *testing.T) {
	o := NewSilverSpotOracle()
	for i := 0; i < 10; i++ {
		o.UpdatePrice(float64(75 + i))
	}
	if len(o.History) != 10 {
		t.Fatalf("expected 10 history entries, got %d", len(o.History))
	}
}


// TestUpdatePriceRecordsHistory handles the TestUpdatePriceRecordsHistory HTTP request.
func TestUpdatePriceRecordsHistory(t *testing.T) {
	o := NewSilverSpotOracle()
	o.UpdatePrice(80.0)
	if len(o.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(o.History))
	}
	if o.History[0].Price != DefaultSilverSpotUSDperOZ {
		t.Fatalf("history should store previous price, got %f", o.History[0].Price)
	}
}
