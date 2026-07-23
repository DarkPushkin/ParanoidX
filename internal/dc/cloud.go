// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	ContainerPieceSize = 256 * 1024
	DefaultReplication = 3
	ManifestDir        = "manifests"
	ContainerDir       = "containers"
	PieceCacheDir      = "pieces"
)

type ContainerMeta struct {
	Infohash   string   `json:"infohash"`
	Size       int64    `json:"size"`
	PieceCount int      `json:"piece_count"`
	Seeders    []string `json:"seeders,omitempty"`
	Leechers   []string `json:"leechers,omitempty"`
	Replicas   int      `json:"replicas"`
	TargetRep  int      `json:"target_replication"`
	Created    int64    `json:"created"`
	Updated    int64    `json:"updated"`
}

type Cloud struct {
	mu         sync.RWMutex
	dataDir    string
	containers map[string]*ContainerMeta
	seeding    map[string]*ActiveSeed
	transport  Transport
	swarm      *SwarmManager
	stopCh     chan struct{}
	stopOnce   sync.Once
}

type ActiveSeed struct {
	Infohash   string
	Manifest   *Manifest
	PieceCount int
	Started    time.Time
}

type Transport interface {
	AnnounceHave(hash string)
	AddPeer(addr, id string)
	RequestPiece(peerAddr, infohash string, pieceIndex int) ([]byte, error)
	RequestManifest(peerAddr, infohash string) (*Manifest, error)
}


// NewCloud handles the NewCloud HTTP request.
func NewCloud(dataDir string) *Cloud {
	dcDir := filepath.Join(dataDir, "dc")
	for _, d := range []string{ManifestDir, ContainerDir, PieceCacheDir, "seeding"} {
		os.MkdirAll(filepath.Join(dcDir, d), 0755)
	}
	c := &Cloud{
		dataDir:    dcDir,
		containers: make(map[string]*ContainerMeta),
		seeding:    make(map[string]*ActiveSeed),
		stopCh:     make(chan struct{}),
	}
	c.swarm = NewSwarmManager(c)
	return c
}


// Start handles the Start HTTP request.
func (c *Cloud) Start() {
	go c.cleanLoop()
	go c.healingLoop()
	go c.swarm.Run()
	slog.Info("dc cloud started", "dir", c.dataDir)
}


// Stop handles the Stop HTTP request.
func (c *Cloud) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
		c.mu.Lock()
		for hash, s := range c.seeding {
			c.saveSeedingState(hash, s)
		}
		c.mu.Unlock()
	})
}


// ListContainers handles the ListContainers HTTP request.
func (c *Cloud) ListContainers() []ContainerMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]ContainerMeta, 0, len(c.containers))
	for _, m := range c.containers {
		out = append(out, *m)
	}
	return out
}


// GetContainer handles the GetContainer HTTP request.
func (c *Cloud) GetContainer(infohash string) *ContainerMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.containers[infohash]
	if !ok {
		return nil
	}
	return m
}


// RegisterContainer handles the RegisterContainer HTTP request.
func (c *Cloud) RegisterContainer(infohash string, size int64, pieceCount int) *ContainerMeta {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registerContainerLocked(infohash, size, pieceCount)
}

func (c *Cloud) registerContainerLocked(infohash string, size int64, pieceCount int) *ContainerMeta {
	m := &ContainerMeta{
		Infohash:   infohash,
		Size:       size,
		PieceCount: pieceCount,
		TargetRep:  DefaultReplication,
		Replicas:   1,
		Created:    time.Now().Unix(),
		Updated:    time.Now().Unix(),
	}
	c.containers[infohash] = m
	c.saveState()
	return m
}


// AddSeeder handles the AddSeeder HTTP request.
func (c *Cloud) AddSeeder(infohash, peerAddr string) {
	if infohash == "" || peerAddr == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.containers[infohash]
	if !ok {
		return
	}
	for _, s := range m.Seeders {
		if s == peerAddr {
			return
		}
	}
	m.Seeders = append(m.Seeders, peerAddr)
	m.Replicas = len(m.Seeders)
	m.Updated = time.Now().Unix()
	c.saveState()
}


