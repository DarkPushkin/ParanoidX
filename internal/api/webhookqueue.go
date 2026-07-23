// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

// WebhookDelivery represents a single queued webhook delivery attempt with retry state.
type WebhookDelivery struct {
	ID        string `json:"id"`
	Event     string `json:"event"`
	Payload   any    `json:"payload"`
	URL       string `json:"url"`
	Secret    string `json:"secret,omitempty"`
	Retries   int    `json:"retries"`
	MaxRetry  int    `json:"max_retry"`
	Delay     int    `json:"delay"`
	CreatedAt string `json:"created_at"`
	LastTry   string `json:"last_try,omitempty"`
	Status    string `json:"status"` // pending, delivered, failed, dead
	Error     string `json:"error,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
}

// WebhookQueueStats provides aggregate delivery metrics for the webhook queue.
type WebhookQueueStats struct {
	Pending    int     `json:"pending"`
	Delivered  int     `json:"delivered"`
	Failed     int     `json:"failed"`
	Dead       int     `json:"dead"`
	Total      int     `json:"total"`
	AvgLatency float64 `json:"avg_latency_ms"`
}

// WebhookQueue manages persistent delivery of webhook events with retry and HMAC signing.
type WebhookQueue struct {
	mu       sync.Mutex
	entries  []WebhookDelivery
	filePath string
	seq      int64
}

// GlobalWebhookQueue is the singleton webhook delivery queue instance.
var GlobalWebhookQueue *WebhookQueue


// NewWebhookQueue creates a WebhookQueue, loads persisted entries, and starts the delivery loop.
func NewWebhookQueue(filePath string) *WebhookQueue {
	q := &WebhookQueue{
		filePath: filePath,
	}
	q.load()
	go q.deliveryLoop()
	return q
}

func (q *WebhookQueue) load() {
	b, err := os.ReadFile(q.filePath)
	if err != nil {
		q.entries = make([]WebhookDelivery, 0)
		return
	}
	json.Unmarshal(b, &q.entries)
	if q.entries == nil {
		q.entries = make([]WebhookDelivery, 0)
	}
	for _, e := range q.entries {
		var id int64
		if _, err := fmt.Sscanf(e.ID, "whq-%d", &id); err == nil && id > q.seq {
			q.seq = id
		}
	}
}

func (q *WebhookQueue) flush() {
	b, _ := json.Marshal(q.entries)
	tmp := q.filePath + ".tmp"
	os.WriteFile(tmp, b, 0600)
	os.Rename(tmp, q.filePath)
}


// Enqueue adds a new webhook delivery to the queue.
func (q *WebhookQueue) Enqueue(event string, payload any, url, secret string, maxRetry, delay int) *WebhookDelivery {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.seq++
	d := &WebhookDelivery{
		ID:        fmt.Sprintf("whq-%d", q.seq),
		Event:     event,
		Payload:   payload,
		URL:       url,
		Secret:    secret,
		Retries:   0,
		MaxRetry:  maxRetry,
		Delay:     delay,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Status:    "pending",
	}
	q.entries = append(q.entries, *d)
	q.flush()
	return d
}


// List returns webhook deliveries with a given limit (newest first).
func (q *WebhookQueue) List(limit int) []WebhookDelivery {
	q.mu.Lock()
	defer q.mu.Unlock()
	if limit <= 0 || limit > len(q.entries) {
		limit = len(q.entries)
	}
	out := make([]WebhookDelivery, limit)
	copy(out, q.entries[len(q.entries)-limit:])
	return out
}


// History returns delivered webhook entries, newest first, up to the given limit.
func (q *WebhookQueue) History(limit int) []WebhookDelivery {
	q.mu.Lock()
	defer q.mu.Unlock()
	entries := make([]WebhookDelivery, 0, len(q.entries))
	for _, e := range q.entries {
		if e.Status != "pending" {
			entries = append(entries, e)
		}
	}
	if limit <= 0 || limit > len(entries) {
		limit = len(entries)
	}
	if len(entries) == 0 {
		return []WebhookDelivery{}
	}
	out := make([]WebhookDelivery, limit)
	copy(out, entries[len(entries)-limit:])
	return out
}


// PendingCount returns the number of entries awaiting delivery.
func (q *WebhookQueue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	count := 0
	for _, e := range q.entries {
		if e.Status == "pending" {
			count++
		}
	}
	return count
}


// Stats returns aggregate delivery metrics (pending, delivered, failed, dead, avg latency).
func (q *WebhookQueue) Stats() WebhookQueueStats {
	q.mu.Lock()
	defer q.mu.Unlock()
	var s WebhookQueueStats
	var totalLatency int64
	var latencyCount int64
	for _, e := range q.entries {
		s.Total++
		switch e.Status {
		case "pending":
			s.Pending++
		case "delivered":
			s.Delivered++
			if e.LatencyMs > 0 {
				totalLatency += e.LatencyMs
				latencyCount++
			}
		case "failed":
			s.Failed++
		case "dead":
			s.Dead++
		}
	}
	if latencyCount > 0 {
		s.AvgLatency = float64(totalLatency) / float64(latencyCount)
	}
	return s
}


// RetryDead retries all dead-letter deliveries, returning the count retried.
func (q *WebhookQueue) RetryDead() int {
	q.mu.Lock()
	count := 0
	for i, e := range q.entries {
		if e.Status == "dead" {
			q.entries[i].Status = "pending"
			q.entries[i].Retries = 0
			q.entries[i].Error = ""
			q.entries[i].LastTry = ""
			count++
		}
	}
	q.flush()
	q.mu.Unlock()
	slog.Info("webhook queue: retried dead webhooks", "count", count)
	return count
}

func (q *WebhookQueue) deliveryLoop() {
	for {
		time.Sleep(10 * time.Second)
		q.processPending()
	}
}

func backoffDelay(retries int) time.Duration {
	delays := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		32 * time.Second,
	}
	if retries < len(delays) {
		return delays[retries]
	}
	return 32 * time.Second
}

func (q *WebhookQueue) processPending() {
	q.mu.Lock()
	pending := make([]int, 0)
	for i, e := range q.entries {
		if e.Status != "pending" {
			continue
		}
		if e.LastTry != "" {
			lastTry, err := time.Parse(time.RFC3339, e.LastTry)
			if err != nil {
				continue
			}
			wait := time.Duration(e.Delay) * time.Second
			if e.Retries > 0 {
				wait = backoffDelay(e.Retries - 1)
			}
			if time.Since(lastTry) < wait {
				continue
			}
		}
		pending = append(pending, i)
	}
	q.mu.Unlock()

	for _, idx := range pending {
		q.mu.Lock()
		e := q.entries[idx]
		q.mu.Unlock()

		start := time.Now()
		err := q.deliver(e)
		latency := time.Since(start).Milliseconds()

		q.mu.Lock()
		e.LastTry = time.Now().UTC().Format(time.RFC3339)
		e.LatencyMs = latency
		if err != nil {
			e.Retries++
			e.Error = err.Error()
			if e.Retries >= 5 {
				e.Status = "dead"
				slog.Warn("webhook delivery marked dead after 5 failures", "id", e.ID, "event", e.Event)
				q.saveDeadLetter(e)
			} else if e.Retries >= e.MaxRetry {
				e.Status = "failed"
				slog.Warn("webhook delivery failed permanently", "id", e.ID, "event", e.Event, "retries", e.Retries, "error", err)
			} else {
				e.Status = "pending"
				nextDelay := backoffDelay(e.Retries - 1)
				slog.Warn("webhook delivery failed, will retry", "id", e.ID, "event", e.Event, "retry", e.Retries, "backoff", nextDelay.String(), "error", err)
			}
		} else {
			e.Status = "delivered"
			slog.Info("webhook delivered", "id", e.ID, "event", e.Event, "latency_ms", latency)
		}
		q.entries[idx] = e
		q.flush()
		q.mu.Unlock()
	}
}

func (q *WebhookQueue) saveDeadLetter(e WebhookDelivery) {
	dlPath := q.filePath + ".dead"
	var dead []WebhookDelivery
	if b, err := os.ReadFile(dlPath); err == nil {
		json.Unmarshal(b, &dead)
	}
	if dead == nil {
		dead = make([]WebhookDelivery, 0)
	}
	dead = append(dead, e)
	b, _ := json.Marshal(dead)
	os.WriteFile(dlPath, b, 0600)
}


// ListDead returns all webhook deliveries that have permanently failed.
func (q *WebhookQueue) ListDead() []WebhookDelivery {
	dlPath := q.filePath + ".dead"
	q.mu.Lock()
	defer q.mu.Unlock()
	var dead []WebhookDelivery
	if b, err := os.ReadFile(dlPath); err == nil {
		json.Unmarshal(b, &dead)
	}
	if dead == nil {
		return []WebhookDelivery{}
	}
	return dead
}


// RetryAllDead re-enqueues all dead-letter deliveries for retry.
func (q *WebhookQueue) RetryAllDead() int {
	dlPath := q.filePath + ".dead"
	q.mu.Lock()
	var dead []WebhookDelivery
	if b, err := os.ReadFile(dlPath); err == nil {
		json.Unmarshal(b, &dead)
	}
	if dead == nil {
		q.mu.Unlock()
		return 0
	}
	os.Remove(dlPath)
	q.mu.Unlock()

	count := 0
	for _, e := range dead {
		if e.URL == "" {
			continue
		}
		q.Enqueue(e.Event, e.Payload, e.URL, e.Secret, e.MaxRetry, e.Delay)
		count++
	}
	return count
}

func (q *WebhookQueue) deliver(e WebhookDelivery) error {
	body, err := json.Marshal(map[string]any{
		"event":   e.Event,
		"payload": e.Payload,
		"id":      e.ID,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequest("POST", e.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	now := time.Now()
	if e.Secret != "" {
		mac := hmac.New(sha256.New, []byte(e.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature", "sha256="+sig)
	}
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", now.Unix()))
	req.Header.Set("X-Webhook-ID", e.ID)
	req.Header.Set("User-Agent", "simplex-node-webhook/1.0")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	return nil
}


// WebhookQueueHandler manages the webhook queue (GET list/pending/delivered/failed, DELETE clear).
func WebhookQueueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			limitStr := r.URL.Query().Get("limit")
			limit := 20
			fmt.Sscanf(limitStr, "%d", &limit)
			if r.URL.Query().Get("history") == "true" {
				writeJSON(w, map[string]any{"ok": true, "deliveries": GlobalWebhookQueue.History(limit), "pending": GlobalWebhookQueue.PendingCount()})
			} else {
				writeJSON(w, map[string]any{"ok": true, "deliveries": GlobalWebhookQueue.List(limit), "pending": GlobalWebhookQueue.PendingCount()})
			}
		case "POST":
			var req struct {
				Event    string `json:"event"`
				Payload  any    `json:"payload"`
				URL      string `json:"url"`
				Secret   string `json:"secret"`
				Retries  int    `json:"retries"`
				Delay    int    `json:"delay"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, map[string]any{"ok": false, "error": "bad json"})
				return
			}
			if req.Event == "" || req.URL == "" {
				writeJSON(w, map[string]any{"ok": false, "error": "event and url required"})
				return
			}
			if req.Retries <= 0 {
				req.Retries = 10
			}
			if req.Delay <= 0 {
				req.Delay = 30
			}
			d := GlobalWebhookQueue.Enqueue(req.Event, req.Payload, req.URL, req.Secret, req.Retries, req.Delay)
			writeJSON(w, map[string]any{"ok": true, "delivery": d})
		case "DELETE":
			GlobalWebhookQueue.mu.Lock()
			GlobalWebhookQueue.entries = make([]WebhookDelivery, 0)
			GlobalWebhookQueue.mu.Unlock()
			GlobalWebhookQueue.flush()
			writeJSON(w, map[string]any{"ok": true, "cleared": true})
		default:
			http.Error(w, "GET/POST/DELETE", 405)
		}
	}
}


// WebhookQueueStatsHandler returns delivery metrics (pending/delivered/failed/dead counts and avg latency).
func WebhookQueueStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := GlobalWebhookQueue.Stats()
		writeJSON(w, map[string]any{"ok": true, "stats": stats})
	}
}


// WebhookQueueRetryDeadHandler retries a specific dead-letter delivery by ID.
func WebhookQueueRetryDeadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		count := GlobalWebhookQueue.RetryDead()
		writeJSON(w, map[string]any{"ok": true, "retried": count})
	}
}


// WebhookQueueDeadHandler lists all dead-letter webhook deliveries.
func WebhookQueueDeadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dead := GlobalWebhookQueue.ListDead()
		writeJSON(w, map[string]any{"ok": true, "dead": dead, "count": len(dead)})
	}
}


// WebhookQueueRetryAllHandler re-enqueues all dead-letter deliveries for retry.
func WebhookQueueRetryAllHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		count := GlobalWebhookQueue.RetryAllDead()
		writeJSON(w, map[string]any{"ok": true, "retried": count})
	}
}
