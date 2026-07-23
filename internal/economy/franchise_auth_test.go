// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestMintAuthManagerNew handles the TestMintAuthManagerNew HTTP request.
func TestMintAuthManagerNew(t *testing.T) {
	m := NewMintAuthManager()
	if len(m.Auths) != 0 {
		t.Fatal("expected empty")
	}
}


// TestMintAuthRequest handles the TestMintAuthRequest HTTP request.
func TestMintAuthRequest(t *testing.T) {
	m := NewMintAuthManager()
	auth := m.Request("lic1", "MB-COMMON-100", "common", NGPerTLR)
	if auth.LicenseID != "lic1" {
		t.Fatalf("expected lic1, got %s", auth.LicenseID)
	}
	if auth.Approved {
		t.Fatal("should not be approved initially")
	}
}


// TestMintAuthApprove handles the TestMintAuthApprove HTTP request.
func TestMintAuthApprove(t *testing.T) {
	m := NewMintAuthManager()
	auth := m.Request("lic2", "MB-RARE-050", "rare", 5*NGPerTLR)
	approved, err := m.Approve(auth.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !approved.Approved {
		t.Fatal("should be approved")
	}
}


// TestMintAuthApproveTwice handles the TestMintAuthApproveTwice HTTP request.
func TestMintAuthApproveTwice(t *testing.T) {
	m := NewMintAuthManager()
	auth := m.Request("lic3", "MB-EPIC-010", "epic", 25*NGPerTLR)
	m.Approve(auth.ID)
	_, err := m.Approve(auth.ID)
	if err == nil {
		t.Fatal("expected error for double approve")
	}
}


// TestMintAuthApproveNotFound handles the TestMintAuthApproveNotFound HTTP request.
func TestMintAuthApproveNotFound(t *testing.T) {
	m := NewMintAuthManager()
	_, err := m.Approve("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}


// TestMintAuthPending handles the TestMintAuthPending HTTP request.
func TestMintAuthPending(t *testing.T) {
	m := NewMintAuthManager()
	m.Request("l1", "s1", "common", NGPerTLR)
	m.Request("l2", "s2", "rare", 5*NGPerTLR)
	auths := m.Pending()
	if len(auths) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(auths))
	}
}


// TestMintAuthPendingAfterApproval handles the TestMintAuthPendingAfterApproval HTTP request.
func TestMintAuthPendingAfterApproval(t *testing.T) {
	m := NewMintAuthManager()
	a := m.Request("l1", "s1", "common", NGPerTLR)
	m.Request("l2", "s2", "rare", 5*NGPerTLR)
	m.Approve(a.ID)
	pending := m.Pending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending after approval, got %d", len(pending))
	}
}


// TestMintAuthList handles the TestMintAuthList HTTP request.
func TestMintAuthList(t *testing.T) {
	m := NewMintAuthManager()
	m.Request("l1", "s1", "common", NGPerTLR)
	m.Request("l2", "s2", "rare", 5*NGPerTLR)
	list := m.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}


// TestMintAuthSaveAndLoad handles the TestMintAuthSaveAndLoad HTTP request.
func TestMintAuthSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	m := NewMintAuthManager()
	a := m.Request("l1", "s1", "common", NGPerTLR)
	m.Approve(a.ID)
	m.Save(dir)

	loaded := LoadMintAuths(dir)
	if len(loaded.Auths) != 1 {
		t.Fatalf("expected 1 auth, got %d", len(loaded.Auths))
	}
	if !loaded.Auths[0].Approved {
		t.Fatal("should be approved after load")
	}
}


// TestTemplateManagerNew handles the TestTemplateManagerNew HTTP request.
func TestTemplateManagerNew(t *testing.T) {
	tm := NewTemplateManager()
	if len(tm.Templates) != 0 {
		t.Fatal("expected empty")
	}
}


// TestTemplateCreate handles the TestTemplateCreate HTTP request.
func TestTemplateCreate(t *testing.T) {
	tm := NewTemplateManager()
	tpl, err := tm.Create("tmpl1", "lic1", "National Standard", `{"color":"blue"}`)
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "National Standard" {
		t.Fatalf("expected 'National Standard', got %s", tpl.Name)
	}
}


// TestTemplateCreateDuplicate handles the TestTemplateCreateDuplicate HTTP request.
func TestTemplateCreateDuplicate(t *testing.T) {
	tm := NewTemplateManager()
	tm.Create("tmpl2", "lic1", "First", `{}`)
	_, err := tm.Create("tmpl2", "lic1", "Second", `{}`)
	if err == nil {
		t.Fatal("expected error for duplicate")
	}
}


// TestTemplateGet handles the TestTemplateGet HTTP request.
func TestTemplateGet(t *testing.T) {
	tm := NewTemplateManager()
	tm.Create("t1", "l1", "T1", `{"theme":"dark"}`)
	tpl, err := tm.Get("t1")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.DesignJSON != `{"theme":"dark"}` {
		t.Fatalf("expected design json, got %s", tpl.DesignJSON)
	}
}


// TestTemplateGetNotFound handles the TestTemplateGetNotFound HTTP request.
func TestTemplateGetNotFound(t *testing.T) {
	tm := NewTemplateManager()
	_, err := tm.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}


// TestTemplateList handles the TestTemplateList HTTP request.
func TestTemplateList(t *testing.T) {
	tm := NewTemplateManager()
	tm.Create("a", "l1", "A", `{}`)
	tm.Create("b", "l2", "B", `{}`)
	list := tm.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}


// TestTemplateDelete handles the TestTemplateDelete HTTP request.
func TestTemplateDelete(t *testing.T) {
	tm := NewTemplateManager()
	tm.Create("del1", "l1", "Delete Me", `{}`)
	err := tm.Delete("del1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tm.Get("del1")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}


// TestTemplateSaveAndLoad handles the TestTemplateSaveAndLoad HTTP request.
func TestTemplateSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	tm := NewTemplateManager()
	tm.Create("st1", "l1", "Save Test", `{"color":"gold"}`)
	tm.Save(dir)

	loaded := LoadTemplates(dir)
	tpl, err := loaded.Get("st1")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "Save Test" {
		t.Fatalf("expected 'Save Test', got %s", tpl.Name)
	}
}
