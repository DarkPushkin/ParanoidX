// Package config provides configuration management for the server
package config

import (
	"os"
	"path/filepath"
	"testing"
)


// TestDefaultConfig handles the TestDefaultConfig HTTP request.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Listen != "0.0.0.0:8080" {
		t.Fatalf("expected default listen, got %q", cfg.Listen)
	}
	if cfg.DataDir == "" {
		t.Fatal("expected non-empty DataDir")
	}
	if cfg.VaultQuotaMB != 2048 {
		t.Fatalf("expected 2048 MB vault quota, got %d", cfg.VaultQuotaMB)
	}
	if cfg.BillingPricesNg.InitSilverRound <= 0 {
		t.Fatal("expected positive InitSilverRound price")
	}
}


// TestLoadFromFile handles the TestLoadFromFile HTTP request.
func TestLoadFromFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfgPath := filepath.Join(dir, "simplex-node.json")
	content := `{"listen": "0.0.0.0:9090", "vault_quota_mb": 4096}`
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := Load(cfgPath)
	if cfg.Listen != "0.0.0.0:9090" {
		t.Fatalf("expected listen from file, got %q", cfg.Listen)
	}
	if cfg.VaultQuotaMB != 4096 {
		t.Fatalf("expected vault quota from file, got %d", cfg.VaultQuotaMB)
	}
}


// TestLoadEmptyPath handles the TestLoadEmptyPath HTTP request.
func TestLoadEmptyPath(t *testing.T) {
	cfg := Load("")
	if cfg.Listen != "0.0.0.0:8080" {
		t.Fatalf("expected default listen, got %q", cfg.Listen)
	}
}


// TestLoadNonExistentFile handles the TestLoadNonExistentFile HTTP request.
func TestLoadNonExistentFile(t *testing.T) {
	cfg := Load("/nonexistent/path/config.json")
	if cfg.Listen != "0.0.0.0:8080" {
		t.Fatalf("expected default listen for missing file, got %q", cfg.Listen)
	}
}


// TestLoadMergesDefaults handles the TestLoadMergesDefaults HTTP request.
func TestLoadMergesDefaults(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	// Only set listen, everything else should use defaults
	cfgPath := filepath.Join(dir, "partial.json")
	content := `{"listen": "127.0.0.1:8080"}`
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := Load(cfgPath)
	if cfg.Listen != "127.0.0.1:8080" {
		t.Fatalf("expected listen from file, got %q", cfg.Listen)
	}
	if cfg.VaultQuotaMB != 2048 {
		t.Fatalf("expected default vault quota, got %d", cfg.VaultQuotaMB)
	}
	if cfg.DataDir == "" {
		t.Fatal("expected default DataDir")
	}
}


// TestSave handles the TestSave HTTP request.
func TestSave(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfg := DefaultConfig()
	cfg.Listen = "0.0.0.0:9090"
	cfg.VaultQuotaMB = 8192

	savePath := filepath.Join(dir, "saved.json")
	if err := cfg.Save(savePath); err != nil {
		t.Fatal("Save:", err)
	}

	loaded := Load(savePath)
	if loaded.Listen != "0.0.0.0:9090" {
		t.Fatalf("expected listen 0.0.0.0:9090, got %q", loaded.Listen)
	}
	if loaded.VaultQuotaMB != 8192 {
		t.Fatalf("expected vault 8192, got %d", loaded.VaultQuotaMB)
	}
}


// TestLoadInvalidJSON handles the TestLoadInvalidJSON HTTP request.
func TestLoadInvalidJSON(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfgPath := filepath.Join(dir, "bad.json")
	content := `not json at all`
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := Load(cfgPath)
	if cfg.Listen != "0.0.0.0:8080" {
		t.Fatalf("expected default listen after bad JSON, got %q", cfg.Listen)
	}
}


// TestLoadBillingOverride handles the TestLoadBillingOverride HTTP request.
func TestLoadBillingOverride(t *testing.T) {
	dir, err := os.MkdirTemp("", "config-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfgPath := filepath.Join(dir, "billing.json")
	content := `{"billing_prices_ng": {"init_silver_round": 999}}`
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := Load(cfgPath)
	if cfg.BillingPricesNg.InitSilverRound != 999 {
		t.Fatalf("expected 999, got %d", cfg.BillingPricesNg.InitSilverRound)
	}
	if cfg.BillingPricesNg.RwaRegister != 50000 {
		t.Fatalf("expected default RWA price 50000, got %d", cfg.BillingPricesNg.RwaRegister)
	}
}
