// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestTierCosts handles the TestTierCosts HTTP request.
func TestTierCosts(t *testing.T) {
	tests := []struct {
		tier Tier
		cost int64
	}{
		{TierColonist, 0},
		{TierCitizen, 2_000_000_000},
		{TierAristocrat, 20_000_000_000},
	}
	for _, tc := range tests {
		got := tc.tier.MonthlyCostNg()
		if got != tc.cost {
			t.Errorf("%s cost: got %d, want %d", tc.tier, got, tc.cost)
		}
	}
}


// TestTierBenefits handles the TestTierBenefits HTTP request.
func TestTierBenefits(t *testing.T) {
	b := TierColonist.Benefits()
	if b.CanVote {
		t.Error("colonist should not be able to vote")
	}

	b = TierCitizen.Benefits()
	if !b.CanVote {
		t.Error("citizen should be able to vote")
	}
	if b.P2pFeePercent != 0.0 {
		t.Errorf("citizen p2p fee should be 0, got %f", b.P2pFeePercent)
	}

	b = TierAristocrat.Benefits()
	if b.DividendMult != 4 {
		t.Errorf("aristocrat dividend mult should be 4, got %d", b.DividendMult)
	}
	if !b.AutoBid {
		t.Error("aristocrat should have auto-bid")
	}
	if !b.EarlyAccess {
		t.Error("aristocrat should have early access")
	}
}


// TestSubscriptionManager handles the TestSubscriptionManager HTTP request.
func TestSubscriptionManager(t *testing.T) {
	m := NewSubscriptionManager()

	s := m.GetOrCreate("test-pubkey-1")
	if s.Tier != TierColonist {
		t.Fatalf("expected colonist, got %s", s.Tier)
	}
	if !s.IsActive() {
		t.Fatal("new subscription should be active")
	}
	if s.DaysRemaining() < 360 {
		t.Fatalf("expected ~365 days, got %d", s.DaysRemaining())
	}
}


// TestSubscribeUpgrade handles the TestSubscribeUpgrade HTTP request.
func TestSubscribeUpgrade(t *testing.T) {
	m := NewSubscriptionManager()

	s, err := m.Subscribe("pk-1", TierCitizen, 30)
	if err != nil {
		t.Fatal(err)
	}
	if s.Tier != TierCitizen {
		t.Fatalf("expected citizen, got %s", s.Tier)
	}
}


// TestSubscribeExtendsExisting handles the TestSubscribeExtendsExisting HTTP request.
func TestSubscribeExtendsExisting(t *testing.T) {
	m := NewSubscriptionManager()
	m.Subscribe("pk-2", TierCitizen, 30)
	initial := m.Subs["pk-2"].DaysRemaining()

	m.Subscribe("pk-2", TierAristocrat, 60)
	after := m.Subs["pk-2"].DaysRemaining()

	if after <= initial {
		t.Fatalf("subscription should extend: was %d days, now %d", initial, after)
	}
	if m.Subs["pk-2"].Tier != TierAristocrat {
		t.Fatalf("expected aristocrat, got %s", m.Subs["pk-2"].Tier)
	}
}


// TestGetTierDefaultsToColonist handles the TestGetTierDefaultsToColonist HTTP request.
func TestGetTierDefaultsToColonist(t *testing.T) {
	m := NewSubscriptionManager()
	tier := m.GetTier("unknown-pubkey")
	if tier != TierColonist {
		t.Fatalf("expected colonist for unknown, got %s", tier)
	}
}


// TestGetTierReturnsActive handles the TestGetTierReturnsActive HTTP request.
func TestGetTierReturnsActive(t *testing.T) {
	m := NewSubscriptionManager()
	m.Subscribe("pk-3", TierCitizen, 30)
	tier := m.GetTier("pk-3")
	if tier != TierCitizen {
		t.Fatalf("expected citizen, got %s", tier)
	}
}


// TestTierString handles the TestTierString HTTP request.
func TestTierString(t *testing.T) {
	if string(TierColonist) != "colonist" {
		t.Fatalf("unexpected colonist string")
	}
}


// TestBenefitsVaultQuota handles the TestBenefitsVaultQuota HTTP request.
func TestBenefitsVaultQuota(t *testing.T) {
	tests := []struct {
		tier  Tier
		quota int
	}{
		{TierColonist, 512},
		{TierCitizen, 2048},
		{TierAristocrat, 8192},
	}
	for _, tc := range tests {
		b := tc.tier.Benefits()
		if b.VaultQuotaMB != tc.quota {
			t.Errorf("%s vault quota: got %d, want %d", tc.tier, b.VaultQuotaMB, tc.quota)
		}
	}
}