// AddLeecher handles the AddLeecher HTTP request.
func (c *Cloud) AddLeecher(infohash, peerAddr string) {
	if infohash == "" || peerAddr == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.containers[infohash]
	if !ok {
		return
	}
	for _, l := range m.Leechers {
		if l == peerAddr {
			return
		}
	}
	m.Leechers = append(m.Leechers, peerAddr)
	m.Updated = time.Now().Unix()
}


// RemovePeer handles the RemovePeer HTTP request.
func (c *Cloud) RemovePeer(peerAddr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, m := range c.containers {
		filteredS := make([]string, 0, len(m.Seeders))
		for _, s := range m.Seeders {
			if s != peerAddr {
				filteredS = append(filteredS, s)
			}
		}
		m.Seeders = filteredS
		filteredL := make([]string, 0, len(m.Leechers))
		for _, l := range m.Leechers {
			if l != peerAddr {
				filteredL = append(filteredL, l)
			}
		}
		m.Leechers = filteredL
		m.Replicas = len(m.Seeders)
	}
	c.saveState()
}


// SeedingStatus handles the SeedingStatus HTTP request.
func (c *Cloud) SeedingStatus() map[string]*ActiveSeed {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]*ActiveSeed, len(c.seeding))
	for k, v := range c.seeding {
		out[k] = v
	}
	return out
}


// IsSeeding handles the IsSeeding HTTP request.
func (c *Cloud) IsSeeding(infohash string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.seeding[infohash]
	return ok
}


// LoadState handles the LoadState HTTP request.
func (c *Cloud) LoadState() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadState()
}


// SetTransport handles the SetTransport HTTP request.
func (c *Cloud) SetTransport(t Transport) {
	c.mu.Lock()
	c.transport = t
	c.mu.Unlock()
}

func (c *Cloud) getTransport() Transport {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.transport
}

func (c *Cloud) cleanLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now().Unix()
			for hash, m := range c.containers {
				if now-m.Updated > 300 {
					delete(c.containers, hash)
				}
			}
			c.saveState()
			c.mu.Unlock()
		}
	}
}


// ForceHeal handles the ForceHeal HTTP request.
func (c *Cloud) ForceHeal() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, m := range c.containers {
		if m.Replicas < m.TargetRep && m.Replicas > 0 {
			slog.Info("dc force heal", "infohash", m.Infohash, "replicas", m.Replicas, "target", m.TargetRep)
		}
	}
}

func (c *Cloud) healingLoop() {
	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.ForceHeal()
		}
	}
}

func (c *Cloud) statePath() string {
	return filepath.Join(c.dataDir, "dc_state.json")
}

func (c *Cloud) saveState() {
	data, err := json.Marshal(c.containers)
	if err != nil {
		slog.Error("dc save state marshal", "error", err)
		return
	}
	tmpPath := c.statePath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		slog.Error("dc save state write", "error", err)
		return
	}
	os.Rename(tmpPath, c.statePath())
}

func (c *Cloud) loadState() {
	data, err := os.ReadFile(c.statePath())
	if err != nil {
		return
	}
	var containers map[string]*ContainerMeta
	if err := json.Unmarshal(data, &containers); err != nil {
		return
	}
	c.containers = containers
}

func (c *Cloud) saveSeedingState(hash string, s *ActiveSeed) {
	path := filepath.Join(c.dataDir, "seeding", hash+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		slog.Error("dc save seeding mkdir", "error", err)
		return
	}
	data, err := json.Marshal(s)
	if err != nil {
		slog.Error("dc save seeding marshal", "error", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		slog.Error("dc save seeding write", "error", err)
	}
}


// Infohash handles the Infohash HTTP request.
func Infohash(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
