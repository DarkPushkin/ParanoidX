// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type RegistryClient struct {
	baseURL    string
	nodeID     string
	nodeName   string
	directAddr string
	httpCli    *http.Client
}


// NewRegistryClient handles the NewRegistryClient HTTP request.
func NewRegistryClient(baseURL, nodeID, nodeName, directAddr string) *RegistryClient {
	return &RegistryClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		nodeID:     nodeID,
		nodeName:   nodeName,
		directAddr: directAddr,
		httpCli:    &http.Client{Timeout: 5 * time.Second},
	}
}


// SetRegistry handles the SetRegistry HTTP request.
func (c *Cloud) SetRegistry(rc *RegistryClient) {
	c.mu.Lock()
	defer c.mu.Unlock()
	go c.discoveryLoop(rc)
}

func (c *Cloud) discoveryLoop(rc *RegistryClient) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	rc.announce(false)

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			peers := rc.discoverPeers()
			for _, p := range peers {
				c.importPeerContainer(p)
			}
			rc.announce(false)
		}
	}
}

func (rc *RegistryClient) announce(heartbeat bool) {
	if rc.baseURL == "" {
		return
	}
	body, _ := json.Marshal(map[string]any{
		"id":   rc.nodeID,
		"load": 0,
	})
	if heartbeat {
		resp, err := rc.httpCli.Post(rc.baseURL+"/api/node/heartbeat", "application/json", strings.NewReader(string(body)))
		if err == nil {
			resp.Body.Close()
		}
		return
	}
	body, _ = json.Marshal(map[string]any{
		"id":            rc.nodeID,
		"name":          rc.nodeName,
		"direct_addr":   rc.directAddr,
		"region":        "eu",
		"capabilities":  []string{"dc-seed"},
		"public_key":    rc.nodeID,
		"version":       "b122+DC",
		"status":        "online",
		"stake_ng":      0,
		"client_dist":   0,
	})
	resp, err := rc.httpCli.Post(rc.baseURL+"/api/node/announce", "application/json", strings.NewReader(string(body)))
	if err != nil {
		slog.Debug("dc registry announce failed", "error", err)
		return
	}
	resp.Body.Close()
}

func (rc *RegistryClient) discoverPeers() []PeerInfo {
	if rc.baseURL == "" {
		return nil
	}
	resp, err := rc.httpCli.Get(rc.baseURL + "/api/node/list")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Nodes []struct {
			ID           string   `json:"id"`
			DirectAddr   string   `json:"direct_addr"`
			Capabilities []string `json:"capabilities"`
		} `json:"nodes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	var peers []PeerInfo
	for _, n := range result.Nodes {
		if n.ID == rc.nodeID {
			continue
		}
		for _, cap := range n.Capabilities {
			if cap == "dc-seed" {
				peers = append(peers, PeerInfo{
					ID:   n.ID,
					Addr: n.DirectAddr,
				})
				break
			}
		}
	}
	return peers
}

type PeerInfo struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

func (c *Cloud) importPeerContainer(p PeerInfo) {
	if p.Addr == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	slog.Debug("dc discovered peer", "id", p.ID, "addr", p.Addr)
}

func init() {
	slog.Info("dc discovery module loaded")
}
