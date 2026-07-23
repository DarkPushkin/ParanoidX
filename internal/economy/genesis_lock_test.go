// Package economy implements the island economy system
package economy

import (
	"testing"
	"time"
)


// TestNewGenesisLockState handles the TestNewGenesisLockState HTTP request.
func TestNewGenesisLockState(t *testing.T) {
	gl := NewGenesisLockState()
	if !gl.IsFrozen() {
		t.Fatal("genesis should start frozen")
	}
}


// TestRegisterGenesisCard handles the TestRegisterGenesisCard HTTP request.
func TestRegisterGenesisCard(t *testing.T) {
	gl := NewGenesisLockState()
	gl.RegisterGenesisCard("MB-GENESIS-001")
	if len(gl.GenesisCards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(gl.GenesisCards))
	}
	// Duplicate should not increase count
	gl.RegisterGenesisCard("MB-GENESIS-001")
	if len(gl.GenesisCards) != 1 {
		t.Fatalf("expected 1 card after duplicate, got %d", len(gl.GenesisCards))
	}
}


// TestGenesisAccrueFrozenDividend handles the TestGenesisAccrueFrozenDividend HTTP request.
func TestGenesisAccrueFrozenDividend(t *testing.T) {
	dir := t.TempDir()
	gl := NewGenesisLockState()

	// Create a genesis banknote and a common banknote
	banknotes := []BanknoteV2{
		{Serial: "MB-GENESIS-001", DenominationNg: NGPerTLR, Rarity: "genesis", Holder: "alice", Status: "genesis_locked", FrozenNg: NGPerTLR},
		{Serial: "MB-COMMON-001", DenominationNg: NGPerTLR, Rarity: "common", Holder: "bob", Status: "active", FrozenNg: NGPerTLR},
	}
	SaveBanknotesV2(dir, banknotes)
	gl.RegisterGenesisCard("MB-GENESIS-001")

	// Genesis weight = NGPerTLR * 20 = 20*NGPerTLR
	// Common weight = NGPerTLR * 1 = NGPerTLR
	// Total = 21*NGPerTLR
	// Genesis share = 20/21 = ~95.24%
	// Dividends = 21*NGPerTLR
	// Frozen = 21*NGPerTLR * 20/21 = 20*NGPerTLR
	frozen, err := gl.AccrueFrozenDividend(dir, 21*NGPerTLR, "round-1")
	if err != nil {
		t.Fatal(err)
	}
	expectedFrozen := 20 * NGPerTLR
	if frozen != expectedFrozen {
		t.Fatalf("expected %d frozen, got %d", expectedFrozen, frozen)
	}
	if gl.FrozenDividendPoolNg != expectedFrozen {
		t.Fatalf("pool: expected %d, got %d", expectedFrozen, gl.FrozenDividendPoolNg)
	}
}


// TestGenesisCheckSurplusBelowThreshold handles the TestGenesisCheckSurplusBelowThreshold HTTP request.
func TestGenesisCheckSurplusBelowThreshold(t *testing.T) {
	gl := NewGenesisLockState()
	// Treasury well below safety threshold
	surplus := gl.CheckSurplus(100 * NGPerTLR)
	if surplus {
		t.Fatal("should not unlock with tiny treasury")
	}
	if gl.Unlocked {
		t.Fatal("should remain locked")
	}
}


// TestGenesisCheckSurplusAboveThreshold handles the TestGenesisCheckSurplusAboveThreshold HTTP request.
func TestGenesisCheckSurplusAboveThreshold(t *testing.T) {
	gl := NewGenesisLockState()
	// Treasury far above safety threshold
	surplus := gl.CheckSurplus(100000 * NGPerTLR)
	if !surplus {
		t.Fatal("should unlock with sufficient treasury")
	}
	if !gl.Unlocked {
		t.Fatal("should be unlocked")
	}
}


