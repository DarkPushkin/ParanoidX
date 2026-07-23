// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestLicenseManagerNew handles the TestLicenseManagerNew HTTP request.
func TestLicenseManagerNew(t *testing.T) {
	lm := NewLicenseManager()
	if len(lm.Licenses) != 0 {
		t.Fatal("expected empty")
	}
}


// TestLicenseCreate handles the TestLicenseCreate HTTP request.
func TestLicenseCreate(t *testing.T) {
	lm := NewLicenseManager()
	lic, err := lm.Create("f1", "Franchise One", "standard", 50)
	if err != nil {
		t.Fatal(err)
	}
	if lic.ID != "f1" {
		t.Fatalf("expected f1, got %s", lic.ID)
	}
	if lic.FeeNg != 1_000_000_000 {
		t.Fatalf("expected fee 1B ng, got %d", lic.FeeNg)
	}
	if lic.MaxNodes != 1 {
		t.Fatalf("expected standard max 1 node, got %d", lic.MaxNodes)
	}
}


// TestLicenseCreateDuplicate handles the TestLicenseCreateDuplicate HTTP request.
func TestLicenseCreateDuplicate(t *testing.T) {
	lm := NewLicenseManager()
	lm.Create("f1", "F1", "standard", 50)
	_, err := lm.Create("f1", "F1 dup", "standard", 50)
	if err == nil {
		t.Fatal("expected error for duplicate")
	}
}


// TestLicenseCreateInvalidTier handles the TestLicenseCreateInvalidTier HTTP request.
func TestLicenseCreateInvalidTier(t *testing.T) {
	lm := NewLicenseManager()
	_, err := lm.Create("f2", "F2", "nonexistent", 50)
	if err == nil {
		t.Fatal("expected error for invalid tier")
	}
}


// TestLicenseGet handles the TestLicenseGet HTTP request.
func TestLicenseGet(t *testing.T) {
	lm := NewLicenseManager()
	lm.Create("f3", "F3", "premium", 100)
	lic, err := lm.Get("f3")
	if err != nil {
		t.Fatal(err)
	}
	if lic.Tier != "premium" {
		t.Fatalf("expected premium, got %s", lic.Tier)
	}
}


// TestLicenseGetNotFound handles the TestLicenseGetNotFound HTTP request.
func TestLicenseGetNotFound(t *testing.T) {
	lm := NewLicenseManager()
	_, err := lm.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}


// TestLicenseUpdate handles the TestLicenseUpdate HTTP request.
func TestLicenseUpdate(t *testing.T) {
	lm := NewLicenseManager()
	lm.Create("f4", "F4", "standard", 25)
	lic, err := lm.Update("f4", "F4 Updated", "xyz.onion", LicenseActive)
	if err != nil {
		t.Fatal(err)
	}
	if lic.Name != "F4 Updated" {
		t.Fatalf("expected updated name, got %s", lic.Name)
	}
	if lic.OnionAddr != "xyz.onion" {
		t.Fatalf("expected xyz.onion, got %s", lic.OnionAddr)
	}
}


// TestLicenseUpdateSuspended handles the TestLicenseUpdateSuspended HTTP request.
func TestLicenseUpdateSuspended(t *testing.T) {
	lm := NewLicenseManager()
	lm.Create("f5", "F5", "standard", 30)
	lic, err := lm.Update("f5", "", "", LicenseSuspended)
	if err != nil {
		t.Fatal(err)
	}
	if lic.Status != LicenseSuspended {
		t.Fatalf("expected suspended, got %s", lic.Status)
	}
}


// TestLicenseDelete handles the TestLicenseDelete HTTP request.
func TestLicenseDelete(t *testing.T) {
	lm := NewLicenseManager()
	lm.Create("f6", "F6", "standard", 40)
	err := lm.Delete("f6")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lm.Get("f6")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}


// TestLicenseList handles the TestLicenseList HTTP request.
func TestLicenseList(t *testing.T) {
	lm := NewLicenseManager()
	lm.Create("a", "A", "standard", 10)
	lm.Create("b", "B", "premium", 20)
	list := lm.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}


// TestLicenseInvalidRoyalty handles the TestLicenseInvalidRoyalty HTTP request.
func TestLicenseInvalidRoyalty(t *testing.T) {
	lm := NewLicenseManager()
	_, err := lm.Create("f7", "F7", "standard", -1)
	if err == nil {
		t.Fatal("expected error for negative royalty")
	}
	_, err = lm.Create("f8", "F8", "standard", 10001)
	if err == nil {
		t.Fatal("expected error for royalty > 10000")
	}
}


