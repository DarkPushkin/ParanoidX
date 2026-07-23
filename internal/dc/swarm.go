// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type HealStats struct {
	TotalRepairs     int            `json:"total_repairs"`
	HashMismatches   int            `json:"hash_mismatches"`
	LastRepairTime   string         `json:"last_repair_time"`
	LastMismatchInfo string         `json:"last_mismatch_info,omitempty"`
	ByInfohash       map[string]int `json:"by_infohash"`
}

type SwarmManager struct {
	mu             sync.RWMutex
	cloud          *Cloud
	replicationMin int
	checkInterval  time.Duration
	healThreshold  int
	healStats      HealStats
	statsPath      string
}


// NewSwarmManager handles the NewSwarmManager HTTP request.
func NewSwarmManager(cloud *Cloud) *SwarmManager {
	sm := &SwarmManager{
		cloud:          cloud,
		replicationMin: DefaultReplication,
		checkInterval:  120 * time.Second,
		healThreshold:  2,
		healStats: HealStats{
			ByInfohash: make(map[string]int),
		},
	}
	sm.statsPath = filepath.Join(cloud.dataDir, "heal_stats.json")
	sm.loadStats()
	return sm
}

type HealAction struct {
	Infohash   string
	Replicas   int
	TargetRep  int
	MissingRep int
	Seeders    []string
}


// NeedHealing handles the NeedHealing HTTP request.
func (sm *SwarmManager) NeedHealing() []HealAction {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	containers := sm.cloud.ListContainers()
	var actions []HealAction
	for _, m := range containers {
		if m.Replicas < m.TargetRep {
			actions = append(actions, HealAction{
				Infohash:   m.Infohash,
				Replicas:   m.Replicas,
				TargetRep:  m.TargetRep,
				MissingRep: m.TargetRep - m.Replicas,
				Seeders:    m.Seeders,
			})
		}
	}
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].MissingRep > actions[j].MissingRep
	})
	return actions
}


// SetReplicationFactor handles the SetReplicationFactor HTTP request.
func (sm *SwarmManager) SetReplicationFactor(n int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.replicationMin = n
	sm.cloud.mu.Lock()
	for _, m := range sm.cloud.containers {
		m.TargetRep = n
	}
	sm.cloud.mu.Unlock()
}


// Run handles the Run HTTP request.
func (sm *SwarmManager) Run() {
	ticker := time.NewTicker(sm.checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-sm.cloud.stopCh:
			return
		case <-ticker.C:
			actions := sm.NeedHealing()
			for _, a := range actions {
				sm.healContainer(a)
			}
			sm.pruneStalePeers()
		}
	}
}

func (sm *SwarmManager) saveStats() {
	data, err := json.Marshal(sm.healStats)
	if err != nil {
		return
	}
	tmp := sm.statsPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}
	os.Rename(tmp, sm.statsPath)
}

func (sm *SwarmManager) loadStats() {
	data, err := os.ReadFile(sm.statsPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &sm.healStats)
	if sm.healStats.ByInfohash == nil {
		sm.healStats.ByInfohash = make(map[string]int)
	}
}

// GetHealStats returns a copy of the heal statistics.
func (sm *SwarmManager) GetHealStats() HealStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.healStats
}

func (sm *SwarmManager) healContainer(a HealAction) {
	slog.Info("dc healing container",
		"infohash", a.Infohash,
		"replicas", a.Replicas,
		"target", a.TargetRep,
		"missing", a.MissingRep,
		"seeders", a.Seeders,
	)

	// Verify piece integrity of seeded pieces
	manifestPath := filepath.Join(ManifestPath(sm.cloud.dataDir), a.Infohash+".dc")
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		slog.Warn("dc heal: cannot load manifest", "infohash", a.Infohash, "error", err)
		return
	}

	verified := true
	for i, pieceHash := range manifest.Pieces {
		piecePath := filepath.Join(PieceCachePath(sm.cloud.dataDir), pieceHash)
		data, err := os.ReadFile(piecePath)
		if err != nil {
			slog.Warn("dc heal: missing piece", "infohash", a.Infohash, "piece", i)
			verified = false
			continue
		}
		if !manifest.VerifyPiece(i, data) {
			slog.Warn("dc heal: piece hash mismatch",
				"infohash", a.Infohash,
				"piece", i,
				"expected", pieceHash,
			)
			sm.mu.Lock()
			sm.healStats.HashMismatches++
			sm.healStats.LastMismatchInfo = fmt.Sprintf("infohash=%s piece=%d expected=%s", a.Infohash, i, pieceHash)
			sm.mu.Unlock()
			verified = false
		}
	}

	sm.mu.Lock()
	sm.healStats.TotalRepairs++
	sm.healStats.LastRepairTime = time.Now().UTC().Format(time.RFC3339)
	sm.healStats.ByInfohash[a.Infohash]++
	sm.mu.Unlock()
	sm.saveStats()

	if !verified {
		slog.Warn("dc heal: container has integrity issues", "infohash", a.Infohash)
	}
}

func (sm *SwarmManager) pruneStalePeers() {
	sm.cloud.mu.Lock()
	defer sm.cloud.mu.Unlock()
	for _, m := range sm.cloud.containers {
		if time.Now().Unix()-m.Updated > 300 {
			m.Seeders = nil
			m.Leechers = nil
			m.Replicas = 0
		}
	}
}
