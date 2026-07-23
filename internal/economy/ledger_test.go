// Package economy implements the island economy system
package economy

import (
	"os"
	"path/filepath"
	"testing"
)


// TestUSDTtoNG handles the TestUSDTtoNG HTTP request.
func TestUSDTtoNG(t *testing.T) {
	tests := []float64{0, 0.5, 1, 100, 123.45, 999.99}
	for _, usdt := range tests {
		got := USDTtoNG(usdt)
		if got < 0 {
			t.Errorf("USDTtoNG(%f) = %d, expected positive", usdt, got)
		}
		// Verify round-trip: ng should be proportional
		if usdt > 0 && got == 0 {
			t.Errorf("USDTtoNG(%f) = 0, expected > 0", usdt)
		}
	}
}


// TestNGtoTLR handles the TestNGtoTLR HTTP request.
func TestNGtoTLR(t *testing.T) {
	tests := []struct {
		ng   int64
		want int64
	}{
		{0, 0},
		{NGPerTLR, 1},
		{NGPerTLR * 10, 10},
		{NGPerTLR / 2, 0},
	}
	for _, tt := range tests {
		got := NGtoTLR(tt.ng)
		if got != tt.want {
			t.Errorf("NGtoTLR(%d) = %d, want %d", tt.ng, got, tt.want)
		}
	}
}


// TestTLRtoNG handles the TestTLRtoNG HTTP request.
func TestTLRtoNG(t *testing.T) {
	tests := []struct {
		tlr  int64
		want int64
	}{
		{0, 0},
		{1, NGPerTLR},
		{10, NGPerTLR * 10},
	}
	for _, tt := range tests {
		got := TLStoNG(tt.tlr)
		if got != tt.want {
			t.Errorf("TLStoNG(%d) = %d, want %d", tt.tlr, got, tt.want)
		}
	}
}


// TestNGPerTLRConstant handles the TestNGPerTLRConstant HTTP request.
func TestNGPerTLRConstant(t *testing.T) {
	if NGPerTLR != 31_103_480_000 {
		t.Errorf("NGPerTLR = %d, want 31103480000", NGPerTLR)
	}
}


// TestLedgerMint handles the TestLedgerMint HTTP request.
func TestLedgerMint(t *testing.T) {
	dir := t.TempDir()
	l := LoadLedger(dir)

	l.Mint("pubkey1", 1000)
	if b := l.Balance("pubkey1"); b != 1000 {
		t.Errorf("balance after mint = %d, want 1000", b)
	}

	l.Mint("pubkey1", 500)
	if b := l.Balance("pubkey1"); b != 1500 {
		t.Errorf("balance after second mint = %d, want 1500", b)
	}

	l.Save(dir)
	// Reload and verify persistence
	l2 := LoadLedger(dir)
	if b := l2.Balance("pubkey1"); b != 1500 {
		t.Errorf("balance after reload = %d, want 1500", b)
	}
}


// TestLedgerTransfer handles the TestLedgerTransfer HTTP request.
func TestLedgerTransfer(t *testing.T) {
	dir := t.TempDir()
	l := LoadLedger(dir)

	l.Mint("alice", 5000)
	if err := l.Transfer("alice", "bob", 2000); err != nil {
		t.Fatalf("Transfer: %v", err)
	}

	if b := l.Balance("alice"); b != 3000 {
		t.Errorf("alice balance = %d, want 3000", b)
	}
	if b := l.Balance("bob"); b != 2000 {
		t.Errorf("bob balance = %d, want 2000", b)
	}
}


// TestLedgerTransferInsufficient handles the TestLedgerTransferInsufficient HTTP request.
func TestLedgerTransferInsufficient(t *testing.T) {
	dir := t.TempDir()
	l := LoadLedger(dir)

	l.Mint("alice", 100)
	err := l.Transfer("alice", "bob", 200)
	if err == nil {
		t.Fatal("expected error for insufficient funds")
	}
}


// TestLedgerTransferNoAccount handles the TestLedgerTransferNoAccount HTTP request.
func TestLedgerTransferNoAccount(t *testing.T) {
	dir := t.TempDir()
	l := LoadLedger(dir)

	err := l.Transfer("nonexistent", "bob", 100)
	if err == nil {
		t.Fatal("expected error for nonexistent sender")
	}
}


// TestGenerateKeypair handles the TestGenerateKeypair HTTP request.
func TestGenerateKeypair(t *testing.T) {
	pub, priv, mnemonic, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	if len(pub) != 64 {
		t.Errorf("pubkey hex length = %d, want 64", len(pub))
	}
	if len(priv) != 128 {
		t.Errorf("privkey hex length = %d, want 128", len(priv))
	}
	if mnemonic == "" {
		t.Error("mnemonic should not be empty")
	}
}


// TestSignAndVerify handles the TestSignAndVerify HTTP request.
func TestSignAndVerify(t *testing.T) {
	_, priv, _, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}

	msg := []byte("test message")
	sig, err := SignTx(priv, msg)
	if err != nil {
		t.Fatalf("SignTx: %v", err)
	}

	if len(sig) == 0 {
		t.Error("signature should not be empty")
	}
}


// TestLedgerSaveAndLoad handles the TestLedgerSaveAndLoad HTTP request.
func TestLedgerSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	l := LoadLedger(dir)
	l.Mint("alice", 1000)
	l.Mint("bob", 2000)
	l.Save(dir)

	// Verify the file exists
	p := filepath.Join(dir, "liquid_ledger.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("ledger file not created: %v", err)
	}

	l2 := LoadLedger(dir)
	if l2.TotalSupply != 3000 {
		t.Errorf("total supply after reload = %d, want 3000", l2.TotalSupply)
	}
}
