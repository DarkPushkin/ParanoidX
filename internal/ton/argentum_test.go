// Package ton provides TON/Argentum integration
package ton

import "testing"


// TestArgentumConstants handles the TestArgentumConstants HTTP request.
func TestArgentumConstants(t *testing.T) {
	if ArgentumSymbol != "ARGENTUM" {
		t.Fatalf("expected ARGENTUM, got %s", ArgentumSymbol)
	}
	if ArgentumDecimals != 9 {
		t.Fatalf("expected 9 decimals, got %d", ArgentumDecimals)
	}
	if SwapFeeBPS != 50 {
		t.Fatalf("expected 50 bps swap fee, got %d", SwapFeeBPS)
	}
	if MinSwapNg != 1_000_000 {
		t.Fatalf("expected 1M min swap, got %d", MinSwapNg)
	}
}


// TestGetRate handles the TestGetRate HTTP request.
func TestGetRate(t *testing.T) {
	tests := []struct {
		silverUSD float64
		tonUSD    float64
		wantZero  bool
	}{
		{75.0, 2.5, false},
		{0, 2.5, true},
		{75.0, 0, true},
		{0, 0, true},
	}
	for _, tc := range tests {
		rate := GetRate(tc.silverUSD, tc.tonUSD)
		if tc.wantZero && rate != 0 {
			t.Fatalf("expected 0 for silver=%f ton=%f, got %f", tc.silverUSD, tc.tonUSD, rate)
		}
		if !tc.wantZero && rate <= 0 {
			t.Fatalf("expected positive rate for silver=%f ton=%f, got %f", tc.silverUSD, tc.tonUSD, rate)
		}
	}
}


// TestNewTonAPI handles the TestNewTonAPI HTTP request.
func TestNewTonAPI(t *testing.T) {
	api := NewTonAPI("test-key")
	if api.BaseURL != "https://toncenter.com/api/v2" {
		t.Fatalf("expected toncenter URL, got %s", api.BaseURL)
	}
	if api.APIKey != "test-key" {
		t.Fatalf("expected test-key, got %s", api.APIKey)
	}
}
