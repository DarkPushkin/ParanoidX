// Package royal implements the Royal→Sub-node command protocol.
// Royal node signs commands with its Ed25519 key, sub-nodes verify and execute.
package royal

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

const (
	HeartbeatInterval = 5 * time.Minute
	StaleThreshold    = 15 * time.Minute
)

// NodeType indicates whether a node is Royal or Sub.
type NodeType string

const (
	NodeTypeRoyal NodeType = "royal"
	NodeTypeSub   NodeType = "sub"
)

type NodeStatus string

const (
	StatusOnline  NodeStatus = "online"
	StatusOffline NodeStatus = "offline"
	StatusStale   NodeStatus = "stale"
)

// NodeInfo describes a registered sub-node.
type NodeInfo struct {
	Pubkey     string     `json:"pubkey"`
	Label      string     `json:"label"`
	Type       NodeType   `json:"type"`
	Addr       string     `json:"addr"`  // SMP address or onion
	Status     NodeStatus `json:"status"`
	LastSeen   string     `json:"last_seen"`
	Registered string    `json:"registered"`
}

// SignedCommand is a command signed by Royal's Ed25519 key.
type SignedCommand struct {
	Command   string `json:"command"`
	Target    string `json:"target"`     // sub-node pubkey
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`  // hex-encoded Ed25519 sig
}

// CommandReceipt is a sub-node's signed acknowledgment.
type CommandReceipt struct {
	Command   string `json:"command"`
	Status    string `json:"status"` // "accepted", "rejected", "done", "failed"
	Result    string `json:"result,omitempty"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"` // sub-node's sig
}

// Service manages the Royal→Sub node protocol.
type Service struct {
	mu       sync.RWMutex
	Nodes    map[string]*NodeInfo   `json:"nodes"`
	Pubkey   ed25519.PublicKey
	Privkey  ed25519.PrivateKey
	DataDir  string
}

// NewService creates a Service, generating a keypair if none exists.
func NewService(dataDir string) *Service {
	s := &Service{
		Nodes:   make(map[string]*NodeInfo),
		DataDir: dataDir,
	}
	s.loadOrGenerateKey()
	s.loadNodes()
	return s
}

func (s *Service) loadOrGenerateKey() {
	keyFile := s.DataDir + "/royal_ed25519.json"
	type keyData struct {
		Pubkey  string `json:"pubkey"`
		Privkey string `json:"privkey"`
	}
	b, err := os.ReadFile(keyFile)
	if err == nil {
		var kd keyData
		if json.Unmarshal(b, &kd) == nil {
			pub, _ := hex.DecodeString(kd.Pubkey)
			priv, _ := hex.DecodeString(kd.Privkey)
			if len(pub) == ed25519.PublicKeySize && len(priv) == ed25519.PrivateKeySize {
				s.Pubkey = pub
				s.Privkey = priv
				return
			}
		}
	}
	// Generate new keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		slog.Error("royal: generate keypair", "error", err)
		return
	}
	s.Pubkey = pub
	s.Privkey = priv
	kd := keyData{Pubkey: hex.EncodeToString(pub), Privkey: hex.EncodeToString(priv)}
	b, _ = json.MarshalIndent(kd, "", "  ")
	os.WriteFile(keyFile, b, 0600)
	slog.Info("royal: generated new Ed25519 keypair", "pubkey", hex.EncodeToString(pub))
}

func (s *Service) loadNodes() {
	b, err := os.ReadFile(s.DataDir + "/royal_nodes.json")
	if err != nil {
		return
	}
	json.Unmarshal(b, &s.Nodes)
	if s.Nodes == nil {
		s.Nodes = make(map[string]*NodeInfo)
	}
}

func (s *Service) saveNodes() {
	b, _ := json.MarshalIndent(s.Nodes, "", "  ")
	os.WriteFile(s.DataDir+"/royal_nodes.json", b, 0600)
}

// PublicKeyHex returns the hex-encoded Ed25519 public key.
func (s *Service) PublicKeyHex() string {
	if s.Pubkey == nil {
		return ""
	}
	return hex.EncodeToString(s.Pubkey)
}

// RegisterNode registers a sub-node.
func (s *Service) RegisterNode(pubkey, label, addr string) (*NodeInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.Nodes[pubkey]; exists {
		return nil, fmt.Errorf("node already registered: %s", pubkey)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	node := &NodeInfo{
		Pubkey:     pubkey,
		Label:      label,
		Type:       NodeTypeSub,
		Addr:       addr,
		Status:     StatusOnline,
		LastSeen:   now,
		Registered: now,
	}
	s.Nodes[pubkey] = node
	s.saveNodes()
	slog.Info("royal: node registered", "pubkey", pubkey, "label", label)
	return node, nil
}

// SignCommand signs a command for a target sub-node.
func (s *Service) SignCommand(command, targetPubkey string) (*SignedCommand, error) {
	if s.Privkey == nil {
		return nil, fmt.Errorf("royal keypair not loaded")
	}
	nonce := make([]byte, 16)
	rand.Read(nonce)
	sc := &SignedCommand{
		Command:   command,
		Target:    targetPubkey,
		Nonce:     hex.EncodeToString(nonce),
		Timestamp: time.Now().Unix(),
	}
	payload := fmt.Sprintf("%s|%s|%s|%d", sc.Command, sc.Target, sc.Nonce, sc.Timestamp)
	sig := ed25519.Sign(s.Privkey, []byte(payload))
	sc.Signature = hex.EncodeToString(sig)
	return sc, nil
}

// VerifyCommand verifies a signed command.
func VerifyCommand(sc *SignedCommand, pubkey ed25519.PublicKey) bool {
	payload := fmt.Sprintf("%s|%s|%s|%d", sc.Command, sc.Target, sc.Nonce, sc.Timestamp)
	sig, err := hex.DecodeString(sc.Signature)
	if err != nil {
		return false
	}
	return ed25519.Verify(pubkey, []byte(payload), sig)
}

// Heartbeat updates the last-seen timestamp for a sub-node.
func (s *Service) Heartbeat(pubkey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if node, ok := s.Nodes[pubkey]; ok {
		node.LastSeen = time.Now().UTC().Format(time.RFC3339)
		node.Status = StatusOnline
		s.saveNodes()
	}
}

// CheckStale marks nodes as stale if they haven't been seen recently.
func (s *Service) CheckStale() {
	s.mu.Lock()
	defer s.mu.Unlock()
	threshold := time.Now().Add(-StaleThreshold)
	for _, node := range s.Nodes {
		lastSeen, err := time.Parse(time.RFC3339, node.LastSeen)
		if err != nil || lastSeen.Before(threshold) {
			node.Status = StatusStale
		}
	}
	s.saveNodes()
}

// ListNodes returns all registered nodes.
func (s *Service) ListNodes() []*NodeInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*NodeInfo, 0, len(s.Nodes))
	for _, n := range s.Nodes {
		out = append(out, n)
	}
	return out
}

// GetNode returns a node by pubkey.
func (s *Service) GetNode(pubkey string) *NodeInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Nodes[pubkey]
}
