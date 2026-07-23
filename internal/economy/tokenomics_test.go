// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestUtilityPremiumPct handles the TestUtilityPremiumPct HTTP request.
func TestUtilityPremiumPct(t *testing.T) {
	if UtilityPremiumPct != 0.30 {
		t.Fatalf("expected 0.30, got %f", UtilityPremiumPct)
	}
}


// TestSilverPortionNg handles the TestSilverPortionNg HTTP request.
func TestSilverPortionNg(t *testing.T) {
	if SilverPortionNg != 70*NGPerTLR/100 {
		t.Fatalf("expected %d, got %d", 70*NGPerTLR/100, SilverPortionNg)
	}
}


// TestUtilityPremiumNg handles the TestUtilityPremiumNg HTTP request.
func TestUtilityPremiumNg(t *testing.T) {
	if UtilityPremiumNg != 30*NGPerTLR/100 {
		t.Fatalf("expected %d, got %d", 30*NGPerTLR/100, UtilityPremiumNg)
	}
}


// TestSilverPlusPremiumEqualsNGPerTLR handles the TestSilverPlusPremiumEqualsNGPerTLR HTTP request.
func TestSilverPlusPremiumEqualsNGPerTLR(t *testing.T) {
	// SilverPortionNg + UtilityPremiumNg should equal NGPerTLR
	sum := SilverPortionNg + UtilityPremiumNg
	if sum != NGPerTLR {
		t.Fatalf("%d + %d = %d, expected %d", SilverPortionNg, UtilityPremiumNg, sum, NGPerTLR)
	}
}


// TestPremiumPerIssue handles the TestPremiumPerIssue HTTP request.
func TestPremiumPerIssue(t *testing.T) {
	expected := TLRPerIssue * UtilityPremiumNg
	if PremiumPerIssue != expected {
		t.Fatalf("expected %d, got %d", expected, PremiumPerIssue)
	}
}


// TestAllocatePremium handles the TestAllocatePremium HTTP request.
func TestAllocatePremium(t *testing.T) {
	p := AllocatePremium()
	if p.TotalNg != PremiumPerIssue {
		t.Fatalf("premium total: expected %d, got %d", PremiumPerIssue, p.TotalNg)
	}
	// All components should be positive integers
	if p.TreasuryNg <= 0 || p.DividendNg <= 0 || p.SilverBuyNg <= 0 || p.AuctionNg <= 0 || p.BuybackNg <= 0 {
		t.Fatal("all premium components should be positive")
	}
	// Sum should equal total
	sum := p.TreasuryNg + p.DividendNg + p.SilverBuyNg + p.AuctionNg + p.BuybackNg
	if sum != p.TotalNg {
		t.Fatalf("premium sum %d != total %d", sum, p.TotalNg)
	}
}


// TestPremiumConstantsAreIntegers handles the TestPremiumConstantsAreIntegers HTTP request.
func TestPremiumConstantsAreIntegers(t *testing.T) {
	// All premium constants should be multiples of NGPerTLR/10
	base := NGPerTLR / 10
	if TreasuryPremiumNg%base != 0 {
		t.Fatalf("TreasuryPremiumNg %d not multiple of %d", TreasuryPremiumNg, base)
	}
	if DividendPremiumNg%base != 0 {
		t.Fatalf("DividendPremiumNg %d not multiple of %d", DividendPremiumNg, base)
	}
	if SilverBuyPremiumNg%base != 0 {
		t.Fatalf("SilverBuyPremiumNg %d not multiple of %d", SilverBuyPremiumNg, base)
	}
	if AuctionPremiumNg%base != 0 {
		t.Fatalf("AuctionPremiumNg %d not multiple of %d", AuctionPremiumNg, base)
	}
	if BuybackPremiumNg%base != 0 {
		t.Fatalf("BuybackPremiumNg %d not multiple of %d", BuybackPremiumNg, base)
	}
}


// TestPremiumRatios handles the TestPremiumRatios HTTP request.
func TestPremiumRatios(t *testing.T) {
	// 5% : 50% : 10% : 15% : 20% = 100%
	total := float64(PremiumPerIssue)
	tol := 0.01
	ratios := []struct {
		name  string
		value int64
		pct   float64
	}{
		{"treasury", TreasuryPremiumNg, 8.0},
		{"dividend", DividendPremiumNg, 47.0},
		{"silver_buy", SilverBuyPremiumNg, 10.0},
		{"auction", AuctionPremiumNg, 15.0},
		{"buyback", BuybackPremiumNg, 20.0},
	}
	for _, r := range ratios {
		got := float64(r.value) / total * 100
		if got < r.pct-tol || got > r.pct+tol {
			t.Errorf("%s: expected ~%.1f%%, got %.2f%%", r.name, r.pct, got)
		}
	}
}


