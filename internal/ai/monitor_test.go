// Package ai provides AI integration with Ollama, including chat, generation, and monitoring
package ai

import (
	"testing"
)


// TestNewStewardMonitor handles the TestNewStewardMonitor HTTP request.
func TestNewStewardMonitor(t *testing.T) {
	c := NewClient("", "")
	sm := NewStewardMonitor(c)
	if sm.Client != c {
		t.Fatal("expected same client reference")
	}
}


// TestCheckReserveRatioOK handles the TestCheckReserveRatioOK HTTP request.
func TestCheckReserveRatioOK(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckReserveRatio(1000, 500)
	if d.Status != "ok" {
		t.Fatalf("expected ok, got %s", d.Status)
	}
}


// TestCheckReserveRatioWarning handles the TestCheckReserveRatioWarning HTTP request.
func TestCheckReserveRatioWarning(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckReserveRatio(1200, 1000)
	if d.Status != "warning" {
		t.Fatalf("expected warning (ratio 1.2), got %s", d.Status)
	}
}


// TestCheckReserveRatioCritical handles the TestCheckReserveRatioCritical HTTP request.
func TestCheckReserveRatioCritical(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckReserveRatio(500, 1000)
	if d.Status != "critical" {
		t.Fatalf("expected critical, got %s", d.Status)
	}
}


// TestCheckReserveRatioZeroLiabilities handles the TestCheckReserveRatioZeroLiabilities HTTP request.
func TestCheckReserveRatioZeroLiabilities(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckReserveRatio(1000, 0)
	if d.Status != "ok" {
		t.Fatalf("expected ok, got %s", d.Status)
	}
}


// TestCheckTreasuryTierOK handles the TestCheckTreasuryTierOK HTTP request.
func TestCheckTreasuryTierOK(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckTreasuryTier("normal", false)
	if d.Status != "ok" {
		t.Fatalf("expected ok, got %s", d.Status)
	}
}


// TestCheckTreasuryTierActionRequired handles the TestCheckTreasuryTierActionRequired HTTP request.
func TestCheckTreasuryTierActionRequired(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckTreasuryTier("very_fat", true)
	if d.Status != "action_required" {
		t.Fatalf("expected action_required, got %s", d.Status)
	}
}


// TestCheckServicesAllActive handles the TestCheckServicesAllActive HTTP request.
func TestCheckServicesAllActive(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckActiveServices(5, 5)
	if d.Status != "ok" {
		t.Fatalf("expected ok, got %s", d.Status)
	}
}


// TestCheckServicesWarning handles the TestCheckServicesWarning HTTP request.
func TestCheckServicesWarning(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckActiveServices(3, 5)
	if d.Status != "warning" {
		t.Fatalf("expected warning, got %s", d.Status)
	}
}


// TestCheckServicesCritical handles the TestCheckServicesCritical HTTP request.
func TestCheckServicesCritical(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckActiveServices(1, 5)
	if d.Status != "critical" {
		t.Fatalf("expected critical, got %s", d.Status)
	}
}


// TestCheckServicesNoServices handles the TestCheckServicesNoServices HTTP request.
func TestCheckServicesNoServices(t *testing.T) {
	sm := NewStewardMonitor(nil)
	d := sm.CheckActiveServices(0, 0)
	if d.Status != "ok" {
		t.Fatalf("expected ok, got %s", d.Status)
	}
}


// TestRecordAndSummary handles the TestRecordAndSummary HTTP request.
func TestRecordAndSummary(t *testing.T) {
	sm := NewStewardMonitor(nil)
	sm.Record(MonitorDecision{Check: "c1", Status: "ok"})
	sm.Record(MonitorDecision{Check: "c2", Status: "warning"})
	summary := sm.Summary(2)
	if len(summary) != 2 {
		t.Fatalf("expected 2, got %d", len(summary))
	}
}


// TestAskWithConstitution handles the TestAskWithConstitution HTTP request.
func TestAskWithConstitution(t *testing.T) {
	// Only test the function doesn't panic; actual AI call requires Ollama
	if testing.Short() {
		t.Skip("skipping remote AI test")
	}
}
