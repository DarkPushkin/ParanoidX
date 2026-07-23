// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestNewGenesisICOManager handles the TestNewGenesisICOManager HTTP request.
func TestNewGenesisICOManager(t *testing.T) {
	ico := NewGenesisICOManager()
	if ico == nil {
		t.Fatal("expected non-nil manager")
	}
	if ico.CurrentRound != 0 {
		t.Fatalf("expected round 0, got %d", ico.CurrentRound)
	}
}


// TestICORoundPrices handles the TestICORoundPrices HTTP request.
func TestICORoundPrices(t *testing.T) {
	if len(ICORoundPrices) != 4 {
		t.Fatalf("expected 4 round prices, got %d", len(ICORoundPrices))
	}
	for i, p := range ICORoundPrices {
		if p <= 0 {
			t.Fatalf("round %d price should be positive, got %d", i+1, p)
		}
	}
}


// TestGenesisTokensPerCard handles the TestGenesisTokensPerCard HTTP request.
func TestGenesisTokensPerCard(t *testing.T) {
	if GenesisTokensPerCard != 1_000_000 {
		t.Fatalf("expected 1,000,000 tokens per card, got %d", GenesisTokensPerCard)
	}
}


// TestICOStartRequiresGenesisCards handles the TestICOStartRequiresGenesisCards HTTP request.
func TestICOStartRequiresGenesisCards(t *testing.T) {
	dir := t.TempDir()
	ico := NewGenesisICOManager()
	err := ico.StartICO(dir)
	if err == nil {
		t.Fatal("expected error with no genesis cards")
	}
}


// TestICOStartSuccess handles the TestICOStartSuccess HTTP request.
func TestICOStartSuccess(t *testing.T) {
	dir := t.TempDir()
	banknotes := []BanknoteV2{
		{Serial: "MB-GENESIS-001", DenominationNg: NGPerTLR * 100, Rarity: "genesis", Holder: "treasury", Status: "genesis_locked"},
	}
	SaveBanknotesV2(dir, banknotes)

	ico := NewGenesisICOManager()
	if err := ico.StartICO(dir); err != nil {
		t.Fatal(err)
	}
	if !ico.Status().Started {
		t.Fatal("ICO should be started")
	}
}


// TestBuyTokensInsufficientBalance handles the TestBuyTokensInsufficientBalance HTTP request.
func TestBuyTokensInsufficientBalance(t *testing.T) {
	dir := t.TempDir()
	banknotes := []BanknoteV2{
		{Serial: "MB-GENESIS-001", DenominationNg: NGPerTLR * 100, Rarity: "genesis", Holder: "treasury", Status: "genesis_locked"},
	}
	SaveBanknotesV2(dir, banknotes)
	ledger := LoadLedger(dir)

	ico := NewGenesisICOManager()
	ico.StartICO(dir)

	_, err := ico.BuyTokens(dir, "poor_buyer", 1, ledger)
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
}


// TestBuyTokensSuccess handles the TestBuyTokensSuccess HTTP request.
func TestBuyTokensSuccess(t *testing.T) {
	dir := t.TempDir()
	banknotes := []BanknoteV2{
		{Serial: "MB-GENESIS-001", DenominationNg: NGPerTLR * 100, Rarity: "genesis", Holder: "treasury", Status: "genesis_locked", FrozenNg: NGPerTLR * 100},
	}
	SaveBanknotesV2(dir, banknotes)

	ledger := LoadLedger(dir)
	cost := ICORoundPrices[0] * 10 // 10 tokens
	ledger.Mint("investor", cost*2)
	ledger.Save(dir)

	ico := NewGenesisICOManager()
	ico.StartICO(dir)

	tokens, err := ico.BuyTokens(dir, "investor", 10, ledger)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 10 {
		t.Fatalf("expected 10 tokens, got %d", len(tokens))
	}
	if ico.TotalRaisedNg <= 0 {
		t.Fatal("total raised should be positive")
	}
}


// TestICOStatus handles the TestICOStatus HTTP request.
func TestICOStatus(t *testing.T) {
	ico := NewGenesisICOManager()
	s := ico.Status()
	if s.Started {
		t.Fatal("should not be started yet")
	}
}


// TestICOHolderTokens handles the TestICOHolderTokens HTTP request.
func TestICOHolderTokens(t *testing.T) {
	dir := t.TempDir()
	banknotes := []BanknoteV2{
		{Serial: "MB-GENESIS-001", DenominationNg: NGPerTLR * 100, Rarity: "genesis", Holder: "treasury", Status: "genesis_locked"},
	}
	SaveBanknotesV2(dir, banknotes)

	ledger := LoadLedger(dir)
	ledger.Mint("alice", 1000*NGPerTLR)
	ledger.Save(dir)

	ico := NewGenesisICOManager()
	ico.StartICO(dir)
	ico.BuyTokens(dir, "alice", 5, ledger)

	tokens := ico.HolderTokens("alice")
	if len(tokens) != 5 {
		t.Fatalf("expected 5 tokens for alice, got %d", len(tokens))
	}

	noTokens := ico.HolderTokens("bob")
	if len(noTokens) != 0 {
		t.Fatalf("expected 0 tokens for bob, got %d", len(noTokens))
	}
}


// TestBuyTokensNotStarted handles the TestBuyTokensNotStarted HTTP request.
func TestBuyTokensNotStarted(t *testing.T) {
	dir := t.TempDir()
	ledger := LoadLedger(dir)
	ledger.Mint("investor", 1000*NGPerTLR)
	ico := NewGenesisICOManager()
	_, err := ico.BuyTokens(dir, "investor", 1, ledger)
	if err == nil {
		t.Fatal("expected error: ICO not started")
	}
}
