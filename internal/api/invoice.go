// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Invoice represents a payment request sent to a SimpleX contact.
type Invoice struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	FromID      int64   `json:"from_id"`
	ToID        int64   `json:"to_id"`
	ChatID      string  `json:"chat_id"`
	CreatedAt   string  `json:"created_at"`
	PaidAt      string  `json:"paid_at,omitempty"`
}

// InvoiceManager manages creation, persistence, and querying of invoices.
type InvoiceManager struct {
	mu       sync.RWMutex
	invoices []Invoice
	filePath string
}

// GlobalInvoiceManager is the singleton invoice manager instance.
var GlobalInvoiceManager = NewInvoiceManager()

const maxInvoices = 1000


// NewInvoiceManager handles the NewInvoiceManager HTTP request.
// NewInvoiceManager creates a new InvoiceManager with an empty invoice buffer.
func NewInvoiceManager() *InvoiceManager {
	return &InvoiceManager{
		invoices: make([]Invoice, 0, maxInvoices),
	}
}


// WithFile handles the WithFile HTTP request.
// WithFile loads persisted invoices from a JSON file and enables auto-flush.
func (m *InvoiceManager) WithFile(path string) *InvoiceManager {
	m.filePath = path
	b, err := os.ReadFile(path)
	if err == nil {
		var invs []Invoice
		if json.Unmarshal(b, &invs) == nil && invs != nil {
			m.invoices = invs
		}
	}
	return m
}

func (m *InvoiceManager) flush() {
	if m.filePath == "" {
		return
	}
	b, _ := json.Marshal(m.invoices)
	os.WriteFile(m.filePath, b, 0600)
}


// Add handles the Add HTTP request.
// Add appends an invoice to the manager and persists to disk.
func (m *InvoiceManager) Add(inv Invoice) {
	m.mu.Lock()
	m.invoices = append(m.invoices, inv)
	if len(m.invoices) > maxInvoices {
		excess := len(m.invoices) - maxInvoices
		m.invoices = m.invoices[excess:]
	}
	m.mu.Unlock()
	m.flush()
}


// GetByID handles the GetByID HTTP request.
// GetByID returns the invoice with the given ID, or nil if not found.
func (m *InvoiceManager) GetByID(id string) *Invoice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := range m.invoices {
		if m.invoices[i].ID == id {
			return &m.invoices[i]
		}
	}
	return nil
}


// UpdateStatus handles the UpdateStatus HTTP request.
// UpdateStatus changes the status of an invoice and records the paid timestamp if applicable.
func (m *InvoiceManager) UpdateStatus(id, status string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.invoices {
		if m.invoices[i].ID == id {
			m.invoices[i].Status = status
			if status == "paid" {
				m.invoices[i].PaidAt = time.Now().UTC().Format(time.RFC3339)
			}
			m.flush()
			return true
		}
	}
	return false
}


// ExpireOld handles the ExpireOld HTTP request.
// ExpireOld marks pending invoices older than the given duration as expired.
func (m *InvoiceManager) ExpireOld(dur time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().UTC().Add(-dur)
	expired := 0
	for i := range m.invoices {
		if m.invoices[i].Status == "pending" {
			created, err := time.Parse(time.RFC3339, m.invoices[i].CreatedAt)
			if err == nil && created.Before(cutoff) {
				m.invoices[i].Status = "expired"
				expired++
			}
		}
	}
	if expired > 0 {
		m.flush()
	}
	return expired
}


// List handles the List HTTP request.
// List returns all invoices as a copy.
func (m *InvoiceManager) List() []Invoice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Invoice, len(m.invoices))
	copy(out, m.invoices)
	return out
}


// ListByContact handles the ListByContact HTTP request.
// ListByContact returns invoices filtered by chat/contact ID.
func (m *InvoiceManager) ListByContact(chatID string) []Invoice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Invoice
	for _, inv := range m.invoices {
		if inv.ChatID == chatID {
			out = append(out, inv)
		}
	}
	if out == nil {
		out = make([]Invoice, 0)
	}
	return out
}


// ListByStatus handles the ListByStatus HTTP request.
// ListByStatus returns invoices filtered by status (pending, paid, expired).
func (m *InvoiceManager) ListByStatus(status string) []Invoice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Invoice
	for _, inv := range m.invoices {
		if inv.Status == status {
			out = append(out, inv)
		}
	}
	if out == nil {
		out = make([]Invoice, 0)
	}
	return out
}


// Count handles the Count HTTP request.
// Count returns the total number of invoices tracked.
func (m *InvoiceManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.invoices)
}

var invoiceIDCounter int64

func nextInvoiceID() string {
	invoiceIDCounter++
	return fmt.Sprintf("inv-%d-%d", time.Now().Unix(), invoiceIDCounter)
}


