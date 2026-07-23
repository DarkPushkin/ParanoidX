// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestSilverBackingRatio handles the TestSilverBackingRatio HTTP request.
func TestSilverBackingRatio(t *testing.T) {
	if SilverBackingRatio != 0.70 {
		t.Fatalf("expected 0.70, got %f", SilverBackingRatio)
	}
}


// TestReserveOzPerIssue handles the TestReserveOzPerIssue HTTP request.
func TestReserveOzPerIssue(t *testing.T) {
	if ReserveOzPerIssue != 70 {
		t.Fatalf("expected 70, got %d", ReserveOzPerIssue)
	}
}


// TestTLRPerIssue handles the TestTLRPerIssue HTTP request.
func TestTLRPerIssue(t *testing.T) {
	if TLRPerIssue != 100 {
		t.Fatalf("expected 100, got %d", TLRPerIssue)
	}
}


// TestNgPerIssue handles the TestNgPerIssue HTTP request.
func TestNgPerIssue(t *testing.T) {
	expected := int64(100) * NGPerTLR
	if NgPerIssue != expected {
		t.Fatalf("expected %d, got %d", expected, NgPerIssue)
	}
}


// TestCalculateIssuanceSplit handles the TestCalculateIssuanceSplit HTTP request.
func TestCalculateIssuanceSplit(t *testing.T) {
	split := CalculateIssuanceSplit()
	if split.TotalNg != NgPerIssue {
		t.Fatalf("total should be %d, got %d", NgPerIssue, split.TotalNg)
	}
	// 70% of 100 TLR = 70 TLR
	if split.InvestorNg != 70*NGPerTLR {
		t.Fatalf("investor share: expected %d, got %d", 70*NGPerTLR, split.InvestorNg)
	}
}


// TestIssuanceSplitSumsToTotalOld handles the TestIssuanceSplitSumsToTotalOld HTTP request.
func TestIssuanceSplitSumsToTotalOld(t *testing.T) {
	// Old 3-way sum still equals TotalNg (backward compat)
	split := CalculateIssuanceSplit()
	oldSum := split.InvestorNg + split.TreasuryNg + split.DividendPoolNg + split.SilverBuyNg + split.AuctionPoolNg + split.BuybackNg
	diff := split.TotalNg - oldSum
	if diff < -2 || diff > 2 {
		t.Fatalf("6-way sum %d != total %d (diff %d)", oldSum, split.TotalNg, diff)
	}
}


// TestTreasurySharePct handles the TestTreasurySharePct HTTP request.
func TestTreasurySharePct(t *testing.T) {
	split := CalculateIssuanceSplit()
	expected := int64(float64(NgPerIssue) * TreasurySharePct / 100.0)
	diff := split.TreasuryNg - expected
	if diff < -1 || diff > 1 {
		t.Fatalf("treasury: got %d, expected ~%d", split.TreasuryNg, expected)
	}
}


// TestDividendSharePctNew handles the TestDividendSharePctNew HTTP request.
func TestDividendSharePctNew(t *testing.T) {
	split := CalculateIssuanceSplit()
	expected := int64(float64(NgPerIssue) * DividendSharePct / 100.0)
	diff := split.DividendPoolNg - expected
	if diff < -1 || diff > 1 {
		t.Fatalf("dividends: got %d, expected ~%d", split.DividendPoolNg, expected)
	}
}


// TestReserveSilverOzToTLR handles the TestReserveSilverOzToTLR HTTP request.
func TestReserveSilverOzToTLR(t *testing.T) {
	tlr := ReserveSilverOzToTLR(70.0)
	if tlr != 100 {
		t.Fatalf("70 oz → expected 100 TLR, got %d", tlr)
	}
	tlr = ReserveSilverOzToTLR(35.0)
	if tlr != 50 {
		t.Fatalf("35 oz → expected 50 TLR, got %d", tlr)
	}
}


// TestTLRToRequiredSilverOz handles the TestTLRToRequiredSilverOz HTTP request.
func TestTLRToRequiredSilverOz(t *testing.T) {
	oz := TLRToRequiredSilverOz(100)
	if oz != 70.0 {
		t.Fatalf("100 TLR → expected 70 oz, got %f", oz)
	}
}


