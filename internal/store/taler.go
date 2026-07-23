// Package store provides persistent storage backends
package store

import (
	"fmt"
	"sync"
	"time"
)

// TalerTransaction represents a Liquid Taler transfer.
type TalerTransaction struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    int64     `json:"amount"`
	Fee       int64     `json:"fee"`
	Memo      string    `json:"memo,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Confirmed bool      `json:"confirmed"`
}

// TalerStore manages Liquid Taler balances and transactions via SQLite.
type TalerStore struct {
	db *DB
	mu sync.RWMutex
}


// NewTalerStore handles the NewTalerStore HTTP request.
func NewTalerStore(db *DB) *TalerStore {
	s := &TalerStore{db: db}
	s.db.mu.Lock()
	defer s.db.mu.Unlock()
	s.db.Exec(`CREATE TABLE IF NOT EXISTS taler_balances (
		pubkey TEXT PRIMARY KEY,
		balance INTEGER NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL
	)`)
	s.db.Exec(`CREATE TABLE IF NOT EXISTS taler_transactions (
		id TEXT PRIMARY KEY,
		from_pubkey TEXT NOT NULL,
		to_pubkey TEXT NOT NULL,
		amount INTEGER NOT NULL,
		fee INTEGER NOT NULL DEFAULT 0,
		memo TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		confirmed INTEGER NOT NULL DEFAULT 1
	)`)
	return s
}


// Balance handles the Balance HTTP request.
func (s *TalerStore) Balance(pubkey string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var bal int64
	if err := s.db.QueryRow(`SELECT balance FROM taler_balances WHERE pubkey = ?`, pubkey).Scan(&bal); err != nil {
		_ = err // non-existent account → bal=0 is correct
	}
	return bal
}


// Transfer handles the Transfer HTTP request.
func (s *TalerStore) Transfer(from, to string, amount int64) (*TalerTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fee := amount * 228 / 10000
	total := amount + fee

	var fromBal int64
	if err := s.db.QueryRow(`SELECT balance FROM taler_balances WHERE pubkey = ?`, from).Scan(&fromBal); err != nil {
		return nil, fmt.Errorf("get from balance: %w", err)
	}
	if fromBal < total {
		return nil, fmt.Errorf("insufficient balance: have %d, need %d", fromBal, total)
	}

	now := formatTime(time.Now())
	tx := &TalerTransaction{
		ID:        "tlr-" + now + "-" + from,
		From:      from, To: to, Amount: amount, Fee: fee,
		CreatedAt: time.Now(), Confirmed: true,
	}

	s.db.Exec(`INSERT OR REPLACE INTO taler_balances (pubkey, balance, updated_at) VALUES (?, ?, ?)`, from, fromBal-total, now)

	var toBal int64
	if err := s.db.QueryRow(`SELECT balance FROM taler_balances WHERE pubkey = ?`, to).Scan(&toBal); err != nil {
		_ = err // non-existent account → 0
	}
	s.db.Exec(`INSERT OR REPLACE INTO taler_balances (pubkey, balance, updated_at) VALUES (?, ?, ?)`, to, toBal+amount, now)

	var treBal int64
	if err := s.db.QueryRow(`SELECT balance FROM taler_balances WHERE pubkey = ?`, "treasury").Scan(&treBal); err != nil {
		_ = err // non-existent treasury → 0
	}
	s.db.Exec(`INSERT OR REPLACE INTO taler_balances (pubkey, balance, updated_at) VALUES (?, ?, ?)`, "treasury", treBal+fee, now)

	s.db.Exec(`INSERT INTO taler_transactions (id, from_pubkey, to_pubkey, amount, fee, memo, created_at, confirmed) VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
		tx.ID, from, to, amount, fee, tx.Memo, now)

	return tx, nil
}


// History handles the History HTTP request.
func (s *TalerStore) History(pubkey string, limit int) ([]TalerTransaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, from_pubkey, to_pubkey, amount, fee, memo, created_at, confirmed FROM taler_transactions WHERE from_pubkey = ? OR to_pubkey = ? ORDER BY created_at DESC`
	args := []any{pubkey, pubkey}
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []TalerTransaction
	for rows.Next() {
		var tx TalerTransaction
		var createdAt string
		if err := rows.Scan(&tx.ID, &tx.From, &tx.To, &tx.Amount, &tx.Fee, &tx.Memo, &createdAt, &tx.Confirmed); err != nil {
			return nil, err
		}
		tx.CreatedAt, _ = parseTime(createdAt)
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}


// Issue handles the Issue HTTP request.
func (s *TalerStore) Issue(pubkey string, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := formatTime(time.Now())
	var bal int64
	if err := s.db.QueryRow(`SELECT balance FROM taler_balances WHERE pubkey = ?`, pubkey).Scan(&bal); err != nil {
		_ = err // non-existent account → 0
	}
	_, err := s.db.Exec(`INSERT OR REPLACE INTO taler_balances (pubkey, balance, updated_at) VALUES (?, ?, ?)`, pubkey, bal+amount, now)
	return err
}
