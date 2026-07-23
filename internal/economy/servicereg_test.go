// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestServiceRegistryNew handles the TestServiceRegistryNew HTTP request.
func TestServiceRegistryNew(t *testing.T) {
	sr := NewServiceRegistry()
	if len(sr.Services) != 0 {
		t.Fatal("expected empty")
	}
}


// TestServiceRegister handles the TestServiceRegister HTTP request.
func TestServiceRegister(t *testing.T) {
	sr := NewServiceRegistry()
	svc := &RegisteredService{
		ID:          "svc1",
		LicenseID:   "lic1",
		ServiceType: SvcStorage,
		Name:        "Vault Service",
		Endpoint:    "http://node1.onion/vault",
		PricePerUse: 1000,
	}
	err := sr.Register(svc)
	if err != nil {
		t.Fatal(err)
	}
	if svc.Status != "active" {
		t.Fatalf("expected active, got %s", svc.Status)
	}
}


// TestServiceRegisterEmptyID handles the TestServiceRegisterEmptyID HTTP request.
func TestServiceRegisterEmptyID(t *testing.T) {
	sr := NewServiceRegistry()
	err := sr.Register(&RegisteredService{})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}


// TestServiceGet handles the TestServiceGet HTTP request.
func TestServiceGet(t *testing.T) {
	sr := NewServiceRegistry()
	sr.Register(&RegisteredService{ID: "svc2", LicenseID: "l1", Name: "Svc2"})
	svc, err := sr.Get("svc2")
	if err != nil {
		t.Fatal(err)
	}
	if svc.Name != "Svc2" {
		t.Fatalf("expected Svc2, got %s", svc.Name)
	}
}


// TestServiceGetNotFound handles the TestServiceGetNotFound HTTP request.
func TestServiceGetNotFound(t *testing.T) {
	sr := NewServiceRegistry()
	_, err := sr.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}


// TestServiceListByType handles the TestServiceListByType HTTP request.
func TestServiceListByType(t *testing.T) {
	sr := NewServiceRegistry()
	sr.Register(&RegisteredService{ID: "s1", LicenseID: "l1", ServiceType: SvcStorage, Name: "S1", Status: "active"})
	sr.Register(&RegisteredService{ID: "s2", LicenseID: "l1", ServiceType: SvcRadio, Name: "R1", Status: "active"})
	sr.Register(&RegisteredService{ID: "s3", LicenseID: "l2", ServiceType: SvcStorage, Name: "S2", Status: "active"})

	storages := sr.ListByType(SvcStorage)
	if len(storages) != 2 {
		t.Fatalf("expected 2 storage services, got %d", len(storages))
	}
}


// TestServiceListByTypeExcludesInactive handles the TestServiceListByTypeExcludesInactive HTTP request.
func TestServiceListByTypeExcludesInactive(t *testing.T) {
	sr := NewServiceRegistry()
	sr.Register(&RegisteredService{ID: "s1", LicenseID: "l1", ServiceType: SvcStorage, Name: "S1", Status: "offline"})
	sr.Register(&RegisteredService{ID: "s2", LicenseID: "l2", ServiceType: SvcStorage, Name: "S2", Status: "active"})

	storages := sr.ListByType(SvcStorage)
	if len(storages) != 1 {
		t.Fatalf("expected 1 active storage, got %d", len(storages))
	}
}


// TestServiceListAll handles the TestServiceListAll HTTP request.
func TestServiceListAll(t *testing.T) {
	sr := NewServiceRegistry()
	sr.Register(&RegisteredService{ID: "a", LicenseID: "l1", Name: "A"})
	sr.Register(&RegisteredService{ID: "b", LicenseID: "l2", Name: "B"})
	all := sr.ListAll()
	if len(all) != 2 {
		t.Fatalf("expected 2, got %d", len(all))
	}
}


// TestServiceDeregister handles the TestServiceDeregister HTTP request.
func TestServiceDeregister(t *testing.T) {
	sr := NewServiceRegistry()
	sr.Register(&RegisteredService{ID: "svc-del", LicenseID: "l1", Name: "Del Me"})
	err := sr.Deregister("svc-del")
	if err != nil {
		t.Fatal(err)
	}
	_, err = sr.Get("svc-del")
	if err == nil {
		t.Fatal("expected error after deregister")
	}
}


// TestServiceSaveAndLoad handles the TestServiceSaveAndLoad HTTP request.
func TestServiceSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	sr := NewServiceRegistry()
	sr.Register(&RegisteredService{ID: "svc-save", LicenseID: "l1", Name: "Save Test"})
	sr.Save(dir)

	loaded := LoadServiceRegistry(dir)
	svc, err := loaded.Get("svc-save")
	if err != nil {
		t.Fatal(err)
	}
	if svc.Name != "Save Test" {
		t.Fatalf("expected 'Save Test', got %s", svc.Name)
	}
}


// TestMarketplaceNew handles the TestMarketplaceNew HTTP request.
func TestMarketplaceNew(t *testing.T) {
	sm := NewServiceMarketplace()
	if len(sm.Listings) != 0 {
		t.Fatal("expected empty")
	}
}


// TestMarketplaceList handles the TestMarketplaceList HTTP request.
func TestMarketplaceList(t *testing.T) {
	sm := NewServiceMarketplace()
	listing, err := sm.List("svc1", "seller1")
	if err != nil {
		t.Fatal(err)
	}
	if !listing.Available {
		t.Fatal("should be available")
	}
}


// TestMarketplaceListDuplicate handles the TestMarketplaceListDuplicate HTTP request.
func TestMarketplaceListDuplicate(t *testing.T) {
	sm := NewServiceMarketplace()
	sm.List("svc1", "seller1")
	_, err := sm.List("svc1", "seller1")
	if err == nil {
		t.Fatal("expected error for duplicate")
	}
}


// TestMarketplaceBuy handles the TestMarketplaceBuy HTTP request.
func TestMarketplaceBuy(t *testing.T) {
	sm := NewServiceMarketplace()
	listing, _ := sm.List("svc2", "seller2")
	bought, err := sm.Buy(listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bought.Available {
		t.Fatal("should be unavailable after buy")
	}
}


// TestMarketplaceBuyNotFound handles the TestMarketplaceBuyNotFound HTTP request.
func TestMarketplaceBuyNotFound(t *testing.T) {
	sm := NewServiceMarketplace()
	_, err := sm.Buy("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}


// TestMarketplaceRemove handles the TestMarketplaceRemove HTTP request.
func TestMarketplaceRemove(t *testing.T) {
	sm := NewServiceMarketplace()
	listing, _ := sm.List("svc3", "seller3")
	err := sm.Remove(listing.ID)
	if err != nil {
		t.Fatal(err)
	}
}


// TestMarketplaceSearch handles the TestMarketplaceSearch HTTP request.
func TestMarketplaceSearch(t *testing.T) {
	sm := NewServiceMarketplace()
	sm.List("s1", "seller1")
	sm.List("s2", "seller2")
	results := sm.Search(SvcStorage)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}


// TestMarketplaceSaveAndLoad handles the TestMarketplaceSaveAndLoad HTTP request.
func TestMarketplaceSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	sm := NewServiceMarketplace()
	sm.List("svc-save", "seller-save")
	sm.Save(dir)

	loaded := LoadMarketplace(dir)
	results := loaded.Search(SvcStorage)
	if len(results) != 1 {
		t.Fatalf("expected 1 listing, got %d", len(results))
	}
}
