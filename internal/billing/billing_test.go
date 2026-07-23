// Package billing provides billing and subscription management functionality
package billing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)


// TestNew handles the TestNew HTTP request.
func TestNew(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if s.LogFile == "" {
		t.Fatal("expected log file path")
	}
	if !strings.HasPrefix(s.LogFile, dir) {
		t.Fatalf("expected log file in data dir, got %s", s.LogFile)
	}
}


// TestRecordPayment handles the TestRecordPayment HTTP request.
func TestRecordPayment(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.RecordPayment(1000, "test_tx", "ref1")

	b, err := os.ReadFile(s.LogFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "ref1") {
		t.Fatalf("expected ref1 in log, got: %s", string(b))
	}
}


// TestRecordPaymentZeroAmount handles the TestRecordPaymentZeroAmount HTTP request.
func TestRecordPaymentZeroAmount(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.RecordPayment(0, "zero", "")

	if _, err := os.Stat(s.LogFile); err == nil {
		t.Fatal("expected no file for zero amount")
	}
}


// TestGetPricesDefault handles the TestGetPricesDefault HTTP request.
func TestGetPricesDefault(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	p := s.GetPrices()
	if p.InitSilverRoundNg != 5_000_000_000 {
		t.Fatalf("expected 5_000_000_000, got %d", p.InitSilverRoundNg)
	}
	if p.RwaRegisterNg != 1_000_000_000 {
		t.Fatalf("expected 1_000_000_000, got %d", p.RwaRegisterNg)
	}
}


// TestGetPricesCustom handles the TestGetPricesCustom HTTP request.
func TestGetPricesCustom(t *testing.T) {
	dir := t.TempDir()
	pf := filepath.Join(dir, "billing_prices.json")
	os.WriteFile(pf, []byte(`{"init_silver_round_ng":100,"rwa_register_ng":200}`), 0644)

	s := New(dir)
	p := s.GetPrices()
	if p.InitSilverRoundNg != 100 {
		t.Fatalf("expected 100, got %d", p.InitSilverRoundNg)
	}
	if p.RwaRegisterNg != 200 {
		t.Fatalf("expected 200, got %d", p.RwaRegisterNg)
	}
}


// TestRecentPayments handles the TestRecentPayments HTTP request.
func TestRecentPayments(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	s.RecordPayment(10, "a", "")
	s.RecordPayment(20, "b", "")
	s.RecordPayment(30, "c", "")

	recent := s.RecentPayments(2)
	if len(recent) != 2 {
		t.Fatalf("expected 2 recent, got %d", len(recent))
	}
	if recent[0].For != "c" {
		t.Fatalf("expected most recent 'c', got %s", recent[0].For)
	}
	if recent[1].For != "b" {
		t.Fatalf("expected second 'b', got %s", recent[1].For)
	}
}


// TestRecentPaymentsMaxLarge handles the TestRecentPaymentsMaxLarge HTTP request.
func TestRecentPaymentsMaxLarge(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.RecordPayment(1, "x", "")
	s.RecordPayment(2, "y", "")

	recent := s.RecentPayments(100)
	if len(recent) != 2 {
		t.Fatalf("expected 2, got %d", len(recent))
	}
}


// TestRecentPaymentsEmpty handles the TestRecentPaymentsEmpty HTTP request.
func TestRecentPaymentsEmpty(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	recent := s.RecentPayments(10)
	if len(recent) != 0 {
		t.Fatalf("expected 0, got %d", len(recent))
	}
}
