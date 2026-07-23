// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// EndpointLimit defines per-endpoint rate limit parameters (req/sec and burst).
type EndpointLimit struct {
	RequestsPerSec int `json:"requests_per_sec"`
	Burst          int `json:"burst"`
}

// RateLimitConfig holds global and per-endpoint rate limit configurations.
type RateLimitConfig struct {
	Global    EndpointLimit              `json:"global"`
	Endpoints map[string]EndpointLimit   `json:"endpoints"`
}

// PerEndpointRateLimiter manages rate limiters for each API endpoint with a configurable global default.
type PerEndpointRateLimiter struct {
	mu       sync.RWMutex
	config   RateLimitConfig
	filePath string
	limiters map[string]*tokenBucket
}

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastCheck  time.Time
}

func newTokenBucket(ratePerSec int, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: float64(ratePerSec),
		lastCheck:  time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(tb.lastCheck).Seconds()
	tb.lastCheck = now
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

var GlobalPerEndpointLimiter *PerEndpointRateLimiter


// NewPerEndpointRateLimiter creates a rate limiter manager and loads config from disk.
func NewPerEndpointRateLimiter(dataDir string) *PerEndpointRateLimiter {
	rl := &PerEndpointRateLimiter{
		filePath: filepath.Join(dataDir, "ratelimit_config.json"),
		config: RateLimitConfig{
			Global: EndpointLimit{RequestsPerSec: 10, Burst: 20},
			Endpoints: map[string]EndpointLimit{
				"/api/chat/send":   {RequestsPerSec: 2, Burst: 5},
				"/api/chat/stream": {RequestsPerSec: 5, Burst: 10},
				"/api/admin/*":     {RequestsPerSec: 20, Burst: 50},
			},
		},
		limiters: make(map[string]*tokenBucket),
	}
	rl.load()
	rl.buildLimiters()
	return rl
}

func (rl *PerEndpointRateLimiter) load() {
	b, err := os.ReadFile(rl.filePath)
	if err != nil {
		return
	}
	var cfg RateLimitConfig
	if json.Unmarshal(b, &cfg) == nil {
		rl.config = cfg
	}
}

func (rl *PerEndpointRateLimiter) save() {
	b, _ := json.MarshalIndent(rl.config, "", "  ")
	os.WriteFile(rl.filePath, b, 0600)
}

func (rl *PerEndpointRateLimiter) buildLimiters() {
	rl.limiters["__global__"] = newTokenBucket(rl.config.Global.RequestsPerSec, rl.config.Global.Burst)
	for path, limit := range rl.config.Endpoints {
		rl.limiters[path] = newTokenBucket(limit.RequestsPerSec, limit.Burst)
	}
}


// Allow handles the Allow HTTP request.
// Allow checks whether a request to the given path is permitted under the current rate limit.
func (rl *PerEndpointRateLimiter) Allow(path string) bool {
	rl.mu.RLock()
	limiter, exact := rl.limiters[path]
	if !exact {
		for pattern, l := range rl.limiters {
			if strings.HasSuffix(pattern, "/*") {
				prefix := strings.TrimSuffix(pattern, "/*")
				if strings.HasPrefix(path, prefix) {
					limiter = l
					break
				}
			}
		}
	}
	global := rl.limiters["__global__"]
	rl.mu.RUnlock()
	if limiter != nil && !limiter.allow() {
		return false
	}
	if global != nil && !global.allow() {
		return false
	}
	return true
}


// GetConfig handles the GetConfig HTTP request.
// GetConfig returns the current rate limit configuration.
func (rl *PerEndpointRateLimiter) GetConfig() RateLimitConfig {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.config
}


// UpdateConfig handles the UpdateConfig HTTP request.
// UpdateConfig replaces the rate limit configuration and rebuilds limiters.
func (rl *PerEndpointRateLimiter) UpdateConfig(cfg RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.config = cfg
	rl.limiters = make(map[string]*tokenBucket)
	rl.buildLimiters()
	rl.save()
}


// RateLimitConfigHandlerV2 returns or updates the global rate limit configuration (GET/POST).
func RateLimitConfigHandlerV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			cfg := GlobalPerEndpointLimiter.GetConfig()
			writeJSON(w, map[string]any{"ok": true, "config": cfg})
		case "PUT":
			var req RateLimitConfig
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, map[string]any{"ok": false, "error": "bad json"})
				return
			}
			if req.Global.RequestsPerSec <= 0 {
				req.Global.RequestsPerSec = 10
			}
			if req.Global.Burst <= 0 {
				req.Global.Burst = 20
			}
			GlobalPerEndpointLimiter.UpdateConfig(req)
			writeJSON(w, map[string]any{"ok": true, "config": GlobalPerEndpointLimiter.GetConfig()})
		default:
			http.Error(w, "GET/PUT", 405)
		}
	}
}
