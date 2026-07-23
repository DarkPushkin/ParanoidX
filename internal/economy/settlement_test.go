// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestSettlementManagerNew handles the TestSettlementManagerNew HTTP request.
func TestSettlementManagerNew(t *testing.T) {
	sm := NewSettlementManager()
	if len(sm.Settlements) != 0 {
		t.Fatal("expected empty")
	}
}


// TestSettlementCreate handles the TestSettlementCreate HTTP request.
func TestSettlementCreate(t *testing.T) {
	sm := NewSettlementManager()
	s, err := sm.Create("lic-from", "lic-to", "monthly settlement", 1000*NGPerTLR)
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != "pending" {
		t.Fatalf("expected pending, got %s", s.Status)
	}
}


// TestSettlementCreateNegative handles the TestSettlementCreateNegative HTTP request.
func TestSettlementCreateNegative(t *testing.T) {
	sm := NewSettlementManager()
	_, err := sm.Create("from", "to", "test", -100)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}


// TestSettlementComplete handles the TestSettlementComplete HTTP request.
func TestSettlementComplete(t *testing.T) {
	sm := NewSettlementManager()
	s, _ := sm.Create("f1", "f2", "payment", 500*NGPerTLR)
	completed, err := sm.Complete(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != "completed" {
		t.Fatalf("expected completed, got %s", completed.Status)
	}
}


// TestSettlementCompleteNotFound handles the TestSettlementCompleteNotFound HTTP request.
func TestSettlementCompleteNotFound(t *testing.T) {
	sm := NewSettlementManager()
	_, err := sm.Complete("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}


// TestSettlementCompleteTwice handles the TestSettlementCompleteTwice HTTP request.
func TestSettlementCompleteTwice(t *testing.T) {
	sm := NewSettlementManager()
	s, _ := sm.Create("f1", "f2", "test", 100)
	sm.Complete(s.ID)
	_, err := sm.Complete(s.ID)
	if err == nil {
		t.Fatal("expected error for double complete")
	}
}


// TestSettlementList handles the TestSettlementList HTTP request.
func TestSettlementList(t *testing.T) {
	sm := NewSettlementManager()
	sm.Create("a", "b", "1", 100)
	sm.Create("c", "d", "2", 200)
	list := sm.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}


// TestSettlementSaveAndLoad handles the TestSettlementSaveAndLoad HTTP request.
func TestSettlementSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	sm := NewSettlementManager()
	sm.Create("f1", "f2", "test", 1000)
	sm.Save(dir)

	loaded := LoadSettlements(dir)
	if len(loaded.Settlements) != 1 {
		t.Fatalf("expected 1 settlement, got %d", len(loaded.Settlements))
	}
}


// TestRoyaltyManagerNew handles the TestRoyaltyManagerNew HTTP request.
func TestRoyaltyManagerNew(t *testing.T) {
	rm := NewRoyaltyManager()
	if len(rm.Payments) != 0 {
		t.Fatal("expected empty")
	}
}


// TestRoyaltyCreateDue handles the TestRoyaltyCreateDue HTTP request.
func TestRoyaltyCreateDue(t *testing.T) {
	rm := NewRoyaltyManager()
	p := rm.CreateDue("lic1", "2026-06", 10*NGPerTLR)
	if p.Paid {
		t.Fatal("should not be paid initially")
	}
}


// TestRoyaltyPay handles the TestRoyaltyPay HTTP request.
func TestRoyaltyPay(t *testing.T) {
	rm := NewRoyaltyManager()
	p := rm.CreateDue("lic1", "2026-06", 10*NGPerTLR)
	paid, err := rm.Pay(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !paid.Paid {
		t.Fatal("should be paid")
	}
}


// TestRoyaltyPayTwice handles the TestRoyaltyPayTwice HTTP request.
func TestRoyaltyPayTwice(t *testing.T) {
	rm := NewRoyaltyManager()
	p := rm.CreateDue("lic1", "2026-06", 10*NGPerTLR)
	rm.Pay(p.ID)
	_, err := rm.Pay(p.ID)
	if err == nil {
		t.Fatal("expected error for double pay")
	}
}


// TestRoyaltyPayNotFound handles the TestRoyaltyPayNotFound HTTP request.
func TestRoyaltyPayNotFound(t *testing.T) {
	rm := NewRoyaltyManager()
	_, err := rm.Pay("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}


// TestRoyaltyPending handles the TestRoyaltyPending HTTP request.
func TestRoyaltyPending(t *testing.T) {
	rm := NewRoyaltyManager()
	rm.CreateDue("l1", "2026-06", 10*NGPerTLR)
	p2 := rm.CreateDue("l2", "2026-06", 20*NGPerTLR)
	rm.Pay(p2.ID)
	pending := rm.Pending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}
}


// TestRoyaltyList handles the TestRoyaltyList HTTP request.
func TestRoyaltyList(t *testing.T) {
	rm := NewRoyaltyManager()
	rm.CreateDue("l1", "2026-06", 10*NGPerTLR)
	rm.CreateDue("l2", "2026-07", 20*NGPerTLR)
	list := rm.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}


// TestRoyaltySaveAndLoad handles the TestRoyaltySaveAndLoad HTTP request.
func TestRoyaltySaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	rm := NewRoyaltyManager()
	rm.CreateDue("lic1", "2026-06", 10*NGPerTLR)
	rm.Save(dir)

	loaded := LoadRoyalties(dir)
	if len(loaded.Payments) != 1 {
		t.Fatalf("expected 1, got %d", len(loaded.Payments))
	}
}
