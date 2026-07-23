// Package paranoidx implements the ParanoidX multi-layer proxy chain
package paranoidx

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// V2RayManager handles V2Ray proxy configuration and lifecycle.
type V2RayManager struct {
	ConfigDir string
	BinPath   string
	cmd       *exec.Cmd
}

// V2RayConfig represents the V2Ray client configuration structure.
type V2RayConfig struct {
	Log    json.RawMessage `json:"log"`
	Inbounds  []json.RawMessage `json:"inbounds"`
	Outbounds []json.RawMessage `json:"outbounds"`
	Routing   json.RawMessage   `json:"routing,omitempty"`
}

// NewV2RayManager creates a V2Ray manager with config in the given directory.
func NewV2RayManager(dataDir string) *V2RayManager {
	return &V2RayManager{
		ConfigDir: filepath.Join(dataDir, "v2ray"),
		BinPath:   "/usr/bin/v2ray",
	}
}

// EnsureConfig writes a default V2Ray client config if none exists.
func (m *V2RayManager) EnsureConfig(serverAddr, serverPort, userID string) error {
	if err := os.MkdirAll(m.ConfigDir, 0755); err != nil {
		return fmt.Errorf("mkdir v2ray config: %w", err)
	}
	configPath := filepath.Join(m.ConfigDir, "config.json")
	if _, err := os.Stat(configPath); err == nil {
		return nil
	}
	cfg := m.defaultClientConfig(serverAddr, serverPort, userID)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal v2ray config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write v2ray config: %w", err)
	}
	return nil
}

// defaultClientConfig generates a V2Ray client config with WebSocket + TLS.
func (m *V2RayManager) defaultClientConfig(addr, port, uid string) map[string]any {
	return map[string]any{
		"log": map[string]any{
			"loglevel": "warning",
		},
		"inbounds": []map[string]any{{
			"port":     10808,
			"listen":   "127.0.0.1",
			"protocol": "socks",
			"settings": map[string]any{
				"udp": true,
			},
			"tag": "socks-in",
		}},
		"outbounds": []map[string]any{{
			"protocol": "vmess",
			"settings": map[string]any{
				"vnext": []map[string]any{{
					"address": addr,
					"port":    port,
					"users": []map[string]any{{
						"id":       uid,
						"security": "aes-128-gcm",
					}},
				}},
			},
			"streamSettings": map[string]any{
				"network": "tcp",
				"security": "tls",
			},
			"tag": "proxy",
		}, {
			"protocol": "freedom",
			"tag":      "direct",
		}},
		"routing": map[string]any{
			"domainStrategy": "AsIs",
			"rules": []map[string]any{{
				"type": "field",
				"inboundTag": []string{"socks-in"},
				"outboundTag": "proxy",
			}},
		},
	}
}

// Start launches the V2Ray process with the generated config.
func (m *V2RayManager) Start() error {
	configPath := filepath.Join(m.ConfigDir, "config.json")
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("v2ray config not found: %w", err)
	}
	m.cmd = exec.Command(m.BinPath, "-config", configPath)
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("start v2ray: %w", err)
	}
	go func() {
		m.cmd.Wait()
		SetLayerStatus(LayerV2Ray, false, 0, "v2ray process exited")
	}()
	time.Sleep(500 * time.Millisecond)
	SetLayerStatus(LayerV2Ray, true, 5, fmt.Sprintf("v2ray running (pid %d)", m.cmd.Process.Pid))
	return nil
}

// Stop terminates the V2Ray process.
func (m *V2RayManager) Stop() {
	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Kill()
		m.cmd.Wait()
	}
	SetLayerStatus(LayerV2Ray, false, 0, "v2ray stopped")
}

// CheckHealth probes the V2Ray SOCKS5 proxy port.
func (m *V2RayManager) CheckHealth() bool {
	start := time.Now()
	c := exec.Command("bash", "-c", "curl -s --connect-timeout 2 -x socks5://127.0.0.1:10808 http://httpbin.org/ip >/dev/null 2>&1")
	if err := c.Run(); err != nil {
		SetLayerStatus(LayerV2Ray, false, 0, "health check failed")
		return false
	}
	latencyMs := time.Since(start).Milliseconds()
	SetLayerStatus(LayerV2Ray, true, latencyMs, "healthy")
	return true
}
