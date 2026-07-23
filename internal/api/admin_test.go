// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)


// TestConfigHandlerGetSet handles the TestConfigHandlerGetSet HTTP request.
func TestConfigHandlerGetSet(t *testing.T) {
	dir := t.TempDir()
	handler := ConfigHandler(dir)

	// GET config
	req := httptest.NewRequest("GET", "/api/admin/config", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp)
	}
	cfg, ok := resp["config"].(map[string]any)
	if !ok {
		t.Fatal("expected config field")
	}
	if cfg["disk_warn_pct"] == nil {
		t.Fatal("expected disk_warn_pct in config")
	}

	// POST update config
	body := `{"disk_warn_pct": 75, "disk_fail_pct": 90}`
	req = httptest.NewRequest("POST", "/api/admin/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler(w, req)

	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp)
	}

	// Verify updated config
	req = httptest.NewRequest("GET", "/api/admin/config", nil)
	w = httptest.NewRecorder()
	handler(w, req)

	json.NewDecoder(w.Body).Decode(&resp)
	cfg, ok = resp["config"].(map[string]any)
	if !ok {
		t.Fatal("expected config field")
	}
	if cfg["disk_warn_pct"] != float64(75) {
		t.Fatalf("expected disk_warn_pct=75, got %v", cfg["disk_warn_pct"])
	}
}


// TestAuditLogHandler handles the TestAuditLogHandler HTTP request.
func TestAuditLogHandler(t *testing.T) {
	handler := AuditLogHandler()

	// Log some entries
	logAudit("test_action", "tester", "test detail")

	// GET audit log
	req := httptest.NewRequest("GET", "/api/admin/audit-log?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp)
	}
	entries, ok := resp["entries"].([]any)
	if !ok {
		t.Fatal("expected entries array")
	}
	if len(entries) < 1 {
		t.Fatal("expected at least 1 audit entry")
	}
}


// TestWebhookQueueStats handles the TestWebhookQueueStats HTTP request.
func TestWebhookQueueStats(t *testing.T) {
	if GlobalWebhookQueue == nil {
		t.Skip("GlobalWebhookQueue not initialized")
	}

	handler := WebhookQueueStatsHandler()
	req := httptest.NewRequest("GET", "/api/admin/webhook-queue/stats", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp)
	}
}


// TestRateLimitConfigV2Handler handles the TestRateLimitConfigV2Handler HTTP request.
func TestRateLimitConfigV2Handler(t *testing.T) {
	GlobalPerEndpointLimiter = NewPerEndpointRateLimiter(t.TempDir())
	handler := RateLimitConfigHandlerV2()

	// GET
	req := httptest.NewRequest("GET", "/api/admin/rate-limit-config/v2", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp)
	}

	// PUT
	body := `{"global": {"requests_per_sec": 20, "burst": 50}, "endpoints": {"/api/test": {"requests_per_sec": 5, "burst": 10}}}`
	req = httptest.NewRequest("PUT", "/api/admin/rate-limit-config/v2", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler(w, req)

	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true on PUT, got %v", resp)
	}
}


// TestContentFilterEngine handles the TestContentFilterEngine HTTP request.
func TestContentFilterEngine(t *testing.T) {
	filter := NewContentFilterEngine(t.TempDir())

	// Add rules
	filter.AddRule("badword", "block", "")
	filter.AddRule("replaceme", "replace", "***")
	filter.AddRule("spamword", "flag", "")

	// Test block
	_, _, blocked := filter.Filter("this has badword in it")
	if !blocked {
		t.Fatal("expected blocked")
	}

	// Test replace
	text, _, blocked := filter.Filter("please replaceme")
	if blocked {
		t.Fatal("expected not blocked")
	}
	if text != "please ***" {
		t.Fatalf("expected 'please ***', got '%s'", text)
	}

	// Test flag
	_, note, blocked := filter.Filter("this has spamword")
	if !blocked {
		t.Fatal("expected blocked for flag")
	}
	if note == "" {
		t.Fatal("expected note for flagged message")
	}

	// Test clean text
	text, _, blocked = filter.Filter("this is clean")
	if blocked {
		t.Fatal("expected not blocked")
	}
	if text != "this is clean" {
		t.Fatalf("expected original text, got '%s'", text)
	}

	// Test get rules
	rules := filter.GetRules()
	expectCount := 9 // 6 defaults + 3 added
	if len(rules) != expectCount {
		t.Fatalf("expected %d rules, got %d", expectCount, len(rules))
	}

	// Test delete by word
	filter.RemoveRule(rules[0].Word)
	rules = filter.GetRules()
	if len(rules) != expectCount-1 {
		t.Fatalf("expected %d rules after delete, got %d", expectCount-1, len(rules))
	}
}
