// Package treasury monitors USDT TRC20 deposits and manages silver reserves.
package treasury

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	USDTContract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	DefaultAPI   = "https://api.trongrid.io"
)

type Deposit struct {
	TxID       string  `json:"tx"`
	From       string  `json:"from"`
	AmountUSDT float64 `json:"amount_usdt"`
	Time       string  `json:"time"`
	Logged     bool    `json:"-"`
}

type tronTx struct {
	TransactionID  string `json:"transaction_id"`
	From           string `json:"from"`
	To             string `json:"to"`
	Value          string `json:"value"`
	BlockTimestamp int64  `json:"block_timestamp"`
}

type tronResp struct {
	Data []tronTx `json:"data"`
}

type Monitor struct {
	mu          sync.Mutex
	DataDir     string
	Addr        string
	APIKey      string
	APIBase     string
	Deposits    []Deposit
	TotalUSDT   float64
	Seen        map[string]bool
	LastPoll    time.Time
}


// New handles the New HTTP request.
func New(dataDir, addr, apiKey string) *Monitor {
	return &Monitor{
		DataDir:  dataDir,
		Addr:     addr,
		APIKey:   apiKey,
		APIBase:  DefaultAPI,
		Deposits: []Deposit{},
		Seen:     make(map[string]bool),
	}
}


// Start handles the Start HTTP request.
func (m *Monitor) Start() {
	slog.Info("tron monitor: starting", "addr", m.Addr)
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		m.poll()
		<-ticker.C
	}
}

func (m *Monitor) poll() {
	if m.Addr == "" || m.Addr == "TYourTreasuryAddressHereForDemo" {
		return
	}
	base := m.APIBase
	if base == "" {
		base = DefaultAPI
	}
	url := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?contract_address=%s&limit=50&only_to=true&order_by=block_timestamp,desc", base, m.Addr, USDTContract)
	if m.APIKey != "" {
		url += "&api_key=" + m.APIKey
	}
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("tron monitor: query failed", "err", err)
		return
	}
	defer resp.Body.Close()
	var result tronResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("tron monitor: decode failed", "err", err)
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, tx := range result.Data {
		if tx.To != m.Addr {
			continue
		}
		if m.Seen[tx.TransactionID] {
			continue
		}
		m.Seen[tx.TransactionID] = true
		amt := 0.0
		fmt.Sscanf(tx.Value, "%f", &amt)
		amt /= 1_000_000.0
		d := Deposit{
			TxID:       tx.TransactionID,
			From:       tx.From,
			AmountUSDT: amt,
			Time:       time.UnixMilli(tx.BlockTimestamp).Format(time.RFC3339),
		}
		m.Deposits = append(m.Deposits, d)
		m.TotalUSDT += amt
		m.appendLog(d)
		slog.Info("tron monitor: new deposit", "tx", d.TxID, "from", d.From, "usdt", d.AmountUSDT)
	}
	m.LastPoll = time.Now()
}

func (m *Monitor) appendLog(d Deposit) {
	p := filepath.Join(m.DataDir, "treasury_usdt.log")
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	entry, _ := json.Marshal(d)
	fmt.Fprintln(f, string(entry))
}


// ReadLog handles the ReadLog HTTP request.
func (m *Monitor) ReadLog() string {
	p := filepath.Join(m.DataDir, "treasury_usdt.log")
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}


// Stats handles the Stats HTTP request.
func (m *Monitor) Stats() (float64, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.TotalUSDT, len(m.Deposits)
}

// AllDeposits returns a copy of all deposits for safe concurrent access.
func (m *Monitor) AllDeposits() []Deposit {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Deposit, len(m.Deposits))
	copy(out, m.Deposits)
	return out
}

// PollNow triggers an immediate poll (for API handlers).
func (m *Monitor) PollNow() {
	m.poll()
}
