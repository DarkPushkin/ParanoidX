// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestNextRarity handles the TestNextRarity HTTP request.
func TestNextRarity(t *testing.T) {
	tests := []struct {
		in  string
		out string
		err bool
	}{
		{"common", "rare", false},
		{"rare", "epic", false},
		{"epic", "legendary", false},
		{"legendary", "genesis", false},
		{"genesis", "", true},
		{"unknown", "", true},
	}
	for _, tc := range tests {
		got, err := NextRarity(tc.in)
		if tc.err && err == nil {
			t.Errorf("NextRarity(%q) expected error", tc.in)
		}
		if !tc.err && got != tc.out {
			t.Errorf("NextRarity(%q) = %q, want %q", tc.in, got, tc.out)
		}
	}
}


// TestUpgradeInputCount handles the TestUpgradeInputCount HTTP request.
func TestUpgradeInputCount(t *testing.T) {
	if UpgradeInputCount != 5 {
		t.Fatalf("expected 5, got %d", UpgradeInputCount)
	}
}


// TestSumDenominations handles the TestSumDenominations HTTP request.
func TestSumDenominations(t *testing.T) {
	notes := []BanknoteV2{
		{DenominationNg: NGPerTLR},
		{DenominationNg: 5 * NGPerTLR},
	}
	total := sumDenominations(notes)
	if total != 6*NGPerTLR {
		t.Fatalf("expected 6 TLR, got %d ng", total)
	}
}


// TestCraftUpgradeWrongCount handles the TestCraftUpgradeWrongCount HTTP request.
func TestCraftUpgradeWrongCount(t *testing.T) {
	cm := NewCraftingManager()
	_, _, err := cm.CraftUpgrade("", "holder", []string{"a", "b"})
	if err == nil {
		t.Fatal("expected error for <5 inputs")
	}
}


