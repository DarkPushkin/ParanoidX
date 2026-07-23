// Package economy implements the island economy system
package economy

import (
	"testing"
	"time"
)


// TestPickTierReturnsValid handles the TestPickTierReturnsValid HTTP request.
func TestPickTierReturnsValid(t *testing.T) {
	seen := map[WheelRewardTier]bool{}
	for i := 0; i < 1000; i++ {
		tier := pickTier()
		seen[tier] = true
	}
	if len(seen) < 4 {
		t.Fatalf("expected all 4 tiers after 1000 spins, got %d: %v", len(seen), seen)
	}
}


// TestSpinWheelReturnsValid handles the TestSpinWheelReturnsValid HTTP request.
func TestSpinWheelReturnsValid(t *testing.T) {
	r := SpinWheel()
	if r.Label == "" {
		t.Fatal("expected non-empty label")
	}
}


// TestSpinWheelNgAwardNonNegative handles the TestSpinWheelNgAwardNonNegative HTTP request.
func TestSpinWheelNgAwardNonNegative(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := SpinWheel()
		if r.NgAward < 0 {
			t.Fatalf("negative ng award: %d", r.NgAward)
		}
	}
}


// TestCanSpinInitially handles the TestCanSpinInitially HTTP request.
func TestCanSpinInitially(t *testing.T) {
	ws := NewWheelSpinner()
	if !ws.CanSpin("pk-1") {
		t.Fatal("new pubkey should be able to spin")
	}
}


// TestSpinOnce handles the TestSpinOnce HTTP request.
func TestSpinOnce(t *testing.T) {
	ws := NewWheelSpinner()
	r, err := ws.Spin("pk-1")
	if err != nil {
		t.Fatal(err)
	}
	if r.Label == "" {
		t.Fatal("expected reward label")
	}
}


// TestSpinTwiceRejected handles the TestSpinTwiceRejected HTTP request.
func TestSpinTwiceRejected(t *testing.T) {
	ws := NewWheelSpinner()
	_, err := ws.Spin("pk-2")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ws.Spin("pk-2")
	if err == nil {
		t.Fatal("expected error on second same-day spin")
	}
}


// TestDifferentPubkeyCanSpin handles the TestDifferentPubkeyCanSpin HTTP request.
func TestDifferentPubkeyCanSpin(t *testing.T) {
	ws := NewWheelSpinner()
	ws.Spin("pk-a")
	if !ws.CanSpin("pk-b") {
		t.Fatal("different pubkey should be able to spin same day")
	}
}


// TestNextSpinTime handles the TestNextSpinTime HTTP request.
func TestNextSpinTime(t *testing.T) {
	ws := NewWheelSpinner()
	ws.Spin("pk-3")
	next := ws.NextSpinTime("pk-3")
	if next.IsZero() {
		t.Fatal("expected non-zero next spin time")
	}
	now := time.Now()
	if next.Before(now) {
		t.Fatal("next spin time should be in the future")
	}
}


// TestNextSpinTimeForNeverSpun handles the TestNextSpinTimeForNeverSpun HTTP request.
func TestNextSpinTimeForNeverSpun(t *testing.T) {
	ws := NewWheelSpinner()
	next := ws.NextSpinTime("pk-new")
	if !next.IsZero() {
		t.Fatal("expected zero time for never-spun pubkey")
	}
}
