// Package ai provides AI integration with Ollama, including chat, generation, and monitoring
package ai

import (
	"testing"
)


// TestNewClientDefaults handles the TestNewClientDefaults HTTP request.
func TestNewClientDefaults(t *testing.T) {
	c := NewClient("", "")
	if c.BaseURL != "http://localhost:11434" {
		t.Fatalf("expected default localhost, got %s", c.BaseURL)
	}
	if c.Model != "minimax-m3:cloud" {
		t.Fatalf("expected minimax-m3:cloud, got %s", c.Model)
	}
}


// TestNewClientCustom handles the TestNewClientCustom HTTP request.
func TestNewClientCustom(t *testing.T) {
	c := NewClient("http://192.168.1.129:11434", "qwen2.5:7b")
	if c.BaseURL != "http://192.168.1.129:11434" {
		t.Fatalf("expected custom URL, got %s", c.BaseURL)
	}
	if c.Model != "qwen2.5:7b" {
		t.Fatalf("expected qwen2.5:7b, got %s", c.Model)
	}
}


// TestNewSteward handles the TestNewSteward HTTP request.
func TestNewSteward(t *testing.T) {
	c := NewClient("", "")
	s := NewSteward(c)
	if s.Client != c {
		t.Fatal("expected steward to hold client reference")
	}
}


// TestIsAvailable handles the TestIsAvailable HTTP request.
func TestIsAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping remote Ollama test")
	}
	c := NewClient("", "")
	if !c.IsAvailable() {
		t.Skip("Ollama not available at localhost:11434")
	}
}


// TestEconomySummary handles the TestEconomySummary HTTP request.
func TestEconomySummary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping remote Ollama test")
	}
	s := NewSteward(NewClient("", ""))
	if !s.Client.IsAvailable() {
		t.Skip("Ollama not available — skipping integration test")
	}
	summary, err := s.EconomySummary(`{"reserve_ng":1000000000000,"supply_ng":500000000000,"holders":42}`)
	if err != nil {
		t.Fatal(err)
	}
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
	t.Logf("Summary: %s", summary)
}


// TestModerationCheck handles the TestModerationCheck HTTP request.
func TestModerationCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping remote Ollama test")
	}
	s := NewSteward(NewClient("", ""))
	if !s.Client.IsAvailable() {
		t.Skip("Ollama not available — skipping integration test")
	}
	safe, reason, err := s.ModerationCheck("Hello, how is the weather today?")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Safe: %v, Reason: %s", safe, reason)
	if !safe {
		t.Log("note: flagged as unsafe (may be conservative)")
	}
}