// InvoiceCreateHandler handles the InvoiceCreateHandler HTTP request.
// InvoiceCreateHandler creates a new invoice and sends it via chat bridge to the contact.
func InvoiceCreateHandler(manager *InvoiceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		var req struct {
			Amount      float64 `json:"amount"`
			Currency    string  `json:"currency"`
			Description string  `json:"description"`
			ContactID   int64   `json:"contact_id"`
			ChatID      string  `json:"chat_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"ok": false, "error": "bad json"})
			return
		}
		if req.Amount <= 0 {
			writeJSON(w, map[string]any{"ok": false, "error": "amount must be positive"})
			return
		}
		if req.Currency == "" {
			req.Currency = "XAG"
		}

		inv := Invoice{
			ID:          nextInvoiceID(),
			Amount:      req.Amount,
			Currency:    req.Currency,
			Description: req.Description,
			Status:      "pending",
			FromID:      1,
			ToID:        req.ContactID,
			ChatID:      req.ChatID,
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		}
		manager.Add(inv)

		msgText := fmt.Sprintf("📄 INVOICE %s\nAmount: %.2f %s\nDescription: %s\nStatus: pending\nID: %s",
			inv.ID, inv.Amount, inv.Currency, inv.Description, inv.ID)
		chatMsg := ChatMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			From:      "admin",
			Text:      msgText,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			IsUser:    true,
			ChatID:    req.ChatID,
			Status:    StatusSending,
		}
		GlobalChatHub.AddMessage(chatMsg)

		if BridgeSendFunc != nil && req.ContactID > 0 {
			BridgeSendFunc(msgText, 1, req.ContactID)
		}

		writeJSON(w, map[string]any{"ok": true, "invoice": inv})
	}
}


// InvoiceListHandler handles the InvoiceListHandler HTTP request.
// InvoiceListHandler returns invoices, optionally filtered by chat_id or status.
func InvoiceListHandler(manager *InvoiceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := r.URL.Query().Get("chat_id")
		status := r.URL.Query().Get("status")

		var invoices []Invoice
		switch {
		case chatID != "":
			invoices = manager.ListByContact(chatID)
		case status != "":
			invoices = manager.ListByStatus(status)
		default:
			invoices = manager.List()
		}

		writeJSON(w, map[string]any{"ok": true, "invoices": invoices})
	}
}


// CountByStatus handles the CountByStatus HTTP request.
// CountByStatus returns the count of invoices with a given status.
func (m *InvoiceManager) CountByStatus(status string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, inv := range m.invoices {
		if inv.Status == status {
			count++
		}
	}
	return count
}


// InvoicePayHandler handles the InvoicePayHandler HTTP request.
// InvoicePayHandler marks a pending invoice as paid and sends a confirmation message.
func InvoicePayHandler(manager *InvoiceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"ok": false, "error": "bad json"})
			return
		}
		if req.ID == "" {
			writeJSON(w, map[string]any{"ok": false, "error": "id required"})
			return
		}

		inv := manager.GetByID(req.ID)
		if inv == nil {
			writeJSON(w, map[string]any{"ok": false, "error": "invoice not found"})
			return
		}
		if inv.Status != "pending" {
			writeJSON(w, map[string]any{"ok": false, "error": "invoice already "+inv.Status})
			return
		}

		manager.UpdateStatus(req.ID, "paid")

		paidText := fmt.Sprintf("✅ PAID: Invoice %s (%.2f %s)", req.ID, inv.Amount, inv.Currency)
		chatMsg := ChatMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			From:      "admin",
			Text:      paidText,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			IsUser:    true,
			ChatID:    inv.ChatID,
			Status:    StatusSent,
		}
		GlobalChatHub.AddMessage(chatMsg)

		if BridgeSendFunc != nil && inv.ToID > 0 {
			BridgeSendFunc(paidText, 1, inv.ToID)
		}

		writeJSON(w, map[string]any{"ok": true, "invoice": manager.GetByID(req.ID)})
	}
}


// InvoiceStatsHandler handles the InvoiceStatsHandler HTTP request.
// InvoiceStatsHandler returns aggregate invoice counts (total, pending, paid).
func InvoiceStatsHandler(manager *InvoiceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"ok":      true,
			"total":   manager.Count(),
			"pending": manager.CountByStatus("pending"),
			"paid":    manager.CountByStatus("paid"),
		})
	}
}


// InvoiceExportCSVHandler handles the InvoiceExportCSVHandler HTTP request.
// InvoiceExportCSVHandler exports invoices as a CSV file download.
func InvoiceExportCSVHandler(manager *InvoiceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		manager.mu.RLock()
		var filtered []Invoice
		for _, inv := range manager.invoices {
			if status == "" || inv.Status == status {
				filtered = append(filtered, inv)
			}
		}
		manager.mu.RUnlock()

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=invoices.csv")
		w.Write([]byte("ID,Amount,Currency,Description,Status,FromID,ToID,ChatID,CreatedAt,PaidAt\n"))
		for _, inv := range filtered {
			paidAt := inv.PaidAt
			if paidAt == "" {
				paidAt = "N/A"
			}
			desc := strings.ReplaceAll(inv.Description, "\"", "\"\"")
			line := fmt.Sprintf("%s,%.2f,%s,\"%s\",%s,%d,%d,%s,%s,%s\n",
				inv.ID, inv.Amount, inv.Currency, desc, inv.Status,
				inv.FromID, inv.ToID, inv.ChatID, inv.CreatedAt, paidAt)
			w.Write([]byte(line))
		}
	}
}
