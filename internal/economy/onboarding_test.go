// Package economy implements the island economy system
package economy

import (
	"os"
	"path/filepath"
	"testing"
)


// TestOnboardingStatusDefault handles the TestOnboardingStatusDefault HTTP request.
func TestOnboardingStatusDefault(t *testing.T) {
	dir := t.TempDir()
	om := NewOnboardingManager(dir)
	s := om.Status("alice")
	if s.ClaimedWelcome {
		t.Error("expected false")
	}
	if s.BoughtStarter {
		t.Error("expected false")
	}
}


// TestClaimWelcomeBanknote handles the TestClaimWelcomeBanknote HTTP request.
func TestClaimWelcomeBanknote(t *testing.T) {
	dir := t.TempDir()
	om := NewOnboardingManager(dir)

	note, err := om.ClaimWelcome("bob")
	if err != nil {
		t.Fatal(err)
	}
	if note.Rarity != "common" {
		t.Errorf("expected common, got %s", note.Rarity)
	}
	if note.Holder != "bob" {
		t.Errorf("expected bob, got %s", note.Holder)
	}

	s := om.Status("bob")
	if !s.ClaimedWelcome {
		t.Error("expected claimed_welcome")
	}

	_, err = om.ClaimWelcome("bob")
	if err == nil {
		t.Error("expected error on double claim")
	}
}


// TestBuyStarterPack handles the TestBuyStarterPack HTTP request.
func TestBuyStarterPack(t *testing.T) {
	dir := t.TempDir()
	ledger := LoadLedger(dir)
	ledger.Mint("charlie", StarterPackPriceNg*10)
	ledger.Save(dir)

	om := NewOnboardingManager(dir)
	notes, err := om.BuyStarter("charlie")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 5 {
		t.Errorf("expected 5 notes, got %d", len(notes))
	}

	s := om.Status("charlie")
	if !s.BoughtStarter {
		t.Error("expected bought_starter")
	}

	_, err = om.BuyStarter("charlie")
	if err == nil {
		t.Error("expected error on double purchase")
	}
}


// TestBuyStarterInsufficientBalance handles the TestBuyStarterInsufficientBalance HTTP request.
func TestBuyStarterInsufficientBalance(t *testing.T) {
	dir := t.TempDir()
	om := NewOnboardingManager(dir)
	_, err := om.BuyStarter("dave")
	if err == nil {
		t.Error("expected error for insufficient balance")
	}
}


// TestCompleteGuide handles the TestCompleteGuide HTTP request.
func TestCompleteGuide(t *testing.T) {
	dir := t.TempDir()
	om := NewOnboardingManager(dir)

	if err := om.CompleteGuide("eve"); err != nil {
		t.Fatal(err)
	}
	s := om.Status("eve")
	if !s.CompletedGuide {
		t.Error("expected completed_guide")
	}
}


// TestIsOnboarded handles the TestIsOnboarded HTTP request.
func TestIsOnboarded(t *testing.T) {
	dir := t.TempDir()
	om := NewOnboardingManager(dir)

	if om.IsOnboarded("frank") {
		t.Error("expected false for fresh user")
	}

	om.ClaimWelcome("frank")
	if om.IsOnboarded("frank") {
		t.Error("expected false after only welcome")
	}

	ledger := LoadLedger(dir)
	ledger.Mint("frank", StarterPackPriceNg*10)
	ledger.Save(dir)
	om.BuyStarter("frank")

	om.CompleteGuide("frank")
	if !om.IsOnboarded("frank") {
		t.Error("expected true after all steps")
	}
}


// TestOnboardingPersistence handles the TestOnboardingPersistence HTTP request.
func TestOnboardingPersistence(t *testing.T) {
	dir := t.TempDir()

	om1 := NewOnboardingManager(dir)
	om1.ClaimWelcome("grace")

	om2 := NewOnboardingManager(dir)
	s := om2.Status("grace")
	if !s.ClaimedWelcome {
		t.Error("expected claimed_welcome to persist")
	}

	os.RemoveAll(filepath.Join(dir, "onboarding"))
}
