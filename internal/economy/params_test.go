// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestDefaultDynamicParams handles the TestDefaultDynamicParams HTTP request.
func TestDefaultDynamicParams(t *testing.T) {
	p := DefaultDynamicParams()
	if p.TreasuryCommissionBPS != 228 {
		t.Errorf("expected 228, got %d", p.TreasuryCommissionBPS)
	}
	if p.MaxTotalFeeBPS != 420 {
		t.Errorf("expected 420, got %d", p.MaxTotalFeeBPS)
	}
	if p.SilverBackingRatio != 0.70 {
		t.Errorf("expected 0.70, got %f", p.SilverBackingRatio)
	}
}


// TestAdjustTreasuryCommission handles the TestAdjustTreasuryCommission HTTP request.
func TestAdjustTreasuryCommission(t *testing.T) {
	p := DefaultDynamicParams()
	if !p.Adjust("treasury_commission_bps", 200) {
		t.Error("expected adjust to succeed")
	}
	if p.TreasuryCommissionBPS != 200 {
		t.Errorf("expected 200, got %d", p.TreasuryCommissionBPS)
	}
	// Out of range
	if p.Adjust("treasury_commission_bps", 50) {
		t.Error("expected adjust to fail for out-of-range value")
	}
}


// TestAdjustSilverBacking handles the TestAdjustSilverBacking HTTP request.
func TestAdjustSilverBacking(t *testing.T) {
	p := DefaultDynamicParams()
	if !p.Adjust("silver_backing_ratio", 0.65) {
		t.Error("expected adjust to succeed")
	}
	if p.Adjust("silver_backing_ratio", 0.10) {
		t.Error("expected adjust to fail for out-of-range value")
	}
}


// TestAdjustAuctionFee handles the TestAdjustAuctionFee HTTP request.
func TestAdjustAuctionFee(t *testing.T) {
	p := DefaultDynamicParams()
	if !p.Adjust("auction_listing_fee_bps", 80) {
		t.Error("expected adjust to succeed")
	}
	if p.AuctionListingFeeBPS != 80 {
		t.Errorf("expected 80, got %d", p.AuctionListingFeeBPS)
	}
}


// TestAdjustInvalidName handles the TestAdjustInvalidName HTTP request.
func TestAdjustInvalidName(t *testing.T) {
	p := DefaultDynamicParams()
	if p.Adjust("nonexistent", 42) {
		t.Error("expected adjust to fail for unknown param")
	}
}


// TestDynamicParamsAll handles the TestDynamicParamsAll HTTP request.
func TestDynamicParamsAll(t *testing.T) {
	p := DefaultDynamicParams()
	all := p.All()
	if len(all) == 0 {
		t.Fatal("expected non-empty params map")
	}
	if all["treasury_commission_bps"] != 228 {
		t.Errorf("expected 228, got %f", all["treasury_commission_bps"])
	}
}


// TestDynamicParamsGet handles the TestDynamicParamsGet HTTP request.
func TestDynamicParamsGet(t *testing.T) {
	p := DefaultDynamicParams()
	if p.Get("treasury_commission_bps") != 228 {
		t.Errorf("expected 228, got %f", p.Get("treasury_commission_bps"))
	}
	if p.Get("nonexistent") != 0 {
		t.Errorf("expected 0 for unknown param")
	}
}


// TestDynamicParamsPersistence handles the TestDynamicParamsPersistence HTTP request.
func TestDynamicParamsPersistence(t *testing.T) {
	dir := t.TempDir()

	p1 := DefaultDynamicParams()
	p1.TreasuryCommissionBPS = 250
	p1.Save(dir)

	p2 := LoadDynamicParams(dir)
	if p2.TreasuryCommissionBPS != 250 {
		t.Errorf("expected 250, got %d", p2.TreasuryCommissionBPS)
	}
}
