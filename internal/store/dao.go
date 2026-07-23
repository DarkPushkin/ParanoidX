// Package store provides persistent storage backends
package store

import (
	"log"
	"sync"
	"time"
)

// ProposalStatus tracks the lifecycle of a governance proposal.
type ProposalStatus string

const (
	ProposalActive   ProposalStatus = "active"
	ProposalPassed   ProposalStatus = "passed"
	ProposalRejected ProposalStatus = "rejected"
	ProposalExpired  ProposalStatus = "expired"
)

// Proposal represents a DAO governance proposal.
type Proposal struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	Status      ProposalStatus `json:"status"`
	VotesFor    int64          `json:"votes_for"`
	VotesAgainst int64         `json:"votes_against"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
}

// DAOStore manages governance proposals and voting via SQLite.
type DAOStore struct {
	db *DB
	mu sync.RWMutex
}


// NewDAOStore handles the NewDAOStore HTTP request.
func NewDAOStore(db *DB) *DAOStore {
	s := &DAOStore{db: db}
	s.db.mu.Lock()
	defer s.db.mu.Unlock()
	s.db.Exec(`CREATE TABLE IF NOT EXISTS dao_proposals (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		author TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		votes_for INTEGER NOT NULL DEFAULT 0,
		votes_against INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		expires_at TEXT NOT NULL
	)`)
	s.db.Exec(`CREATE TABLE IF NOT EXISTS dao_votes (
		id TEXT PRIMARY KEY,
		proposal_id TEXT NOT NULL,
		voter TEXT NOT NULL,
		support INTEGER NOT NULL,
		weight INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL,
		UNIQUE(proposal_id, voter)
	)`)
	return s
}


// CreateProposal handles the CreateProposal HTTP request.
func (s *DAOStore) CreateProposal(title, description, author string, duration time.Duration) (*Proposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	p := &Proposal{
		ID:          "prop-" + now.Format("150405") + "-" + author,
		Title:       title,
		Description: description,
		Author:      author,
		Status:      ProposalActive,
		CreatedAt:   now,
		ExpiresAt:   now.Add(duration),
	}

	_, err := s.db.Exec(`INSERT INTO dao_proposals (id, title, description, author, status, votes_for, votes_against, created_at, expires_at) VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?)`,
		p.ID, p.Title, p.Description, p.Author, string(p.Status), formatTime(p.CreatedAt), formatTime(p.ExpiresAt))
	if err != nil {
		return nil, err
	}
	log.Printf("[dao] proposal created: %s — %s", p.ID, p.Title)
	return p, nil
}


// Vote handles the Vote HTTP request.
func (s *DAOStore) Vote(proposalID, voter string, support bool, weight int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check proposal exists and is active
	var status string
	var expiresAt string
	err := s.db.QueryRow(`SELECT status, expires_at FROM dao_proposals WHERE id = ?`, proposalID).Scan(&status, &expiresAt)
	if err != nil {
		return err
	}
	if status != string(ProposalActive) {
		return nil
	}
	exp, _ := parseTime(expiresAt)
	if time.Now().After(exp) {
		s.db.Exec(`UPDATE dao_proposals SET status = ? WHERE id = ?`, string(ProposalExpired), proposalID)
		return nil
	}

	now := formatTime(time.Now())
	voteID := proposalID + "-" + voter
	_, err = s.db.Exec(`INSERT OR REPLACE INTO dao_votes (id, proposal_id, voter, support, weight, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		voteID, proposalID, voter, boolToInt(support), weight, now)
	if err != nil {
		return err
	}

	if support {
		s.db.Exec(`UPDATE dao_proposals SET votes_for = votes_for + ? WHERE id = ?`, weight, proposalID)
	} else {
		s.db.Exec(`UPDATE dao_proposals SET votes_against = votes_against + ? WHERE id = ?`, weight, proposalID)
	}
	return nil
}


// ListProposals handles the ListProposals HTTP request.
func (s *DAOStore) ListProposals(status ProposalStatus) ([]Proposal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, title, description, author, status, votes_for, votes_against, created_at, expires_at FROM dao_proposals`
	args := []any{}
	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, string(status))
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Proposal
	for rows.Next() {
		var p Proposal
		var createdAt, expiresAt string
		var statusStr string
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Author, &statusStr, &p.VotesFor, &p.VotesAgainst, &createdAt, &expiresAt); err != nil {
			return nil, err
		}
		p.Status = ProposalStatus(statusStr)
		p.CreatedAt, _ = parseTime(createdAt)
		p.ExpiresAt, _ = parseTime(expiresAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

// ExchangeRate represents the value of Liquid Taler in a reference currency.
type ExchangeRate struct {
	Currency string  `json:"currency"`
	Rate     float64 `json:"rate"`
	Updated  time.Time `json:"updated"`
}

// ExchangeOracle provides exchange rates for Liquid Taler.
type ExchangeOracle struct {
	mu   sync.RWMutex
	rate ExchangeRate
}


// NewExchangeOracle handles the NewExchangeOracle HTTP request.
func NewExchangeOracle() *ExchangeOracle {
	return &ExchangeOracle{
		rate: ExchangeRate{
			Currency: "TALER",
			Rate:     1.0,
			Updated:  time.Now(),
		},
	}
}


// SetRate handles the SetRate HTTP request.
func (o *ExchangeOracle) SetRate(rate float64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.rate.Rate = rate
	o.rate.Updated = time.Now()
	log.Printf("[exchange oracle] rate updated: 1 TALER = %.4f USD", rate)
}


// GetRate handles the GetRate HTTP request.
func (o *ExchangeOracle) GetRate() ExchangeRate {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.rate
}


// ConvertToUSD handles the ConvertToUSD HTTP request.
func (o *ExchangeOracle) ConvertToUSD(taler int64) float64 {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return float64(taler) * o.rate.Rate
}