// TestCalculateFullIssuance handles the TestCalculateFullIssuance HTTP request.
func TestCalculateFullIssuance(t *testing.T) {
	f := CalculateFullIssuance()
	if f.InvestorNg != SilverPortionNg*TLRPerIssue {
		t.Fatalf("investor: expected %d, got %d", SilverPortionNg*TLRPerIssue, f.InvestorNg)
	}
	// All premium pools
	if f.TreasuryNg != TreasuryPremiumNg {
		t.Fatalf("treasury: expected %d, got %d", TreasuryPremiumNg, f.TreasuryNg)
	}
	if f.DividendPoolNg != DividendPremiumNg {
		t.Fatalf("dividend: expected %d, got %d", DividendPremiumNg, f.DividendPoolNg)
	}
	// Total should be all-in
	expectedTotal := f.InvestorNg + f.TreasuryNg + f.DividendPoolNg + f.SilverBuyPoolNg + f.AuctionPoolNg + f.BuybackPoolNg
	if f.TotalNg != expectedTotal {
		t.Fatalf("total: expected %d, got %d", expectedTotal, f.TotalNg)
	}
}

// === Rarity Weight Tests ===

func TestRarityWeight(t *testing.T) {
	cases := []struct {
		rarity string
		want   int
	}{
		{"common", 1},
		{"rare", 2},
		{"epic", 5},
		{"legendary", 10},
		{"genesis", 20},
		{"unknown", 1},
	}
	for _, c := range cases {
		got := RarityWeight(c.rarity)
		if got != c.want {
			t.Errorf("RarityWeight(%q) = %d, want %d", c.rarity, got, c.want)
		}
	}
}


// TestDividendWeight handles the TestDividendWeight HTTP request.
func TestDividendWeight(t *testing.T) {
	b := BanknoteV2{DenominationNg: NGPerTLR, Rarity: "rare"}
	w := DividendWeight(b)
	if w != NGPerTLR*2 {
		t.Fatalf("rare banknote: expected weight %d, got %d", NGPerTLR*2, w)
	}
}


// TestHolderDividendShare handles the TestHolderDividendShare HTTP request.
func TestHolderDividendShare(t *testing.T) {
	banknotes := []BanknoteV2{
		{Serial: "A", DenominationNg: NGPerTLR, Rarity: "common", Holder: "alice", Status: "active"},
		{Serial: "B", DenominationNg: NGPerTLR, Rarity: "rare", Holder: "bob", Status: "active"},
		{Serial: "C", DenominationNg: NGPerTLR, Rarity: "common", Holder: "alice", Status: "active"},
	}
	// alice: 1*NGPerTLR + 1*NGPerTLR = 2*NGPerTLR
	// bob: 2*NGPerTLR
	// total: 4*NGPerTLR
	// alice share: 2/4 = 0.5
	// bob share: 2/4 = 0.5
	alice := HolderDividendShare("alice", banknotes)
	if alice < 0.49 || alice > 0.51 {
		t.Fatalf("alice share: expected ~0.50, got %f", alice)
	}
	shares := AllHoldersDividendShares(banknotes)
	if len(shares) != 2 {
		t.Fatalf("expected 2 holders, got %d", len(shares))
	}
}


// TestAllHoldersDividendSharesSkipsBurned handles the TestAllHoldersDividendSharesSkipsBurned HTTP request.
func TestAllHoldersDividendSharesSkipsBurned(t *testing.T) {
	banknotes := []BanknoteV2{
		{Serial: "A", DenominationNg: NGPerTLR, Rarity: "common", Holder: "alice", Status: "active"},
		{Serial: "B", DenominationNg: NGPerTLR, Rarity: "common", Holder: "bob", Status: "burned"},
	}
	shares := AllHoldersDividendShares(banknotes)
	if _, ok := shares["bob"]; ok {
		t.Fatal("bob should not get dividends (burned banknote)")
	}
}


