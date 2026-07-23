// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestDetectTier handles the TestDetectTier HTTP request.
func TestDetectTier(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	tests := []struct {
		name string
		ng   int64
		want TreasuryTier
	}{
		{"zero", 0, TierThin},
		{"thin", cfg.MonthlyOpsNg * 2, TierThin},
		{"normal low", cfg.MonthlyOpsNg * 3, TierNormal},
		{"normal high", cfg.MonthlyOpsNg * 5, TierNormal},
		{"fat low", cfg.MonthlyOpsNg * 6, TierFat},
		{"fat high", cfg.MonthlyOpsNg * 11, TierFat},
		{"very fat", cfg.MonthlyOpsNg * 12, TierVeryFat},
		{"mega fat", cfg.MonthlyOpsNg * 100, TierVeryFat},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectTier(tc.ng, cfg)
			if got != tc.want {
				t.Errorf("DetectTier(%d) = %v, want %v", tc.ng, got, tc.want)
			}
		})
	}
}


// TestTierStringOutput handles the TestTierStringOutput HTTP request.
func TestTierStringOutput(t *testing.T) {
	tests := []struct {
		tier TreasuryTier
		want string
	}{
		{TierThin, "thin"},
		{TierNormal, "normal"},
		{TierFat, "fat"},
		{TierVeryFat, "very_fat"},
	}
	for _, tc := range tests {
		if tc.tier.String() != tc.want {
			t.Errorf("got %s, want %s", tc.tier.String(), tc.want)
		}
	}
}


// TestDefaultScheduleHasSets handles the TestDefaultScheduleHasSets HTTP request.
func TestDefaultScheduleHasSets(t *testing.T) {
	s := NewDefaultSchedule()
	if len(s.Sets) != 3 {
		t.Fatalf("expected 3 sets, got %d", len(s.Sets))
	}
	if s.Sets[0].Tier != TierNormal {
		t.Errorf("first set should be TierNormal")
	}
	if s.Sets[2].Tier != TierVeryFat {
		t.Errorf("third set should be TierVeryFat")
	}
}


// TestCheckAndMintNoTriggerOnThin handles the TestCheckAndMintNoTriggerOnThin HTTP request.
func TestCheckAndMintNoTriggerOnThin(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	s := NewDefaultSchedule()

	// Thin tier — no minting
	notes, triggers, err := s.CheckAndMint(cfg.MonthlyOpsNg, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes on thin tier, got %d", len(notes))
	}
	if len(triggers) != 0 {
		t.Fatalf("expected 0 triggers, got %d", len(triggers))
	}
}


// TestCheckAndMintNormalTier handles the TestCheckAndMintNormalTier HTTP request.
func TestCheckAndMintNormalTier(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	s := NewDefaultSchedule()

	// Normal tier: 3x monthly ops
	notes, triggers, err := s.CheckAndMint(cfg.MonthlyOpsNg*3, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 3 {
		t.Fatalf("expected 3 notes (common set), got %d", len(notes))
	}
	if len(triggers) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(triggers))
	}
	if notes[0].Holder != "treasury" {
		t.Errorf("expected treasury holder, got %s", notes[0].Holder)
	}
	if notes[0].Rarity != "common" {
		t.Errorf("expected common rarity, got %s", notes[0].Rarity)
	}
}


// TestCheckAndMintIdempotent handles the TestCheckAndMintIdempotent HTTP request.
func TestCheckAndMintIdempotent(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	s := NewDefaultSchedule()

	// First call at normal tier
	s.CheckAndMint(cfg.MonthlyOpsNg*3, cfg, "")

	// Second call same tier — no minting
	notes, triggers, err := s.CheckAndMint(cfg.MonthlyOpsNg*3, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes on second call, got %d", len(notes))
	}
	if len(triggers) != 0 {
		t.Fatalf("expected 0 triggers on second call, got %d", len(triggers))
	}
}


// TestCheckAndMintCumulativeTiers handles the TestCheckAndMintCumulativeTiers HTTP request.
func TestCheckAndMintCumulativeTiers(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	s := NewDefaultSchedule()

	// Jump directly to very fat: should trigger all 3 sets (normal + fat + very fat)
	notes, triggers, err := s.CheckAndMint(cfg.MonthlyOpsNg*12, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 6 {
		t.Fatalf("expected 6 notes (3+2+1), got %d", len(notes))
	}
	if len(triggers) != 3 {
		t.Fatalf("expected 3 triggers, got %d", len(triggers))
	}
}


// TestMergeMintedBanknotes handles the TestMergeMintedBanknotes HTTP request.
func TestMergeMintedBanknotes(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultTreasuryConfig()
	s := NewDefaultSchedule()

	notes, _, err := s.CheckAndMint(cfg.MonthlyOpsNg*6, cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	// Pre-create empty banknotes registry
	SaveBanknotesV2(dir, []BanknoteV2{})

	err = MergeMintedBanknotes(dir, notes)
	if err != nil {
		t.Fatal(err)
	}

	banknotes, err := LoadBanknotesV2(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(banknotes) != 5 {
		t.Fatalf("expected 5 banknotes in registry, got %d", len(banknotes))
	}
}


// TestSaveAndLoadSchedule handles the TestSaveAndLoadSchedule HTTP request.
func TestSaveAndLoadSchedule(t *testing.T) {
	dir := t.TempDir()
	s := NewDefaultSchedule()
	s.LastTier = TierFat
	s.Save(dir)

	loaded := LoadOrCreateSchedule(dir)
	if loaded.LastTier != TierFat {
		t.Fatalf("expected TierFat, got %v", loaded.LastTier)
	}
}


// TestDetectTierEdgeCases handles the TestDetectTierEdgeCases HTTP request.
func TestDetectTierEdgeCases(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	if got := DetectTier(-1, cfg); got != TierThin {
		t.Errorf("negative: expected thin, got %v", got)
	}
	tier := DetectTier(cfg.MonthlyOpsNg*12, cfg)
	if tier != TierVeryFat {
		t.Errorf("12x: expected very_fat, got %v", tier)
	}
	tier = DetectTier(cfg.MonthlyOpsNg*12-1, cfg)
	if tier != TierFat {
		t.Errorf("12x-1: expected fat, got %v", tier)
	}
}


// TestMergeNoBanknotes handles the TestMergeNoBanknotes HTTP request.
func TestMergeNoBanknotes(t *testing.T) {
	dir := t.TempDir()
	SaveBanknotesV2(dir, []BanknoteV2{})
	err := MergeMintedBanknotes(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
}


// TestCheckAndMintTierTransition handles the TestCheckAndMintTierTransition HTTP request.
func TestCheckAndMintTierTransition(t *testing.T) {
	cfg := DefaultTreasuryConfig()
	s := NewDefaultSchedule()

	// Start at thin
	s.CheckAndMint(cfg.MonthlyOpsNg, cfg, "")
	if s.LastTier != TierThin {
		t.Errorf("expected thin, got %v", s.LastTier)
	}

	// Move to normal
	s.CheckAndMint(cfg.MonthlyOpsNg*4, cfg, "")
	if s.LastTier != TierNormal {
		t.Errorf("expected normal, got %v", s.LastTier)
	}
}


// TestTierStringUnknown handles the TestTierStringUnknown HTTP request.
func TestTierStringUnknown(t *testing.T) {
	var u TreasuryTier = 99
	if u.String() != "unknown" {
		t.Errorf("expected unknown, got %s", u.String())
	}
}