// TestIssuanceUnits handles the TestIssuanceUnits HTTP request.
func TestIssuanceUnits(t *testing.T) {
	units := IssuanceUnits(140)
	if units != 2 {
		t.Fatalf("140 oz → expected 2 units, got %d", units)
	}
	units = IssuanceUnits(70)
	if units != 1 {
		t.Fatalf("70 oz → expected 1 unit, got %d", units)
	}
	units = IssuanceUnits(30)
	if units != 0 {
		t.Fatalf("30 oz → expected 0 units, got %d", units)
	}
}


// TestUnusedReserveOz handles the TestUnusedReserveOz HTTP request.
func TestUnusedReserveOz(t *testing.T) {
	rem := UnusedReserveOz(75)
	if rem != 5 {
		t.Fatalf("75 oz unused → expected 5, got %d", rem)
	}
	rem = UnusedReserveOz(70)
	if rem != 0 {
		t.Fatalf("70 oz unused → expected 0, got %d", rem)
	}
}


// TestSilverStandardSummary handles the TestSilverStandardSummary HTTP request.
func TestSilverStandardSummary(t *testing.T) {
	s := SilverStandardSummary(70*NGPerTLR, 100*NGPerTLR)
	if s == "" {
		t.Fatal("expected non-empty summary")
	}
}


// TestNominalPriceMultiplier handles the TestNominalPriceMultiplier HTTP request.
func TestNominalPriceMultiplier(t *testing.T) {
	m := NominalPriceMultiplier()
	if m != 1.0/0.70 {
		t.Fatalf("expected ~1.4286, got %f", m)
	}
}


// TestNewNominalOffer handles the TestNewNominalOffer HTTP request.
func TestNewNominalOffer(t *testing.T) {
	offer := NewNominalOffer("round-1", 100)
	if !offer.Active {
		t.Fatal("offer should be active")
	}
	if offer.TotalTLR != 100 {
		t.Fatalf("expected 100 TLR, got %d", offer.TotalTLR)
	}
	ngf := float64(NGPerTLR)
	expectedPrice := int64(ngf / SilverBackingRatio)
	if offer.PricePerTLR != expectedPrice {
		t.Fatalf("expected price %d, got %d", expectedPrice, offer.PricePerTLR)
	}
}


// TestNominalOfferBuy handles the TestNominalOfferBuy HTTP request.
func TestNominalOfferBuy(t *testing.T) {
	dir := t.TempDir()
	ledger := LoadLedger(dir)
	// Mint enough ng to the buyer
	ngf := float64(NGPerTLR)
	costPerTLR := int64(ngf / SilverBackingRatio)
	ledger.Mint("investor1", costPerTLR*50)
	ledger.Save(dir)

	offer := NewNominalOffer("round-1", 100)
	issued, err := offer.Buy("investor1", 50, ledger, dir)
	if err != nil {
		t.Fatal(err)
	}
	if issued != 50*NGPerTLR {
		t.Fatalf("expected %d issued, got %d", 50*NGPerTLR, issued)
	}
	if offer.RemainingTLR != 50 {
		t.Fatalf("expected 50 remaining, got %d", offer.RemainingTLR)
	}
}


// TestNominalOfferInsufficientBalance handles the TestNominalOfferInsufficientBalance HTTP request.
func TestNominalOfferInsufficientBalance(t *testing.T) {
	dir := t.TempDir()
	ledger := LoadLedger(dir)
	offer := NewNominalOffer("round-2", 100)
	_, err := offer.Buy("poor_investor", 10, ledger, dir)
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}
}


// TestNominalOfferExhausted handles the TestNominalOfferExhausted HTTP request.
func TestNominalOfferExhausted(t *testing.T) {
	dir := t.TempDir()
	ledger := LoadLedger(dir)
	ngf := float64(NGPerTLR)
	cost := int64(ngf / SilverBackingRatio)
	ledger.Mint("buyer", cost*200)
	ledger.Save(dir)

	offer := NewNominalOffer("round-3", 50)
	offer.Buy("buyer", 50, ledger, dir)
	_, err := offer.Buy("buyer", 1, ledger, dir)
	if err == nil {
		t.Fatal("expected error for exhausted offer")
	}
}


