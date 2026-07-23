// Package treasury implements treasury round management
package treasury

import (
	"os"
	"path/filepath"
	"testing"
)


// TestReservePaths handles the TestReservePaths HTTP request.
func TestReservePaths(t *testing.T) {
	if reservePath("/tmp") != "/tmp/silver_reserve_ng.txt" {
		t.Fatal("unexpected reserve path")
	}
	if registryPath("/tmp") != "/tmp/banknotes_registry.json" {
		t.Fatal("unexpected registry path")
	}
}


// TestReadWriteReserve handles the TestReadWriteReserve HTTP request.
func TestReadWriteReserve(t *testing.T) {
	dir := t.TempDir()
	initial := readReserve(dir)
	if initial != 0 {
		t.Fatalf("expected 0, got %d", initial)
	}

	writeReserve(dir, 1000000000000)
	val := readReserve(dir)
	if val != 1000000000000 {
		t.Fatalf("expected 1000000000000, got %d", val)
	}

	writeReserve(dir, 0)
	val = readReserve(dir)
	if val != 0 {
		t.Fatalf("expected 0, got %d", val)
	}
}


// TestReadWriteHolders handles the TestReadWriteHolders HTTP request.
func TestReadWriteHolders(t *testing.T) {
	dir := t.TempDir()
	holders := readHolders(dir)
	if len(holders) != 0 {
		t.Fatal("expected empty holders")
	}

	testHolders := []map[string]any{
		{"holder": "alice", "denomination_tlr": 100.0, "serial": "S001"},
		{"holder": "bob", "denomination_tlr": 200.0, "serial": "S002"},
	}
	writeHolders(dir, testHolders)

	loaded := readHolders(dir)
	if len(loaded) != 2 {
		t.Fatalf("expected 2 holders, got %d", len(loaded))
	}
	if loaded[0]["holder"] != "alice" {
		t.Fatal("expected alice as first holder")
	}
}


// TestExecuteRound handles the TestExecuteRound HTTP request.
func TestExecuteRound(t *testing.T) {
	dir := t.TempDir()
	writeHolders(dir, []map[string]any{
		{"holder": "addr1", "denomination_tlr": 500.0, "serial": "S001"},
	})
	writeReserve(dir, 1000000000000)

	result, err := ExecuteRound(RoundParams{
		DataDir:     dir,
		USDT:        2000.0,
		AnnounceDir: dir,
	})
	if err != nil {
		t.Fatalf("ExecuteRound: %v", err)
	}
	if result.USDTIn != 2000.0 {
		t.Fatalf("expected USDTIn 2000, got %f", result.USDTIn)
	}
	if result.NewSilverNg <= 0 {
		t.Fatal("expected positive NewSilverNg")
	}
	if result.TreasuryShareNg <= 0 {
		t.Fatal("expected positive TreasuryShareNg")
	}
	if result.CurrentReserve <= 1000000000000 {
		t.Fatal("expected reserve to increase")
	}
	if len(result.Dividends) != 1 {
		t.Fatalf("expected 1 dividend, got %d", len(result.Dividends))
	}
	if result.Dividends[0].Holder != "addr1" {
		t.Fatal("expected dividend for addr1")
	}
}


// TestExecuteRoundEmptyHolders handles the TestExecuteRoundEmptyHolders HTTP request.
func TestExecuteRoundEmptyHolders(t *testing.T) {
	dir := t.TempDir()
	writeReserve(dir, 500000000000)

	result, err := ExecuteRound(RoundParams{
		DataDir:     dir,
		USDT:        1000.0,
		AnnounceDir: dir,
	})
	if err != nil {
		t.Fatalf("ExecuteRound: %v", err)
	}
	if len(result.Dividends) != 0 {
		t.Fatalf("expected 0 dividends with no holders")
	}
}


// TestExecuteRoundZeroUSDT handles the TestExecuteRoundZeroUSDT HTTP request.
func TestExecuteRoundZeroUSDT(t *testing.T) {
	dir := t.TempDir()
	writeReserve(dir, 0)

	result, err := ExecuteRound(RoundParams{
		DataDir:     dir,
		USDT:        0,
		AnnounceDir: dir,
	})
	if err != nil {
		t.Fatalf("ExecuteRound: %v", err)
	}
	// When USDT is 0, it defaults to 1000
	if result.USDTIn != 1000.0 {
		t.Fatalf("expected default USDTIn 1000, got %f", result.USDTIn)
	}
}


// TestFileCreation handles the TestFileCreation HTTP request.
func TestFileCreation(t *testing.T) {
	dir := t.TempDir()
	rPath := reservePath(dir)
	if _, err := os.Stat(rPath); !os.IsNotExist(err) {
		t.Fatal("reserve file should not exist yet")
	}

	writeReserve(dir, 999)
	if _, err := os.Stat(rPath); err != nil {
		t.Fatal("reserve file should exist")
	}

	data, _ := os.ReadFile(rPath)
	if string(data) != "999" {
		t.Fatalf("expected '999', got '%s'", string(data))
	}
}


// TestRoundLogPath handles the TestRoundLogPath HTTP request.
func TestRoundLogPath(t *testing.T) {
	p := roundLogPath("/data/test")
	if !filepath.IsAbs(p) {
		t.Fatal("expected absolute path")
	}
	if filepath.Base(p) != "silver_rounds.log" {
		t.Fatal("expected silver_rounds.log")
	}
}
