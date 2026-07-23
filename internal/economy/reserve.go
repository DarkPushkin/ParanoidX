// Package economy implements the island economy system
package economy

import (
	"fmt"
	"sync"
)

// Hybrid Digital Silver Standard (30/70):
//   - 70% физическое серебро в резерве
//   - 30% утилитарная ценность сети (удобство расчётов)
//   - Каждые 70 oz реального серебра → эмиссия 100 TLR (3,110,348,000,000 ng)
//   - 100 TLR = 70 oz серебра (21,772,436,000 ng × 100) + 30% премия (9,331,044,000 ng × 100)
//   - Премия распределяется: казна 4.2%, дивиденды 12.9%, премия серебра 3.0%, аукционы 6.6%, байбэк 3.3%

const (
	// SilverBackingRatio — 1 TLR backed by 0.7 oz silver (70% standard)
	SilverBackingRatio = 0.70

	// ReserveOzPerIssue — ounces of silver required for one full issuance unit
	ReserveOzPerIssue = 70

	// TLRPerIssue — Liquid Talers issued per 70 oz of silver
	TLRPerIssue = 100

	// NgPerIssue — total ng value of one issuance unit
	NgPerIssue = TLRPerIssue * NGPerTLR // 100 × 31,103,480,000 = 3,110,348,000,000

	// InvestorSharePct — silver portion returned to the investor (70%)
	InvestorSharePct = 70.0

	// TreasurySharePct — network operations (4.2% of total NgPerIssue)
	TreasurySharePct = 4.2

	// DividendSharePct — distributed to banknote holders by rarity/denomination weight
	DividendSharePct = 12.9

	// SilverBuyPremiumPct — premium to buy silver above spot (3.0%)
	SilverBuyPremiumPct = 3.0

	// AuctionPoolPct — minting jubilee series + auction operations (6.6%)
	AuctionPoolPct = 6.6

	// BuybackReservePct — market buyback program (3.3%)
	BuybackReservePct = 3.3
)

// IssuanceSplit breaks down a single issuance unit into 6 components.
type IssuanceSplit struct {
	InvestorNg     int64 `json:"investor_ng"`
	TreasuryNg     int64 `json:"treasury_ng"`
	DividendPoolNg int64 `json:"dividend_pool_ng"`
	SilverBuyNg    int64 `json:"silver_buy_ng"`
	AuctionPoolNg  int64 `json:"auction_pool_ng"`
	BuybackNg      int64 `json:"buyback_ng"`
	TotalNg        int64 `json:"total_ng"`
}

// CalculateIssuanceSplit returns the full 6-way split for one issuance unit (70 oz → 100 TLR).
// Все значения — целые числа ng (кратны 1 ng).
func CalculateIssuanceSplit() IssuanceSplit {
	totalNg := NgPerIssue
	return IssuanceSplit{
		InvestorNg:     int64(float64(totalNg) * InvestorSharePct / 100.0),
		TreasuryNg:     int64(float64(totalNg) * TreasurySharePct / 100.0),
		DividendPoolNg: int64(float64(totalNg) * DividendSharePct / 100.0),
		SilverBuyNg:    int64(float64(totalNg) * SilverBuyPremiumPct / 100.0),
		AuctionPoolNg:  int64(float64(totalNg) * AuctionPoolPct / 100.0),
		BuybackNg:      int64(float64(totalNg) * BuybackReservePct / 100.0),
		TotalNg:        totalNg,
	}
}

// ReserveSilverOzToTLR converts actual silver oz to issuable TLR under 70% backing.
func ReserveSilverOzToTLR(silverOz float64) int64 {
	return int64(silverOz / SilverBackingRatio)
}

// TLRToRequiredSilverOz returns how many oz of silver back the given TLR.
func TLRToRequiredSilverOz(tlr int64) float64 {
	return float64(tlr) * SilverBackingRatio
}

// IssuanceUnits returns how many full issuance units a given silver reserve supports.
func IssuanceUnits(reserveOz int64) int {
	return int(reserveOz / ReserveOzPerIssue)
}

// UnusedReserveOz returns the remainder silver oz not used in full issuance units.
func UnusedReserveOz(reserveOz int64) int64 {
	return reserveOz % ReserveOzPerIssue
}

// SilverStandardSummary returns a human-readable summary of the current reserve state,
// including coverage ratio against the 70% silver backing standard.
func SilverStandardSummary(reserveNg int64, totalSupplyNg int64) string {
	reserveOz := float64(reserveNg) / float64(NGPerTLR)
	supplyTLR := totalSupplyNg / NGPerTLR
	backedOz := TLRToRequiredSilverOz(supplyTLR)
	ratio := 0.0
	if supplyTLR > 0 {
		ratio = reserveOz / backedOz
	}
	return fmt.Sprintf(
		"Silver reserve: %.2f oz | Supply: %d TLR | Required backing: %.2f oz | Coverage: %.2f%% of 70%% standard",
		reserveOz, supplyTLR, backedOz, ratio*100,
	)
}

// NominalPriceMultiplier returns the multiplier for treasury nominal price offers.
// Under 70% backing, 1 TLR costs 1/0.7 = ~1.4286 oz of silver worth.
func NominalPriceMultiplier() float64 {
	return 1.0 / SilverBackingRatio
}

// --- Treasury Nominal Price Offer ---

type NominalOffer struct {
	mu           sync.Mutex
	RoundID      string `json:"round_id"`
	TotalTLR     int64  `json:"total_tlr"`
	RemainingTLR int64  `json:"remaining_tlr"`
	PricePerTLR  int64  `json:"price_per_tlr_ng"` // nominal price in ng per 1 TLR
	Active       bool   `json:"active"`
}


// NewNominalOffer handles the NewNominalOffer HTTP request.
func NewNominalOffer(roundID string, totalTLR int64) *NominalOffer {
	// Price = 1 TLR worth of ng at 70% backing = NGPerTLR / SilverBackingRatio
	ngf := float64(NGPerTLR)
	pricePerTLR := int64(ngf / SilverBackingRatio)
	return &NominalOffer{
		RoundID:      roundID,
		TotalTLR:     totalTLR,
		RemainingTLR: totalTLR,
		PricePerTLR:  pricePerTLR,
		Active:       true,
	}
}

// Buy executes a purchase of TLR at the nominal price. It deducts ng from the
// investor and credits TLR-equivalent ng (at the 70% backing rate).
func (o *NominalOffer) Buy(pubkey string, tlrAmount int64, ledger *Ledger, dataDir string) (int64, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.Active {
		return 0, fmt.Errorf("offer %s is no longer active", o.RoundID)
	}
	if tlrAmount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}
	if tlrAmount > o.RemainingTLR {
		return 0, fmt.Errorf("only %d TLR remaining in offer", o.RemainingTLR)
	}

	costNg := tlrAmount * o.PricePerTLR

	// Check investor has enough balance
	if ledger.Balance(pubkey) < costNg {
		return 0, fmt.Errorf("insufficient balance: need %d ng, have %d ng", costNg, ledger.Balance(pubkey))
	}

	// Deduct ng, credit TLR-equivalent ng (issued at 70% backing)
	ledger.Mint(pubkey, -costNg)
	issuanceNg := tlrAmount * NGPerTLR
	ledger.Mint(pubkey, issuanceNg)

	o.RemainingTLR -= tlrAmount
	if o.RemainingTLR <= 0 {
		o.Active = false
	}

	ledger.Save(dataDir)
	return issuanceNg, nil
}
