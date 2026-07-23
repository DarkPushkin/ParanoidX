// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DiskSnapshot records disk usage at a single point in time.
type DiskSnapshot struct {
	Timestamp  string  `json:"timestamp"`
	UsedPct    float64 `json:"used_pct"`
	UsedGB     float64 `json:"used_gb"`
	TotalGB    float64 `json:"total_gb"`
}

// DiskTrend tracks disk usage snapshots over time for trend analysis.
type DiskTrend struct {
	mu        sync.RWMutex
	dataDir   string
	snapshots []DiskSnapshot
	maxSnap   int
}

var GlobalDiskTrend *DiskTrend


// InitDiskTrend creates a DiskTrend, loads persisted snapshots, and sets the global instance.
func InitDiskTrend(dataDir string) *DiskTrend {
	dt := &DiskTrend{
		dataDir:   dataDir,
		snapshots: make([]DiskSnapshot, 0, 288),
		maxSnap:   288,
	}
	dt.load()
	GlobalDiskTrend = dt
	return dt
}

func (dt *DiskTrend) trendPath() string {
	return filepath.Join(dt.dataDir, "disk_trend.json")
}

func (dt *DiskTrend) load() {
	data, err := os.ReadFile(dt.trendPath())
	if err != nil {
		return
	}
	var snaps []DiskSnapshot
	if err := json.Unmarshal(data, &snaps); err != nil {
		return
	}
	dt.snapshots = snaps
}

func (dt *DiskTrend) save() {
	data, err := json.Marshal(dt.snapshots)
	if err != nil {
		return
	}
	tmp := dt.trendPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}
	os.Rename(tmp, dt.trendPath())
}


// Record takes a disk usage snapshot and appends it to the trend history.
func (dt *DiskTrend) Record() {
	usedPct, usedGB, totalGB := getDiskUsage()
	if usedPct < 0 {
		return
	}
	dt.mu.Lock()
	defer dt.mu.Unlock()
	snap := DiskSnapshot{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		UsedPct:   usedPct,
		UsedGB:    usedGB,
		TotalGB:   totalGB,
	}
	dt.snapshots = append(dt.snapshots, snap)
	if len(dt.snapshots) > dt.maxSnap {
		dt.snapshots = dt.snapshots[len(dt.snapshots)-dt.maxSnap:]
	}
	dt.save()
}


// GetTrend returns disk usage snapshots within the specified number of hours.
func (dt *DiskTrend) GetTrend(hours int) []DiskSnapshot {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	var out []DiskSnapshot
	for _, s := range dt.snapshots {
		t, err := time.Parse(time.RFC3339, s.Timestamp)
		if err == nil && t.After(cutoff) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp < out[j].Timestamp
	})
	return out
}


// StartRecording begins periodic disk usage snapshot recording in a background goroutine.
func (dt *DiskTrend) StartRecording(interval time.Duration) {
	go func() {
		dt.Record()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			dt.Record()
		}
	}()
}

func getDiskUsage() (usedPct float64, usedGB float64, totalGB float64) {
	out, err := exec.Command("df", "-B1", "/").Output()
	if err != nil {
		return -1, 0, 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return -1, 0, 0
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return -1, 0, 0
	}
	total, _ := strconv.ParseFloat(fields[1], 64)
	used, _ := strconv.ParseFloat(fields[2], 64)
	pctStr := strings.TrimSuffix(fields[4], "%")
	pct, _ := strconv.ParseFloat(pctStr, 64)
	if total <= 0 {
		return -1, 0, 0
	}
	return pct, used / (1 << 30), total / (1 << 30)
}


// DiskTrendHandler returns disk usage trend data with min/max/avg statistics.
func DiskTrendHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hours := 24
		if h := r.URL.Query().Get("hours"); h != "" {
			if v, err := strconv.Atoi(h); err == nil && v > 0 && v <= 168 {
				hours = v
			}
		}
		if GlobalDiskTrend == nil {
			writeJSON(w, map[string]any{"error": "disk trend not initialized"})
			return
		}
		snapshots := GlobalDiskTrend.GetTrend(hours)
		current := DiskSnapshot{}
		if len(snapshots) > 0 {
			current = snapshots[len(snapshots)-1]
		}
		var minPct, maxPct, avgPct float64
		if len(snapshots) > 0 {
			minPct = 100
			var sum float64
			for _, s := range snapshots {
				if s.UsedPct < minPct {
					minPct = s.UsedPct
				}
				if s.UsedPct > maxPct {
					maxPct = s.UsedPct
				}
				sum += s.UsedPct
			}
			avgPct = sum / float64(len(snapshots))
		}
		writeJSON(w, map[string]any{
			"current":   current,
			"min_pct":   fmt.Sprintf("%.1f", minPct),
			"max_pct":   fmt.Sprintf("%.1f", maxPct),
			"avg_pct":   fmt.Sprintf("%.1f", avgPct),
			"samples":   len(snapshots),
			"hours":     hours,
			"snapshots": snapshots,
		})
	}
}