// TestPremiumTierConfig handles the TestPremiumTierConfig HTTP request.
func TestPremiumTierConfig(t *testing.T) {
	lm := NewLicenseManager()
	lic, err := lm.Create("p1", "Premium", "premium", 100)
	if err != nil {
		t.Fatal(err)
	}
	if lic.MaxNodes != 5 {
		t.Fatalf("expected premium max 5 nodes, got %d", lic.MaxNodes)
	}
	if lic.FeeNg != 5_000_000_000 {
		t.Fatalf("expected 5B ng fee, got %d", lic.FeeNg)
	}
}


// TestRoyalTierConfig handles the TestRoyalTierConfig HTTP request.
func TestRoyalTierConfig(t *testing.T) {
	lm := NewLicenseManager()
	lic, err := lm.Create("r1", "Royal", "royal", 500)
	if err != nil {
		t.Fatal(err)
	}
	if lic.MaxNodes != 100 {
		t.Fatalf("expected royal max 100 nodes, got %d", lic.MaxNodes)
	}
	if lic.FeeNg != 25_000_000_000 {
		t.Fatalf("expected 25B ng fee, got %d", lic.FeeNg)
	}
}


// TestSaveAndLoadLicenses handles the TestSaveAndLoadLicenses HTTP request.
func TestSaveAndLoadLicenses(t *testing.T) {
	dir := t.TempDir()
	lm := NewLicenseManager()
	lm.Create("s1", "Save Test", "standard", 25)
	lm.Save(dir)

	loaded := LoadLicenses(dir)
	lic, err := loaded.Get("s1")
	if err != nil {
		t.Fatal(err)
	}
	if lic.Name != "Save Test" {
		t.Fatalf("expected 'Save Test', got %s", lic.Name)
	}
}


// TestEarmarkManagerNew handles the TestEarmarkManagerNew HTTP request.
func TestEarmarkManagerNew(t *testing.T) {
	em := NewEarmarkManager()
	if len(em.Accounts) != 0 {
		t.Fatal("expected empty")
	}
}


// TestEarmarkCreate handles the TestEarmarkCreate HTTP request.
func TestEarmarkCreate(t *testing.T) {
	dir := t.TempDir()
	em := NewEarmarkManager()
	acct, err := em.Create(dir, "ea1", "pk1", "ops", "lic1", 100*NGPerTLR)
	if err != nil {
		t.Fatal(err)
	}
	if acct.Purpose != "ops" {
		t.Fatalf("expected ops, got %s", acct.Purpose)
	}
}


// TestEarmarkSpend handles the TestEarmarkSpend HTTP request.
func TestEarmarkSpend(t *testing.T) {
	dir := t.TempDir()
	em := NewEarmarkManager()
	em.Create(dir, "ea2", "pk2", "reserve", "lic2", 1000*NGPerTLR)
	acct, err := em.Spend("ea2", 300*NGPerTLR)
	if err != nil {
		t.Fatal(err)
	}
	if acct.SpentNg != 300*NGPerTLR {
		t.Fatalf("expected 300 TLR spent, got %d", acct.SpentNg)
	}
}


// TestEarmarkSpendInsufficient handles the TestEarmarkSpendInsufficient HTTP request.
func TestEarmarkSpendInsufficient(t *testing.T) {
	dir := t.TempDir()
	em := NewEarmarkManager()
	em.Create(dir, "ea3", "pk3", "dividends", "lic3", 100*NGPerTLR)
	_, err := em.Spend("ea3", 200*NGPerTLR)
	if err == nil {
		t.Fatal("expected error for overspend")
	}
}


// TestEarmarkRemaining handles the TestEarmarkRemaining HTTP request.
func TestEarmarkRemaining(t *testing.T) {
	em := NewEarmarkManager()
	// Can't test with data dir — just check Remaining for nonexistent
	r := em.Remaining("nonexistent")
	if r != 0 {
		t.Fatalf("expected 0, got %d", r)
	}
}


// TestEarmarkList handles the TestEarmarkList HTTP request.
func TestEarmarkList(t *testing.T) {
	dir := t.TempDir()
	em := NewEarmarkManager()
	em.Create(dir, "l1", "pk", "ops", "l", 100)
	em.Create(dir, "l2", "pk", "reserve", "l", 200)
	list := em.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}


// TestSaveAndLoadEarmarks handles the TestSaveAndLoadEarmarks HTTP request.
func TestSaveAndLoadEarmarks(t *testing.T) {
	dir := t.TempDir()
	em := NewEarmarkManager()
	em.Create(dir, "se1", "pk", "franchise_dev", "sl", 500*NGPerTLR)
	em.Save(dir)

	loaded := LoadEarmarks(dir)
	if loaded.Accounts["se1"] == nil {
		t.Fatal("expected to find se1 after load")
	}
}
