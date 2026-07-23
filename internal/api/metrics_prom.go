// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var (
	promStartTime = time.Now()
)

// per-endpoint byte counters
var (
	bytesMu    sync.RWMutex
	bytesTotal = map[string]int64{} // path → total bytes
)

// TrackBytes records response bytes for a given path.
func TrackBytes(path string, n int) {
	bytesMu.Lock()
	bytesTotal[path] += int64(n)
	bytesMu.Unlock()
}


// PrometheusMetricsHandler handles the PrometheusMetricsHandler HTTP request.
func PrometheusMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		uptime := time.Since(promStartTime).Hours()

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		msgCount := 0
		sseCount := 0
		if GlobalChatHub != nil {
			msgCount = GlobalChatHub.MessageCount()
			sseCount = GlobalChatHub.SSEClientCount()
		}

		bridgeConnected := 0
		if BridgeConnected {
			bridgeConnected = 1
		}

		lines := []string{
			"# HELP simplex_uptime_hours Server uptime in hours",
			"# TYPE simplex_uptime_hours gauge",
			fmt.Sprintf("simplex_uptime_hours %v", uptime),
			"",
			"# HELP simplex_messages_total Total messages stored",
			"# TYPE simplex_messages_total gauge",
			fmt.Sprintf("simplex_messages_total %d", msgCount),
			"",
			"# HELP simplex_bridge_connected Bridge connection status (1=connected)",
			"# TYPE simplex_bridge_connected gauge",
			fmt.Sprintf("simplex_bridge_connected %d", bridgeConnected),
			"",
			"# HELP simplex_goroutines Current number of goroutines",
			"# TYPE simplex_goroutines gauge",
			fmt.Sprintf("simplex_goroutines %d", runtime.NumGoroutine()),
			"",
			"# HELP simplex_memory_bytes Current memory usage in bytes",
			"# TYPE simplex_memory_bytes gauge",
			fmt.Sprintf("simplex_memory_alloc_bytes %d", memStats.Alloc),
			fmt.Sprintf("simplex_memory_sys_bytes %d", memStats.Sys),
			"",
			"# HELP simplex_sse_clients Current SSE client count",
			"# TYPE simplex_sse_clients gauge",
			fmt.Sprintf("simplex_sse_clients %d", sseCount),
			"",
		}

		globalPerf.mu.RLock()
		for path, e := range globalPerf.entries {
			lines = append(lines,
				fmt.Sprintf("simplex_http_requests_total{path=%q} %d", path, e.Count),
				fmt.Sprintf("simplex_http_request_duration_avg_ms{path=%q} %d", path, e.AvgTime.Milliseconds()),
			)
		}
		globalPerf.mu.RUnlock()

		bytesMu.RLock()
		for path, n := range bytesTotal {
			lines = append(lines,
				fmt.Sprintf("simplex_http_bytes_total{path=%q} %d", path, n),
			)
		}
		bytesMu.RUnlock()

		for _, l := range lines {
			fmt.Fprintln(w, l)
		}
	}
}
