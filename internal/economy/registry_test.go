// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestCalculateTreasurySplit_Thin handles the TestCalculateTreasurySplit_Thin HTTP request.
func TestCalculateTreasurySplit_Thin(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	// 1x monthly ops — thin treasury
	ops, reserve, insurance, toPeople := CalculateTreasurySplit(cfg.MonthlyOpsNg, cfg)

	if ops != cfg.MonthlyOpsNg*75/100 {
		t.Errorf("thin ops = %d, want %d", ops, cfg.MonthlyOpsNg*75/100)
	}
	if reserve != cfg.MonthlyOpsNg*25/100 {
		t.Errorf("thin reserve = %d, want %d", reserve, cfg.MonthlyOpsNg*25/100)
	}
	if insurance != 0 {
		t.Errorf("thin insurance = %d, want 0", insurance)
	}
	if toPeople != 0 {
		t.Errorf("thin toPeople = %d, want 0", toPeople)
	}
}


// TestCalculateTreasurySplit_Normal handles the TestCalculateTreasurySplit_Normal HTTP request.
func TestCalculateTreasurySplit_Normal(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	// 4x monthly ops — normal treasury
	treasury := cfg.MonthlyOpsNg * 4
	ops, reserve, insurance, toPeople := CalculateTreasurySplit(treasury, cfg)

	if ops != treasury*50/100 {
		t.Errorf("normal ops = %d, want %d", ops, treasury*50/100)
	}
	if reserve != treasury*25/100 {
		t.Errorf("normal reserve = %d, want %d", reserve, treasury*25/100)
	}
	if insurance != treasury*25/100 {
		t.Errorf("normal insurance = %d, want %d", insurance, treasury*25/100)
	}
	if toPeople != 0 {
		t.Errorf("normal toPeople = %d, want 0", toPeople)
	}
}


// TestCalculateTreasurySplit_Fat handles the TestCalculateTreasurySplit_Fat HTTP request.
func TestCalculateTreasurySplit_Fat(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	treasury := cfg.MonthlyOpsNg * 8
	ops, reserve, _, toPeople := CalculateTreasurySplit(treasury, cfg)

	if ops != treasury*20/100 {
		t.Errorf("fat ops = %d, want %d", ops, treasury*20/100)
	}
	if reserve != treasury*30/100 {
		t.Errorf("fat reserve = %d, want %d", reserve, treasury*30/100)
	}
	if toPeople != treasury*50/100 {
		t.Errorf("fat toPeople = %d, want %d", toPeople, treasury*50/100)
	}
}


// TestCalculateTreasurySplit_VeryFat handles the TestCalculateTreasurySplit_VeryFat HTTP request.
func TestCalculateTreasurySplit_VeryFat(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	treasury := cfg.MonthlyOpsNg * 15
	ops, reserve, _, toPeople := CalculateTreasurySplit(treasury, cfg)

	if ops != treasury*10/100 {
		t.Errorf("veryfat ops = %d, want %d", ops, treasury*10/100)
	}
	if reserve != treasury*20/100 {
		t.Errorf("veryfat reserve = %d, want %d", reserve, treasury*20/100)
	}
	if toPeople != treasury*30/100 {
		t.Errorf("veryfat toPeople = %d, want %d", toPeople, treasury*30/100)
	}
}


// TestCalculateTreasurySplit_Zero handles the TestCalculateTreasurySplit_Zero HTTP request.
func TestCalculateTreasurySplit_Zero(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	ops, reserve, insurance, toPeople := CalculateTreasurySplit(0, cfg)

	if ops != 0 || reserve != 0 || insurance != 0 || toPeople != 0 {
		t.Errorf("zero treasury split = %d/%d/%d/%d, want all 0", ops, reserve, insurance, toPeople)
	}
}


// TestCalculateWeightedDenom handles the TestCalculateWeightedDenom HTTP request.
func TestCalculateWeightedDenom(t *testing.T) {
	banknotes := []BanknoteV2{
		{DenominationNg: NGPerTLR, Multiplier: 1, Status: "active"},
		{DenominationNg: NGPerTLR * 5, Multiplier: 2, Status: "active"},
	}
	total := CalculateWeightedDenom(banknotes)
	expected := NGPerTLR*1 + NGPerTLR*5*2
	if total != expected {
		t.Errorf("weighted denom = %d, want %d", total, expected)
	}
}


// TestCalculateDividendForHolder handles the TestCalculateDividendForHolder HTTP request.
func TestCalculateDividendForHolder(t *testing.T) {
	banknotes := []BanknoteV2{
		{Serial: "MB-COMMON-001", DenominationNg: NGPerTLR, Multiplier: 1, Status: "active", Holder: "alice"},
		{Serial: "MB-RARE-001", DenominationNg: NGPerTLR * 5, Multiplier: 2, Status: "active", Holder: "bob"},
	}
	pool := int64(100_000_000_000)

	divs := CalculateDividendForHolder(banknotes, pool)
	if divs == nil || len(divs) == 0 {
		t.Fatal("expected dividends, got nil/empty")
	}

	totalDistributed := int64(0)
	for _, d := range divs {
		totalDistributed += d
	}
	if totalDistributed > pool {
		t.Errorf("total distributed %d > pool %d", totalDistributed, pool)
	}
}


// TestDetectRarityFromSerial handles the TestDetectRarityFromSerial HTTP request.
func TestDetectRarityFromSerial(t *testing.T) {
	tests := []struct {
		serial string
		want   string
	}{
		{"MB-COMMON-2026-000001", "common"},
		{"MB-RARE-2026-000001", "rare"},
		{"MB-EPIC-2026-000001", "epic"},
		{"MB-LEGENDARY-2026-000001", "legendary"},
		{"MB-GENESIS-001", "genesis"},
		{"invalid", "common"},
	}
	for _, tt := range tests {
		got := detectRarityFromSerial(tt.serial)
		if got != tt.want {
			t.Errorf("detectRarityFromSerial(%q) = %q, want %q", tt.serial, got, tt.want)
		}
	}
}


// TestRarityMultiplierValues handles the TestRarityMultiplierValues HTTP request.
func TestRarityMultiplierValues(t *testing.T) {
	expected := map[string]int{
		"common":    1,
		"rare":      2,
		"epic":      3,
		"legendary": 4,
		"genesis":   5,
	}
	for rarity, want := range expected {
		if got := rarityMultiplier[rarity]; got != want {
			t.Errorf("rarityMultiplier[%q] = %d, want %d", rarity, got, want)
		}
	}
}