// TestCraftUpgradeMixedRarities handles the TestCraftUpgradeMixedRarities HTTP request.
func TestCraftUpgradeMixedRarities(t *testing.T) {
	dir := t.TempDir()
	SaveBanknotesV2(dir, []BanknoteV2{
		{Serial: "MB-COMMON-001", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
		{Serial: "MB-RARE-001", DenominationNg: 5 * NGPerTLR, Rarity: "rare", Holder: "pk", Status: "active"},
		{Serial: "MB-COMMON-003", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
		{Serial: "MB-COMMON-004", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
		{Serial: "MB-COMMON-005", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
	})
	cm := NewCraftingManager()
	_, _, err := cm.CraftUpgrade(dir, "pk", []string{"MB-COMMON-001", "MB-RARE-001", "MB-COMMON-003", "MB-COMMON-004", "MB-COMMON-005"})
	if err == nil {
		t.Fatal("expected error for mixed rarities")
	}
}


// TestCraftUpgradeSuccess handles the TestCraftUpgradeSuccess HTTP request.
func TestCraftUpgradeSuccess(t *testing.T) {
	dir := t.TempDir()
	SaveBanknotesV2(dir, []BanknoteV2{
		{Serial: "MB-COMMON-010", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
		{Serial: "MB-COMMON-011", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
		{Serial: "MB-COMMON-012", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
		{Serial: "MB-COMMON-013", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
		{Serial: "MB-COMMON-014", DenominationNg: NGPerTLR, Rarity: "common", Holder: "pk", Status: "active"},
	})

	cm := NewCraftingManager()
	burnt, upgraded, err := cm.CraftUpgrade(dir, "pk", []string{"MB-COMMON-010", "MB-COMMON-011", "MB-COMMON-012", "MB-COMMON-013", "MB-COMMON-014"})
	if err != nil {
		t.Fatal(err)
	}
	if len(burnt) != 5 {
		t.Fatalf("expected 5 burnt, got %d", len(burnt))
	}
	if upgraded.Rarity != "rare" {
		t.Fatalf("expected rare upgrade, got %s", upgraded.Rarity)
	}
	if upgraded.Holder != "pk" {
		t.Fatalf("expected holder pk, got %s", upgraded.Holder)
	}

	// Verify banknotes are burned
	notes, _ := LoadBanknotesV2(dir)
	burnedCount := 0
	for _, n := range notes {
		if n.Status == "burned" {
			burnedCount++
		}
	}
	if burnedCount != 5 {
		t.Fatalf("expected 5 burned, got %d", burnedCount)
	}
}


// TestGetLeaderboard handles the TestGetLeaderboard HTTP request.
func TestGetLeaderboard(t *testing.T) {
	notes := []BanknoteV2{
		{Serial: "A", DenominationNg: NGPerTLR, Rarity: "common", Multiplier: 1, Holder: "alice", Status: "active"},
		{Serial: "B", DenominationNg: 5 * NGPerTLR, Rarity: "rare", Multiplier: 2, Holder: "bob", Status: "active"},
		{Serial: "C", DenominationNg: NGPerTLR, Rarity: "common", Multiplier: 1, Holder: "alice", Status: "active"},
		{Serial: "D", DenominationNg: 25 * NGPerTLR, Rarity: "epic", Multiplier: 3, Holder: "charlie", Status: "active"},
		{Serial: "E", DenominationNg: NGPerTLR, Rarity: "common", Multiplier: 1, Holder: "alice", Status: "burned"},
	}

	lb := GetLeaderboard(notes, 3)
	if len(lb) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(lb))
	}
	if lb[0].Holder != "charlie" {
		t.Errorf("expected charlie first, got %s", lb[0].Holder)
	}
	// alice: 2*NGPerTLR, bob: 10*NGPerTLR, charlie: 75*NGPerTLR
	if lb[0].TotalValue != 25*NGPerTLR*3 {
		t.Errorf("charlie value = %d, want %d", lb[0].TotalValue, 25*NGPerTLR*3)
	}
}


// TestGetLeaderboardLimit handles the TestGetLeaderboardLimit HTTP request.
func TestGetLeaderboardLimit(t *testing.T) {
	notes := []BanknoteV2{
		{Serial: "A", DenominationNg: NGPerTLR, Rarity: "common", Multiplier: 1, Holder: "a", Status: "active"},
		{Serial: "B", DenominationNg: NGPerTLR, Rarity: "common", Multiplier: 1, Holder: "b", Status: "active"},
		{Serial: "C", DenominationNg: NGPerTLR, Rarity: "common", Multiplier: 1, Holder: "c", Status: "active"},
		{Serial: "D", DenominationNg: NGPerTLR, Rarity: "common", Multiplier: 1, Holder: "d", Status: "active"},
	}
	lb := GetLeaderboard(notes, 2)
	if len(lb) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(lb))
	}
}


// TestGetLeaderboardEmpty handles the TestGetLeaderboardEmpty HTTP request.
func TestGetLeaderboardEmpty(t *testing.T) {
	lb := GetLeaderboard([]BanknoteV2{}, 10)
	if len(lb) != 0 {
		t.Fatalf("expected empty, got %d", len(lb))
	}
}


// TestAutoReinvestEnable handles the TestAutoReinvestEnable HTTP request.
func TestAutoReinvestEnable(t *testing.T) {
	ar := NewAutoReinvestManager()
	if ar.IsEnabled("pk") {
		t.Fatal("should be disabled initially")
	}
	ar.SetEnabled("pk", true)
	if !ar.IsEnabled("pk") {
		t.Fatal("should be enabled after set")
	}
}


// TestAutoReinvestNotEnabled handles the TestAutoReinvestNotEnabled HTTP request.
func TestAutoReinvestNotEnabled(t *testing.T) {
	ar := NewAutoReinvestManager()
	_, err := ar.Reinvest("", "pk", NGPerTLR)
	if err == nil {
		t.Fatal("expected error when not enabled")
	}
}


// TestAutoReinvestMinimum handles the TestAutoReinvestMinimum HTTP request.
func TestAutoReinvestMinimum(t *testing.T) {
	ar := NewAutoReinvestManager()
	ar.SetEnabled("pk2", true)
	_, err := ar.Reinvest("", "pk2", NGPerTLR-1)
	if err == nil {
		t.Fatal("expected error for below minimum")
	}
}
