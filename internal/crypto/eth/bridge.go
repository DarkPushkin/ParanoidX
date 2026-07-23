// Package eth provides Ethereum bridge and status tracking
package eth

import (
	"fmt"
	"time"
)

type BridgeDirection string

const (
	BridgeLock   BridgeDirection = "lock"
	BridgeBurn   BridgeDirection = "burn"
	BridgeUnlock BridgeDirection = "unlock"
	BridgeMint   BridgeDirection = "mint"
)

type BridgeStatus string

const (
	BridgePending  BridgeStatus = "pending"
	BridgeVerified BridgeStatus = "verified"
	BridgeComplete BridgeStatus = "complete"
	BridgeFailed   BridgeStatus = "failed"
)

type BridgeTransfer struct {
	ID           string        `json:"id"`
	FromChain    string        `json:"from_chain"`
	ToChain      string        `json:"to_chain"`
	Token        string        `json:"token"`
	Amount       int64         `json:"amount"`
	Sender       string        `json:"sender"`
	Recipient    string        `json:"recipient"`
	Direction    BridgeDirection `json:"direction"`
	Status       BridgeStatus  `json:"status"`
	TxHash       string        `json:"tx_hash,omitempty"`
	ProofTxHash  string        `json:"proof_tx_hash,omitempty"`
	CreatedAt    string        `json:"created_at"`
	CompletedAt  string        `json:"completed_at,omitempty"`
}

type BridgeRegistry struct {
	transfers map[string]*BridgeTransfer
}


// NewBridgeRegistry handles the NewBridgeRegistry HTTP request.
func NewBridgeRegistry() *BridgeRegistry {
	return &BridgeRegistry{transfers: make(map[string]*BridgeTransfer)}
}


// Create handles the Create HTTP request.
func (r *BridgeRegistry) Create(fromChain, toChain, token, sender, recipient string, amount int64) *BridgeTransfer {
	id := fmt.Sprintf("bridge-eth-%d", time.Now().UnixNano())
	t := &BridgeTransfer{
		ID:        id,
		FromChain: fromChain,
		ToChain:   toChain,
		Token:     token,
		Amount:    amount,
		Sender:    sender,
		Recipient: recipient,
		Status:    BridgePending,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	r.transfers[id] = t
	return t
}


// Get handles the Get HTTP request.
func (r *BridgeRegistry) Get(id string) *BridgeTransfer {
	return r.transfers[id]
}


// List handles the List HTTP request.
func (r *BridgeRegistry) List() []*BridgeTransfer {
	result := make([]*BridgeTransfer, 0, len(r.transfers))
	for _, t := range r.transfers {
		result = append(result, t)
	}
	return result
}


// Confirm handles the Confirm HTTP request.
func (r *BridgeRegistry) Confirm(id, txHash string) (*BridgeTransfer, error) {
	t, ok := r.transfers[id]
	if !ok {
		return nil, fmt.Errorf("transfer not found")
	}
	if t.Status != BridgePending {
		return nil, fmt.Errorf("transfer already %s", t.Status)
	}
	t.Status = BridgeVerified
	t.TxHash = txHash
	return t, nil
}


// Complete handles the Complete HTTP request.
func (r *BridgeRegistry) Complete(id, proofTxHash string) (*BridgeTransfer, error) {
	t, ok := r.transfers[id]
	if !ok {
		return nil, fmt.Errorf("transfer not found")
	}
	if t.Status != BridgeVerified {
		return nil, fmt.Errorf("transfer must be verified first, current: %s", t.Status)
	}
	t.Status = BridgeComplete
	t.ProofTxHash = proofTxHash
	t.CompletedAt = time.Now().Format(time.RFC3339)
	return t, nil
}


// Fail handles the Fail HTTP request.
func (r *BridgeRegistry) Fail(id string, err error) {
	if t, ok := r.transfers[id]; ok {
		t.Status = BridgeFailed
	}
}
