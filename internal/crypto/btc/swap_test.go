// Package btc provides Bitcoin swap functionality with HTLC support
package btc

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)


// TestCreateSwap handles the TestCreateSwap HTTP request.
func TestCreateSwap(t *testing.T) {
	r := NewRegistry()
	swap, err := r.Create("alice_pubkey", "bob_pubkey", 1000000000, "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789")
	if err != nil {
		t.Fatal(err)
	}
	if swap.Status != SwapPending {
		t.Fatalf("expected pending, got %s", swap.Status)
	}
	if swap.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}


// TestCreateSwapBadHash handles the TestCreateSwapBadHash HTTP request.
func TestCreateSwapBadHash(t *testing.T) {
	r := NewRegistry()
	_, err := r.Create("alice", "bob", 1000, "tooshort")
	if err == nil {
		t.Fatal("expected error for short hash")
	}
}


// TestClaimSwap handles the TestClaimSwap HTTP request.
func TestClaimSwap(t *testing.T) {
	r := NewRegistry()
	secret := "my_secret_value"
	h := sha256Hash(secret)
	swap, _ := r.Create("alice", "bob", 1000, h)
	claimed, err := r.Claim(swap.ID, secret)
	if err != nil {
		t.Fatal(err)
	}
	if claimed.Status != SwapClaimed {
		t.Fatalf("expected claimed, got %s", claimed.Status)
	}
	if claimed.ClaimTxID == "" {
		t.Fatal("expected claim tx id")
	}
}


// TestClaimInvalidSecret handles the TestClaimInvalidSecret HTTP request.
func TestClaimInvalidSecret(t *testing.T) {
	r := NewRegistry()
	secret := "real_secret"
	h := sha256Hash(secret)
	swap, _ := r.Create("alice", "bob", 1000, h)
	_, err := r.Claim(swap.ID, "wrong_secret")
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}


// TestRefundBeforeLock handles the TestRefundBeforeLock HTTP request.
func TestRefundBeforeLock(t *testing.T) {
	r := NewRegistry()
	h := sha256Hash("secret")
	swap, _ := r.Create("alice", "bob", 1000, h)
	_, err := r.Refund(swap.ID)
	if err == nil {
		t.Fatal("expected error before refund time")
	}
}


// TestListSwaps handles the TestListSwaps HTTP request.
func TestListSwaps(t *testing.T) {
	r := NewRegistry()
	h := sha256Hash("s")
	r.Create("a", "b", 100, h)
	r.Create("c", "d", 200, h)
	swaps := r.List()
	if len(swaps) != 2 {
		t.Fatalf("expected 2 swaps, got %d", len(swaps))
	}
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