// TestGenesisDistributeFrozenDividends handles the TestGenesisDistributeFrozenDividends HTTP request.
func TestGenesisDistributeFrozenDividends(t *testing.T) {
	dir := t.TempDir()
	gl := NewGenesisLockState()

	banknotes := []BanknoteV2{
		{Serial: "MB-GENESIS-001", DenominationNg: NGPerTLR, Rarity: "genesis", Holder: "alice", Status: "genesis_locked", FrozenNg: NGPerTLR},
	}
	SaveBanknotesV2(dir, banknotes)

	// Seed frozen dividends
	gl.AccrueFrozenDividend(dir, 10*NGPerTLR, "round-1")
	gl.CheckSurplus(100000 * NGPerTLR) // unlock

	distributed, err := gl.DistributeFrozenDividends(dir)
	if err != nil {
		t.Fatal(err)
	}
	if distributed <= 0 {
		t.Fatalf("expected positive distribution, got %d", distributed)
	}

	// Check alice balance
	ledger := LoadLedger(dir)
	if ledger.Balance("alice") <= 0 {
		t.Fatal("alice should have received frozen dividends")
	}
}


// TestGenesisLockSaveLoad handles the TestGenesisLockSaveLoad HTTP request.
func TestGenesisLockSaveLoad(t *testing.T) {
	dir := t.TempDir()

	// Need genesis banknote in registry for accrual calculation
	banknotes := []BanknoteV2{
		{Serial: "MB-GENESIS-001", DenominationNg: NGPerTLR, Rarity: "genesis", Holder: "alice", Status: "genesis_locked", FrozenNg: NGPerTLR},
	}
	SaveBanknotesV2(dir, banknotes)

	gl := NewGenesisLockState()
	gl.RegisterGenesisCard("MB-GENESIS-001")
	_ = gl.CheckSurplus(100000 * NGPerTLR) // accrue

	// Manually set frozen dividends (simulates accrual)
	gl.mu.Lock()
	gl.FrozenDividendPoolNg = 100 * NGPerTLR
	gl.DividendAccruals = append(gl.DividendAccruals, FrozenAccrual{Round: "round-1", TotalNg: 100 * NGPerTLR})
	gl.mu.Unlock()

	gl.Save(dir)

	// Load fresh
	gl2 := LoadGenesisLock(dir)
	if len(gl2.GenesisCards) != 1 {
		t.Fatalf("expected 1 card after load, got %d", len(gl2.GenesisCards))
	}
	if !gl2.Unlocked {
		t.Fatal("should still be unlocked after load")
	}
	if gl2.FrozenDividendPoolNg <= 0 {
		t.Fatal("frozen dividends should survive save/load")
	}
}


// TestGenesisSummary handles the TestGenesisSummary HTTP request.
func TestGenesisSummary(t *testing.T) {
	gl := NewGenesisLockState()
	gl.RegisterGenesisCard("MB-GENESIS-001")
	s := gl.Summary(500 * NGPerTLR)
	if !s.Frozen {
		t.Fatal("should show frozen")
	}
	if s.ProgressPct <= 0 {
		t.Fatal("should have some progress")
	}
}


// TestTreasuryMonthlyOpsDefault handles the TestTreasuryMonthlyOpsDefault HTTP request.
func TestTreasuryMonthlyOpsDefault(t *testing.T) {
	if TreasuryMonthlyOpsNg != 1000*NGPerTLR {
		t.Fatalf("expected default 1000*NGPerTLR, got %d", TreasuryMonthlyOpsNg)
	}
}


// TestGenesisSafetyThreshold handles the TestGenesisSafetyThreshold HTTP request.
func TestGenesisSafetyThreshold(t *testing.T) {
	expected := 12 * TreasuryMonthlyOpsNg
	if GenesisSafetyThreshold() != expected {
		t.Fatalf("expected %d, got %d", expected, GenesisSafetyThreshold())
	}
	time.Now() // just to use time
}