// TestDividendRoundDistribution handles the TestDividendRoundDistribution HTTP request.
func TestDividendRoundDistribution(t *testing.T) {
	dir := t.TempDir()
	// Create banknotes
	banknotes := []BanknoteV2{
		{Serial: "A", DenominationNg: NGPerTLR, Rarity: "epic", Holder: "alice", Status: "active", FrozenNg: NGPerTLR * 5},
		{Serial: "B", DenominationNg: NGPerTLR, Rarity: "common", Holder: "bob", Status: "active", FrozenNg: NGPerTLR},
	}
	SaveBanknotesV2(dir, banknotes)

	// Seed ledger
	ledger := LoadLedger(dir)
	ledger.Mint("treasury", 100*NGPerTLR)
	ledger.Save(dir)

	dd := NewDividendDistributor()
	// alice: 1*5 = 5 total weight, bob: 1*1 = 1 total weight. Total = 6
	// dividend = 6*NGPerTLR
	round, err := dd.Distribute(dir, 60*NGPerTLR, "silver-round-1")
	if err != nil {
		t.Fatal(err)
	}
	if round.TotalNg != 60*NGPerTLR {
		t.Fatalf("total distributed: expected %d, got %d", 60*NGPerTLR, round.TotalNg)
	}
	if len(round.Payments) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(round.Payments))
	}
	// alice should get ~50*NGPerTLR (5/6 of 60*NGPerTLR), bob ~10*NGPerTLR (1/6)
	// Check alice got more than bob (order-independent)
	aliceAmt, bobAmt := int64(0), int64(0)
	for _, p := range round.Payments {
		switch p.Holder {
		case "alice":
			aliceAmt += p.DividendNg
		case "bob":
			bobAmt += p.DividendNg
		}
	}
	if aliceAmt <= bobAmt {
		t.Fatalf("alice (%d) should get more dividends than bob (%d)", aliceAmt, bobAmt)
	}
}

// === Auction Fee Tests ===

func TestAuctionListingFee(t *testing.T) {
	startPrice := int64(1000 * NGPerTLR)
	fee := startPrice * AuctionListingFeeBPS / 10000
	if fee < AuctionMinListingFeeNg {
		fee = AuctionMinListingFeeNg
	}
	expected := startPrice * 50 / 10000 // 0.5%
	if fee != expected {
		t.Fatalf("listing fee: expected %d, got %d", expected, fee)
	}
}


// TestAuctionMinListingFee handles the TestAuctionMinListingFee HTTP request.
func TestAuctionMinListingFee(t *testing.T) {
	// Use a tiny start price so 0.5% is below the 1M min
	startPrice := int64(1_000_000)
	fee := startPrice * AuctionListingFeeBPS / 10000
	if fee < AuctionMinListingFeeNg {
		fee = AuctionMinListingFeeNg
	}
	if fee != AuctionMinListingFeeNg {
		t.Fatalf("min listing fee: expected %d, got %d", AuctionMinListingFeeNg, fee)
	}
}


// TestAuctionListingFeeAboveMin handles the TestAuctionListingFeeAboveMin HTTP request.
func TestAuctionListingFeeAboveMin(t *testing.T) {
	startPrice := int64(1000 * NGPerTLR)
	fee := startPrice * AuctionListingFeeBPS / 10000
	if fee < AuctionMinListingFeeNg {
		fee = AuctionMinListingFeeNg
	}
	expected := startPrice * AuctionListingFeeBPS / 10000
	if fee != expected {
		t.Fatalf("listing fee: expected %d, got %d", expected, fee)
	}
}


// TestAuctionBuyerPremium handles the TestAuctionBuyerPremium HTTP request.
func TestAuctionBuyerPremium(t *testing.T) {
	finalPrice := int64(1000 * NGPerTLR)
	premium := finalPrice * AuctionBuyerPremiumBPS / 10000
	if premium != finalPrice*250/10000 {
		t.Fatalf("buyer premium: expected %d, got %d", finalPrice*250/10000, premium)
	}
}


// TestAuctionSellerFee handles the TestAuctionSellerFee HTTP request.
func TestAuctionSellerFee(t *testing.T) {
	finalPrice := int64(1000 * NGPerTLR)
	fee := finalPrice * AuctionSellerFeeBPS / 10000
	sellerProceeds := finalPrice - fee
	if sellerProceeds != finalPrice*99/100 {
		t.Fatalf("seller proceeds: expected %d, got %d", finalPrice*99/100, sellerProceeds)
	}
}


// TestAuctionFeeSum handles the TestAuctionFeeSum HTTP request.
func TestAuctionFeeSum(t *testing.T) {
	// On a 1000-TLR sale:
	// Listing: 0.5% = 5 TLR
	// Seller: 1% = 10 TLR
	// Buyer premium: 2.5% = 25 TLR
	// Total fees to treasury = 5 + 10 + 25 = 40 TLR = 4%
	price := int64(1000 * NGPerTLR)
	listingFee := price * AuctionListingFeeBPS / 10000
	if listingFee < AuctionMinListingFeeNg {
		listingFee = AuctionMinListingFeeNg
	}
	sellerFee := price * AuctionSellerFeeBPS / 10000
	buyerPremium := price * AuctionBuyerPremiumBPS / 10000
	totalFees := listingFee + sellerFee + buyerPremium
	expectedFees := price * 400 / 10000 // 4%
	if totalFees != expectedFees {
		t.Fatalf("total fees: expected %d, got %d", expectedFees, totalFees)
	}
}
