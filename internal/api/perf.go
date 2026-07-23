// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"log/slog"
	"net/http"
	"sort"
	"sync"
	"time"
)

type trackedResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
	path   string
}


// WriteHeader implements http.ResponseWriter to track the status code.
func (w *trackedResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}


// Write handles the Write HTTP request.
func (w *trackedResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// ── Performance Profiling (C6) ───────────────────────────────────────────────

// PerfStats tracks request timing metrics across API endpoints.
type PerfStats struct {
	mu       sync.RWMutex
	entries  map[string]*PerfEntry
	slowLog  []SlowRequest
	slowMu   sync.Mutex
}

// PerfEntry holds timing statistics for a single API endpoint.
type PerfEntry struct {
	Count      int           `json:"count"`
	TotalTime  time.Duration `json:"total_time_ns"`
	AvgTime    time.Duration `json:"avg_time_ns"`
	MaxTime    time.Duration `json:"max_time_ns"`
	MinTime    time.Duration `json:"min_time_ns"`
	LastTime   time.Duration `json:"last_time_ns"`
}

// SlowRequest records a request that exceeded the slow threshold (500ms).
type SlowRequest struct {
	Path      string `json:"path"`
	Method    string `json:"method"`
	Duration  int64  `json:"duration_ms"`
	Timestamp string `json:"timestamp"`
}

var globalPerf = &PerfStats{entries: make(map[string]*PerfEntry)}

// PerfMiddleware wraps an http.Handler to track request timing and byte counts.
func PerfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tw := &trackedResponseWriter{ResponseWriter: w, path: r.URL.Path}
		next.ServeHTTP(tw, r)
		dur := time.Since(start)
		key := r.Method + " " + r.URL.Path

		globalPerf.mu.Lock()
		e, ok := globalPerf.entries[key]
		if !ok {
			e = &PerfEntry{MinTime: dur}
			globalPerf.entries[key] = e
		}
		e.Count++
		e.TotalTime += dur
		e.LastTime = dur
		if dur > e.MaxTime { e.MaxTime = dur }
		if dur < e.MinTime { e.MinTime = dur }
		e.AvgTime = e.TotalTime / time.Duration(e.Count)
		globalPerf.mu.Unlock()

		if tw.bytes > 0 {
			TrackBytes(key, tw.bytes)
		}

		if dur > 500*time.Millisecond {
			globalPerf.slowMu.Lock()
			globalPerf.slowLog = append(globalPerf.slowLog, SlowRequest{
				Path:      key,
				Duration:  dur.Milliseconds(),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			})
			if len(globalPerf.slowLog) > 100 {
				globalPerf.slowLog = globalPerf.slowLog[len(globalPerf.slowLog)-100:]
			}
			globalPerf.slowMu.Unlock()
			slog.Warn("slow request", "path", key, "duration_ms", dur.Milliseconds())
		}
	})
}

func (p *PerfStats) reset() {
	p.mu.Lock()
	p.entries = make(map[string]*PerfEntry)
	p.mu.Unlock()
	p.slowMu.Lock()
	p.slowLog = nil
	p.slowMu.Unlock()
}


// PerfStatsHandler returns endpoint timing statistics and slow request log.
func PerfStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		globalPerf.mu.RLock()
		var entries []map[string]any
		for path, e := range globalPerf.entries {
			entries = append(entries, map[string]any{
				"path":      path,
				"count":     e.Count,
				"avg_ms":    e.AvgTime.Milliseconds(),
				"max_ms":    e.MaxTime.Milliseconds(),
				"min_ms":    e.MinTime.Milliseconds(),
				"last_ms":   e.LastTime.Milliseconds(),
			})
		}
		globalPerf.mu.RUnlock()

		sort.Slice(entries, func(i, j int) bool {
			return entries[i]["avg_ms"].(int64) > entries[j]["avg_ms"].(int64)
		})

		globalPerf.slowMu.Lock()
		slow := make([]SlowRequest, len(globalPerf.slowLog))
		copy(slow, globalPerf.slowLog)
		globalPerf.slowMu.Unlock()

		writeJSON(w, map[string]any{
			"ok":             true,
			"endpoints":      entries,
			"total_tracked":  len(entries),
			"slow_requests":  slow,
			"slow_count":     len(slow),
		})
	}
}


// PerfStatsResetHandler clears all accumulated performance statistics.
func PerfStatsResetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		globalPerf.reset()
		writeJSON(w, map[string]any{"ok": true})
	}
}
