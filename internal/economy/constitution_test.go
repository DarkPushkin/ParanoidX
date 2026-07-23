// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestDefaultConstitution handles the TestDefaultConstitution HTTP request.
func TestDefaultConstitution(t *testing.T) {
	c := DefaultConstitution()
	if c.Version != 1 {
		t.Fatalf("expected version 1, got %d", c.Version)
	}
	if len(c.Articles) != 10 {
		t.Fatalf("expected 10 articles, got %d", len(c.Articles))
	}
}


// TestGetArticle handles the TestGetArticle HTTP request.
func TestGetArticle(t *testing.T) {
	c := DefaultConstitution()
	a := c.GetArticle(1)
	if a == nil {
		t.Fatal("expected article 1")
	}
	if a.Title != "Silver Backing" {
		t.Fatalf("expected 'Silver Backing', got %s", a.Title)
	}
}


// TestGetArticleNotFound handles the TestGetArticleNotFound HTTP request.
func TestGetArticleNotFound(t *testing.T) {
	c := DefaultConstitution()
	a := c.GetArticle(99)
	if a != nil {
		t.Fatal("expected nil for nonexistent article")
	}
}


// TestConstitutionSaveAndLoad handles the TestConstitutionSaveAndLoad HTTP request.
func TestConstitutionSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	c := DefaultConstitution()
	c.Version = 2
	c.Save(dir)

	loaded := LoadConstitution(dir)
	if loaded.Version != 2 {
		t.Fatalf("expected version 2, got %d", loaded.Version)
	}
}


// TestDecisionLogNew handles the TestDecisionLogNew HTTP request.
func TestDecisionLogNew(t *testing.T) {
	d := NewDecisionLog()
	if len(d.Decisions) != 0 {
		t.Fatal("expected empty")
	}
}


// TestDecisionLogRecord handles the TestDecisionLogRecord HTTP request.
func TestDecisionLogRecord(t *testing.T) {
	d := NewDecisionLog()
	d.Record(StewardDecision{
		ID: "dec1", Type: "moderation", Decision: "approve", Timestamp: "now",
	})
	if len(d.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(d.Decisions))
	}
}


// TestDecisionLogRecent handles the TestDecisionLogRecent HTTP request.
func TestDecisionLogRecent(t *testing.T) {
	d := NewDecisionLog()
	for i := 0; i < 10; i++ {
		d.Record(StewardDecision{ID: string(rune('0' + i))})
	}
	recent := d.Recent(3)
	if len(recent) != 3 {
		t.Fatalf("expected 3 recent, got %d", len(recent))
	}
}


// TestDecisionLogSaveAndLoad handles the TestDecisionLogSaveAndLoad HTTP request.
func TestDecisionLogSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	d := NewDecisionLog()
	d.Record(StewardDecision{ID: "save-test", Decision: "approve"})
	d.Save(dir)

	loaded := LoadDecisionLog(dir)
	if len(loaded.Decisions) != 1 {
		t.Fatalf("expected 1, got %d", len(loaded.Decisions))
	}
	if loaded.Decisions[0].ID != "save-test" {
		t.Fatalf("expected 'save-test', got %s", loaded.Decisions[0].ID)
	}
}


// TestConstitutionArticles handles the TestConstitutionArticles HTTP request.
func TestConstitutionArticles(t *testing.T) {
	c := DefaultConstitution()
	titles := []string{
		"Silver Backing", "Treasury Split", "Subscription Tiers",
		"Golden Wheel", "Crafting", "Auto-Mint", "Deflation",
		"Franchise Rights", "Auditor Governance", "Amendment",
	}
	for i, title := range titles {
		a := c.GetArticle(i + 1)
		if a == nil {
			t.Fatalf("article %d not found", i+1)
		}
		if a.Title != title {
			t.Fatalf("article %d: expected %q, got %q", i+1, title, a.Title)
		}
	}
}