// TestInvestorSharePctValue handles the TestInvestorSharePctValue HTTP request.
func TestInvestorSharePctValue(t *testing.T) {
	if InvestorSharePct != 70.0 {
		t.Fatalf("expected 70.0, got %f", InvestorSharePct)
	}
}


// TestTreasurySharePctValue handles the TestTreasurySharePctValue HTTP request.
func TestTreasurySharePctValue(t *testing.T) {
	if TreasurySharePct != 4.2 {
		t.Fatalf("expected 4.2, got %f", TreasurySharePct)
	}
}


// TestDividendSharePctValue handles the TestDividendSharePctValue HTTP request.
func TestDividendSharePctValue(t *testing.T) {
	if DividendSharePct != 12.9 {
		t.Fatalf("expected 12.9, got %f", DividendSharePct)
	}
}


// TestSilverBuyPremiumPctValue handles the TestSilverBuyPremiumPctValue HTTP request.
func TestSilverBuyPremiumPctValue(t *testing.T) {
	if SilverBuyPremiumPct != 3.0 {
		t.Fatalf("expected 3.0, got %f", SilverBuyPremiumPct)
	}
}


// TestAuctionPoolPctValue handles the TestAuctionPoolPctValue HTTP request.
func TestAuctionPoolPctValue(t *testing.T) {
	if AuctionPoolPct != 6.6 {
		t.Fatalf("expected 6.6, got %f", AuctionPoolPct)
	}
}


// TestBuybackReservePctValue handles the TestBuybackReservePctValue HTTP request.
func TestBuybackReservePctValue(t *testing.T) {
	if BuybackReservePct != 3.3 {
		t.Fatalf("expected 3.3, got %f", BuybackReservePct)
	}
}


// TestSumOfSharesEquals100 handles the TestSumOfSharesEquals100 HTTP request.
func TestSumOfSharesEquals100(t *testing.T) {
	sum := InvestorSharePct + TreasurySharePct + DividendSharePct + SilverBuyPremiumPct + AuctionPoolPct + BuybackReservePct
	if sum < 99.9 || sum > 100.1 {
		t.Fatalf("shares sum to %f, expected 100", sum)
	}
}


// TestCalculateIssuanceSplitSixWay handles the TestCalculateIssuanceSplitSixWay HTTP request.
func TestCalculateIssuanceSplitSixWay(t *testing.T) {
	split := CalculateIssuanceSplit()
	if split.InvestorNg != 70*NGPerTLR {
		t.Fatalf("investor: expected %d, got %d", 70*NGPerTLR, split.InvestorNg)
	}
	if split.DividendPoolNg <= 0 {
		t.Fatal("dividend pool should be positive")
	}
	if split.SilverBuyNg <= 0 {
		t.Fatal("silver buy pool should be positive")
	}
	if split.AuctionPoolNg <= 0 {
		t.Fatal("auction pool should be positive")
	}
	if split.BuybackNg <= 0 {
		t.Fatal("buyback pool should be positive")
	}
	if split.TotalNg != NgPerIssue {
		t.Fatalf("total: expected %d, got %d", NgPerIssue, split.TotalNg)
	}
}


// TestIssuanceSplitSumToTotalSixWay handles the TestIssuanceSplitSumToTotalSixWay HTTP request.
func TestIssuanceSplitSumToTotalSixWay(t *testing.T) {
	split := CalculateIssuanceSplit()
	sum := split.InvestorNg + split.TreasuryNg + split.DividendPoolNg + split.SilverBuyNg + split.AuctionPoolNg + split.BuybackNg
	diff := split.TotalNg - sum
	if diff < -2 || diff > 2 {
		t.Fatalf("6-way split sum %d != total %d (diff %d)", sum, split.TotalNg, diff)
	}
}
