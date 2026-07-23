// Package bot provides Telegram bot implementations
package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NodeAPI provides methods to query the local simplex-node for live data.
type NodeAPI struct {
	BaseURL string
	client  *http.Client
}

// NewNodeAPI creates a client connected to the local node.
func NewNodeAPI(baseURL string) *NodeAPI {
	return &NodeAPI{
		BaseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *NodeAPI) get(path string, dest any) error {
	url := n.BaseURL + path
	resp, err := n.client.Get(url)
	if err != nil {
		return fmt.Errorf("get %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode %s: %w: %s", path, err, string(body))
	}
	return nil
}

// StatusResult mirrors /api/status
type StatusResult struct {
	Locked    bool                   `json:"locked"`
	Uptime    string                 `json:"uptime"`
	Addresses map[string]string      `json:"addresses"`
	Vault     map[string]any         `json:"vault"`
	Disk      map[string]any         `json:"disk"`
	Containers []map[string]any      `json:"containers"`
	Version   string                 `json:"version"`
}

// NodeStatus fetches /api/status
func (n *NodeAPI) NodeStatus() (*StatusResult, error) {
	var raw map[string]any
	if err := n.get("/api/status", &raw); err != nil {
		return nil, err
	}
	s := &StatusResult{}
	if v, ok := raw["locked"].(bool); ok {
		s.Locked = v
	}
	if v, ok := raw["uptime"].(string); ok {
		s.Uptime = v
	}
	if v, ok := raw["version"].(string); ok {
		s.Version = v
	}
	if v, ok := raw["vault"].(map[string]any); ok {
		s.Vault = v
	}
	if v, ok := raw["disk"].(map[string]any); ok {
		s.Disk = v
	}
	if v, ok := raw["containers"].([]any); ok {
		for _, c := range v {
			if cm, ok := c.(map[string]any); ok {
				s.Containers = append(s.Containers, cm)
			}
		}
	}
	if v, ok := raw["addresses"].(map[string]any); ok {
		s.Addresses = make(map[string]string)
		for k, vv := range v {
			s.Addresses[k] = fmt.Sprintf("%v", vv)
		}
	}
	return s, nil
}

// EconomyState mirrors /api/economy/state
type EconomyState struct {
	TotalSupplyNG    int64 `json:"total_supply_ng"`
	Accounts         int   `json:"accounts"`
	ReserveNG        int64 `json:"reserve_ng"`
	BanknotesActive  int   `json:"banknotes_active"`
	BanknotesTotal   int   `json:"banknotes_total"`
	PreMintAvailable int   `json:"pre_mint_available"`
	BurnedSerials    int   `json:"burned_serials"`
}

// EconomyState fetches /api/economy/state
func (n *NodeAPI) EconomyState() (*EconomyState, error) {
	var raw map[string]any
	if err := n.get("/api/economy/state", &raw); err != nil {
		return nil, err
	}
	s := &EconomyState{}
	if v, ok := raw["total_supply_ng"].(float64); ok {
		s.TotalSupplyNG = int64(v)
	}
	if v, ok := raw["accounts"].(float64); ok {
		s.Accounts = int(v)
	}
	if v, ok := raw["reserve_ng"].(float64); ok {
		s.ReserveNG = int64(v)
	}
	if v, ok := raw["banknotes_active"].(float64); ok {
		s.BanknotesActive = int(v)
	}
	if v, ok := raw["banknotes_total"].(float64); ok {
		s.BanknotesTotal = int(v)
	}
	if v, ok := raw["pre_mint_available"].(float64); ok {
		s.PreMintAvailable = int(v)
	}
	if v, ok := raw["burned_serials"].(float64); ok {
		s.BurnedSerials = int(v)
	}
	return s, nil
}

// DiskResult mirrors /api/disk-check
type DiskResult struct {
	Mounts map[string]map[string]any `json:"-"`
	Raw    map[string]any
}

// DiskUsage fetches /api/disk-check
func (n *NodeAPI) DiskUsage() (map[string]any, error) {
	var raw map[string]any
	if err := n.get("/api/disk-check", &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// HealthResult mirrors /api/health
type HealthResult struct {
	Healthy   bool                   `json:"healthy"`
	Services  []map[string]any       `json:"services"`
	Disk      map[string]any         `json:"disk"`
}

// Health fetches /api/health
func (n *NodeAPI) Health() (*HealthResult, error) {
	var raw map[string]any
	if err := n.get("/api/health", &raw); err != nil {
		return nil, err
	}
	h := &HealthResult{}
	if v, ok := raw["healthy"].(bool); ok {
		h.Healthy = v
	}
	if v, ok := raw["services"].([]any); ok {
		for _, s := range v {
			if sm, ok := s.(map[string]any); ok {
				h.Services = append(h.Services, sm)
			}
		}
	}
	if v, ok := raw["disk"].(map[string]any); ok {
		h.Disk = v
	}
	return h, nil
}

// StewardStatus mirrors /api/steward
type StewardStatus struct {
	Enabled      bool                   `json:"enabled"`
	AutoAdjust   bool                   `json:"auto_adjust"`
	RunCount     int                    `json:"run_count"`
	LastRun      string                 `json:"last_run"`
	ActionCount  int                    `json:"action_count"`
	RecentActions []map[string]any      `json:"recent_actions"`
	Metrics      map[string]any         `json:"metrics"`
	Constitution []map[string]any       `json:"constitution"`
	Params       map[string]any         `json:"params"`
}

// Steward fetches /api/steward
func (n *NodeAPI) Steward() (*StewardStatus, error) {
	var raw map[string]any
	if err := n.get("/api/steward", &raw); err != nil {
		return nil, err
	}
	s := &StewardStatus{}
	if v, ok := raw["enabled"].(bool); ok {
		s.Enabled = v
	}
	if v, ok := raw["auto_adjust"].(bool); ok {
		s.AutoAdjust = v
	}
	if v, ok := raw["run_count"].(float64); ok {
		s.RunCount = int(v)
	}
	if v, ok := raw["last_run"].(string); ok {
		s.LastRun = v
	}
	if v, ok := raw["action_count"].(float64); ok {
		s.ActionCount = int(v)
	}
	if v, ok := raw["recent_actions"].([]any); ok {
		for _, a := range v {
			if am, ok := a.(map[string]any); ok {
				s.RecentActions = append(s.RecentActions, am)
			}
		}
	}
	if v, ok := raw["metrics"].(map[string]any); ok {
		s.Metrics = v
	}
	if v, ok := raw["constitution"].([]any); ok {
		for _, c := range v {
			if cm, ok := c.(map[string]any); ok {
				s.Constitution = append(s.Constitution, cm)
			}
		}
	}
	if v, ok := raw["params"].(map[string]any); ok {
		s.Params = v
	}
	return s, nil
}

// POSStats mirrors /api/pos?action=stats
type POSStats struct {
	TotalVolumeNG int64 `json:"total_volume_ng"`
	TotalInvoices int   `json:"total_invoices"`
	Paid          int   `json:"paid"`
	TotalCommission int64 `json:"total_commission"`
}

// POSStats fetches POS statistics
func (n *NodeAPI) POSStats() (*POSStats, error) {
	var raw map[string]any
	if err := n.get("/api/pos?action=stats", &raw); err != nil {
		return nil, err
	}
	s := &POSStats{}
	if v, ok := raw["total_volume_ng"].(float64); ok {
		s.TotalVolumeNG = int64(v)
	}
	if v, ok := raw["total_invoices"].(float64); ok {
		s.TotalInvoices = int(v)
	}
	if v, ok := raw["paid"].(float64); ok {
		s.Paid = int(v)
	}
	if v, ok := raw["total_commission"].(float64); ok {
		s.TotalCommission = int64(v)
	}
	return s, nil
}
