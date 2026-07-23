// Package btc provides Bitcoin swap functionality with HTLC support
package btc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type SwapStatus string

const (
	SwapPending    SwapStatus = "pending"
	SwapConfirmed  SwapStatus = "confirmed"
	SwapClaimed    SwapStatus = "claimed"
	SwapRefunded   SwapStatus = "refunded"
	SwapCancelled  SwapStatus = "cancelled"
	SwapExpired    SwapStatus = "expired"
)

type AtomicSwap struct {
	ID            string     `json:"id"`
	Initiator     string     `json:"initiator"`
	Counterparty  string     `json:"counterparty"`
	AmountNg      int64      `json:"amount_ng"`
	SecretHash    string     `json:"secret_hash"`
	LockTime      int64      `json:"lock_time"`      // UNIX timestamp
	RefundTime    int64      `json:"refund_time"`     // UNIX timestamp
	Status        SwapStatus `json:"status"`
	ClaimTxID     string     `json:"claim_tx_id,omitempty"`
	RefundTxID    string     `json:"refund_tx_id,omitempty"`
	CreatedAt     string     `json:"created_at"`
	BTCAddress    string     `json:"btc_address,omitempty"`
	BTCAmount     int64      `json:"btc_amount_sats,omitempty"`
}

type SwapRegistry struct {
	swaps map[string]*AtomicSwap
}


// NewRegistry handles the NewRegistry HTTP request.
func NewRegistry() *SwapRegistry {
	return &SwapRegistry{swaps: make(map[string]*AtomicSwap)}
}


// Create handles the Create HTTP request.
func (r *SwapRegistry) Create(initiator, counterparty string, amountNg int64, secretHash string) (*AtomicSwap, error) {
	if len(secretHash) != 64 {
		return nil, fmt.Errorf("secret_hash must be 64 hex chars (SHA-256)")
	}
	id := fmt.Sprintf("swap-btc-%d", time.Now().UnixNano())
	now := time.Now()
	swap := &AtomicSwap{
		ID:           id,
		Initiator:    initiator,
		Counterparty: counterparty,
		AmountNg:     amountNg,
		SecretHash:   secretHash,
		LockTime:     now.Add(24 * time.Hour).Unix(),
		RefundTime:   now.Add(48 * time.Hour).Unix(),
		Status:       SwapPending,
		CreatedAt:    now.Format(time.RFC3339),
	}
	r.swaps[id] = swap
	return swap, nil
}


// Get handles the Get HTTP request.
func (r *SwapRegistry) Get(id string) *AtomicSwap {
	return r.swaps[id]
}


// List handles the List HTTP request.
func (r *SwapRegistry) List() []*AtomicSwap {
	result := make([]*AtomicSwap, 0, len(r.swaps))
	for _, s := range r.swaps {
		result = append(result, s)
	}
	return result
}


// Claim handles the Claim HTTP request.
func (r *SwapRegistry) Claim(id, secret string) (*AtomicSwap, error) {
	swap, ok := r.swaps[id]
	if !ok {
		return nil, fmt.Errorf("swap not found")
	}
	if swap.Status != SwapPending {
		return nil, fmt.Errorf("swap already %s", swap.Status)
	}
	if time.Now().Unix() > swap.LockTime {
		swap.Status = SwapExpired
		return nil, fmt.Errorf("swap expired")
	}
	h := sha256.Sum256([]byte(secret))
	if hex.EncodeToString(h[:]) != swap.SecretHash {
		return nil, fmt.Errorf("invalid secret")
	}
	swap.Status = SwapClaimed
	swap.ClaimTxID = fmt.Sprintf("claim-%s", hex.EncodeToString(h[:8]))
	return swap, nil
}


// Refund handles the Refund HTTP request.
func (r *SwapRegistry) Refund(id string) (*AtomicSwap, error) {
	swap, ok := r.swaps[id]
	if !ok {
		return nil, fmt.Errorf("swap not found")
	}
	if swap.Status != SwapPending {
		return nil, fmt.Errorf("swap already %s", swap.Status)
	}
	if time.Now().Unix() < swap.RefundTime {
		return nil, fmt.Errorf("refund not yet available (wait until %d)", swap.RefundTime)
	}
	swap.Status = SwapRefunded
	swap.RefundTxID = fmt.Sprintf("refund-%s", id[:8])
	return swap, nil
}


// Confirm handles the Confirm HTTP request.
func (r *SwapRegistry) Confirm(id string) (*AtomicSwap, error) {
	swap, ok := r.swaps[id]
	if !ok {
		return nil, fmt.Errorf("swap not found")
	}
	if swap.Status != SwapPending {
		return nil, fmt.Errorf("swap already %s", swap.Status)
	}
	swap.Status = SwapConfirmed
	return swap, nil
}


// Cancel handles the Cancel HTTP request.
func (r *SwapRegistry) Cancel(id string) (*AtomicSwap, error) {
	swap, ok := r.swaps[id]
	if !ok {
		return nil, fmt.Errorf("swap not found")
	}
	if swap.Status == SwapClaimed {
		return nil, fmt.Errorf("cannot cancel claimed swap")
	}
	if swap.Status == SwapExpired || swap.Status == SwapCancelled {
		return nil, fmt.Errorf("swap already %s", swap.Status)
	}
	swap.Status = SwapCancelled
	return swap, nil
}


// CleanExpired handles the CleanExpired HTTP request.
func (r *SwapRegistry) CleanExpired() int {
	count := 0
	now := time.Now().Unix()
	for id, swap := range r.swaps {
		if swap.Status == SwapPending && now > swap.RefundTime+86400 {
			swap.Status = SwapExpired
			count++
			_ = id
		}
	}
	return count
}
