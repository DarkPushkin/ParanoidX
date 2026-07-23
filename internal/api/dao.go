// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

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

// ── DAO Governance (Cycle C3) ────────────────────────────────────────────────

// Proposal represents a governance proposal in the DAO.
type Proposal struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AuthorID    string    `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	Status      string    `json:"status"` // active, passed, rejected, executed
	VotesFor    int       `json:"votes_for"`
	VotesAgainst int      `json:"votes_against"`
	Voters      []Vote    `json:"voters"`
	CreatedAt   time.Time `json:"created_at"`
	Deadline    time.Time `json:"deadline"`
	ExecutedAt  time.Time `json:"executed_at,omitempty"`
	Action      string    `json:"action,omitempty"` // optional: what to execute on pass
}

// Vote records a single vote cast on a DAO proposal.
type Vote struct {
	VoterID   string `json:"voter_id"`
	VoterName string `json:"voter_name"`
	Support   bool   `json:"support"` // true = for, false = against
	Timestamp string `json:"timestamp"`
}

// DAOStore manages governance proposals, voting, and deadline execution.
type DAOStore struct {
	mu        sync.RWMutex
	path      string
	Proposals []Proposal `json:"proposals"`
	nextID    int64
	voters    map[string]bool // who has voted (per proposal)
}


// NewDAOStore creates a DAOStore and loads persisted proposals from disk.
func NewDAOStore(dataDir string) *DAOStore {
	store := &DAOStore{
		path:   filepath.Join(dataDir, "dao.json"),
		nextID: 1,
		voters: map[string]bool{},
	}
	store.load()
	return store
}

func (s *DAOStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, s); err != nil {
		slog.Warn("dao load", "error", err)
		return
	}
	for _, p := range s.Proposals {
		if id := parseDAOID(p.ID); id >= s.nextID {
			s.nextID = id + 1
		}
	}
	slog.Info("dao loaded", "proposals", len(s.Proposals))
}

func (s *DAOStore) save() {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		slog.Error("dao save", "error", err)
		return
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		slog.Error("dao save write", "error", err)
	}
}

func parseDAOID(id string) int64 {
	var n int64
	fmt.Sscanf(id, "prop-%d", &n)
	return n
}


// CreateProposal creates a new DAO governance proposal.
func (s *DAOStore) CreateProposal(title, desc, authorID, authorName, action string, durationHours int) *Proposal {
	s.mu.Lock()
	defer s.mu.Unlock()
	if durationHours <= 0 {
		durationHours = 72 // default 3 days
	}
	id := fmt.Sprintf("prop-%d", s.nextID)
	s.nextID++
	p := &Proposal{
		ID:          id,
		Title:       title,
		Description: desc,
		AuthorID:    authorID,
		AuthorName:  authorName,
		Status:      "active",
		CreatedAt:   time.Now(),
		Deadline:    time.Now().Add(time.Duration(durationHours) * time.Hour),
		Action:      action,
	}
	s.Proposals = append(s.Proposals, *p)
	s.save()
	return p
}


// Vote casts a for/against vote on an active DAO proposal.
func (s *DAOStore) Vote(proposalID, voterID, voterName string, support bool) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.Proposals {
		if p.ID != proposalID {
			continue
		}
		if p.Status != "active" {
			return false, "proposal is not active"
		}
		if time.Now().After(p.Deadline) {
			s.Proposals[i].Status = "rejected"
			s.save()
			return false, "deadline has passed"
		}
		// Check if already voted
		for _, v := range s.Proposals[i].Voters {
			if v.VoterID == voterID {
				return false, "already voted"
			}
		}
		v := Vote{
			VoterID:   voterID,
			VoterName: voterName,
			Support:   support,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		s.Proposals[i].Voters = append(s.Proposals[i].Voters, v)
		if support {
			s.Proposals[i].VotesFor++
		} else {
			s.Proposals[i].VotesAgainst++
		}
		s.save()
		return true, ""
	}
	return false, "proposal not found"
}


// Execute finalises a proposal after its deadline, marking it passed or rejected.
func (s *DAOStore) Execute(proposalID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.Proposals {
		if p.ID == proposalID && p.Status == "active" && time.Now().After(p.Deadline) {
			passed := p.VotesFor > p.VotesAgainst
			if passed {
				s.Proposals[i].Status = "passed"
				s.Proposals[i].ExecutedAt = time.Now()
			} else {
				s.Proposals[i].Status = "rejected"
			}
			s.save()
			return passed
		}
	}
	return false
}


// GetActive returns all proposals with "active" status.
func (s *DAOStore) GetActive() []Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Proposal
	for _, p := range s.Proposals {
		if p.Status == "active" {
			out = append(out, p)
		}
	}
	return out
}


// GetByStatus returns proposals filtered by status (active, passed, rejected, executed).
func (s *DAOStore) GetByStatus(status string) []Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Proposal
	for _, p := range s.Proposals {
		if p.Status == status {
			out = append(out, p)
		}
	}
	return out
}

func (s *DAOStore) checkDeadlines() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for i, p := range s.Proposals {
		if p.Status == "active" && now.After(p.Deadline) {
			passed := p.VotesFor > p.VotesAgainst
			if passed {
				s.Proposals[i].Status = "passed"
				s.Proposals[i].ExecutedAt = now
			} else {
				s.Proposals[i].Status = "rejected"
			}
			slog.Info("dao proposal deadline reached", "id", p.ID, "title", p.Title, "status", s.Proposals[i].Status)
		}
	}
	s.save()
}

// ── HTTP Handler ─────────────────────────────────────────────────────────────

var globalDAO *DAOStore


// InitDAO initialises the global DAO store and starts a deadline checker goroutine.
func InitDAO(dataDir string) {
	globalDAO = NewDAOStore(dataDir)
	// Start deadline checker cron
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			globalDAO.checkDeadlines()
		}
	}()
	slog.Info("dao initialized with deadline checker (interval: 1h)")
}


// DAOHandler serves the DAO governance API (list proposals, create, vote, execute).
func DAOHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if globalDAO == nil {
			writeJSON(w, map[string]any{"ok": false, "error": "dao not initialized"})
			return
		}

		switch r.Method {
		case "GET":
			status := r.URL.Query().Get("status")
			if status == "" {
				status = "active"
			}
			proposals := globalDAO.GetByStatus(status)
			writeJSON(w, map[string]any{
				"ok":        true,
				"proposals": proposals,
				"count":     len(proposals),
				"status":    status,
			})

		case "POST":
			var req struct {
				Action        string `json:"action"`
				Title         string `json:"title"`
				Description   string `json:"description"`
				AuthorID      string `json:"author_id"`
				AuthorName    string `json:"author_name"`
				ActionExec    string `json:"action_exec"`
				DurationHours int    `json:"duration_hours"`
				ProposalID    string `json:"proposal_id"`
				VoterID       string `json:"voter_id"`
				VoterName     string `json:"voter_name"`
				Support       bool   `json:"support"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad json", 400)
				return
			}

			switch req.Action {
			case "create":
				if req.Title == "" || req.AuthorID == "" {
					http.Error(w, "title and author_id required", 400)
					return
				}
				p := globalDAO.CreateProposal(req.Title, req.Description, req.AuthorID, req.AuthorName, req.ActionExec, req.DurationHours)
				writeJSON(w, map[string]any{"ok": true, "proposal": p})

			case "vote":
				if req.ProposalID == "" || req.VoterID == "" {
					http.Error(w, "proposal_id and voter_id required", 400)
					return
				}
				ok, msg := globalDAO.Vote(req.ProposalID, req.VoterID, req.VoterName, req.Support)
				if ok {
					writeJSON(w, map[string]any{"ok": true, "proposal_id": req.ProposalID, "support": req.Support})
				} else {
					writeJSON(w, map[string]any{"ok": false, "error": msg})
				}

			case "execute":
				if req.ProposalID == "" {
					http.Error(w, "proposal_id required", 400)
					return
				}
				passed := globalDAO.Execute(req.ProposalID)
				writeJSON(w, map[string]any{"ok": true, "proposal_id": req.ProposalID, "passed": passed})

			default:
				http.Error(w, fmt.Sprintf("unknown action: %s", req.Action), 400)
			}

		default:
			http.Error(w, "GET or POST", 400)
		}
	}
}
