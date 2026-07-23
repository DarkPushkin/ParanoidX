// Package billing provides billing and subscription management functionality
package billing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SilverPayment struct {
	Time     string `json:"time"`
	AmountNg int64  `json:"amount_ng"`
	For      string `json:"for"`
	Ref      string `json:"ref,omitempty"`
}

type Prices struct {
	InitSilverRoundNg int64 `json:"init_silver_round_ng"`
	RwaRegisterNg     int64 `json:"rwa_register_ng"`
}

type Service struct {
	LogFile    string
	PricesFile string
}


// New handles the New HTTP request.
func New(dataDir string) *Service {
	return &Service{
		LogFile:    filepath.Join(dataDir, "payments.log"),
		PricesFile: filepath.Join(dataDir, "billing_prices.json"),
	}
}


// RecordPayment handles the RecordPayment HTTP request.
func (s *Service) RecordPayment(amountNg int64, forWhat, ref string) {
	if amountNg <= 0 {
		return
	}
	p := SilverPayment{
		Time:     time.Now().Format(time.RFC3339),
		AmountNg: amountNg,
		For:      forWhat,
		Ref:      ref,
	}
	if f, err := os.OpenFile(s.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer f.Close()
		if b, err := json.Marshal(p); err == nil {
			fmt.Fprintln(f, string(b))
		}
	}
}


// GetPrices handles the GetPrices HTTP request.
func (s *Service) GetPrices() Prices {
	p := Prices{InitSilverRoundNg: 5_000_000_000, RwaRegisterNg: 1_000_000_000}
	if b, err := os.ReadFile(s.PricesFile); err == nil {
		json.Unmarshal(b, &p)
	}
	return p
}


// RecentPayments handles the RecentPayments HTTP request.
func (s *Service) RecentPayments(max int) []SilverPayment {
	payments := []SilverPayment{}
	if b, err := os.ReadFile(s.LogFile); err == nil {
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")
		for i := len(lines) - 1; i >= 0 && len(payments) < max; i-- {
			if lines[i] == "" {
				continue
			}
			var p SilverPayment
			if json.Unmarshal([]byte(lines[i]), &p) == nil {
				payments = append(payments, p)
			}
		}
	}
	return payments
}
