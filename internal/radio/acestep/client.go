// Package acestep provides a client for the Acestep AI audio generation API.
// Acestep runs on a separate laptop and converts text prompts into
// style-aware audio: gospel ads, tragic news, royal decrees, and music.
package acestep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Style string

const (
	StyleGospel  Style = "gospel"   // ADVE — black preacher, joyful
	StyleTragedy Style = "tragedy"  // NEWS — breaking news, dramatic
	StyleRoyal   Style = "royal"    // KING — majestic, decree
	StyleMusic   Style = "music"    // MUSIC — ambient, island beats
)

type GenerateRequest struct {
	Prompt   string `json:"prompt"`
	Style    Style  `json:"style"`
	Duration int    `json:"duration"` // seconds
	Voice    string `json:"voice,omitempty"`
}

type GenerateResponse struct {
	AudioURL  string `json:"audio_url"`
	Duration  int    `json:"duration"`
	TrackID   string `json:"track_id"`
	CreatedAt string `json:"created_at"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	cacheDir   string
	mu         sync.RWMutex
	cache      map[string]string // prompt_hash -> local file path
}


// NewClient handles the NewClient HTTP request.
func NewClient(baseURL, cacheDir string) *Client {
	os.MkdirAll(cacheDir, 0755)
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		cacheDir:   cacheDir,
		cache:      make(map[string]string),
	}
}


// Generate handles the Generate HTTP request.
func (c *Client) Generate(prompt string, style Style, duration int) (*GenerateResponse, error) {
	req := &GenerateRequest{
		Prompt:   prompt,
		Style:    style,
		Duration: duration,
	}
	body, _ := json.Marshal(req)
	resp, err := c.httpClient.Post(c.baseURL+"/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("acestep generate: %w", err)
	}
	defer resp.Body.Close()

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("acestep decode: %w", err)
	}
	return &genResp, nil
}


// GenerateAndCache handles the GenerateAndCache HTTP request.
func (c *Client) GenerateAndCache(prompt string, style Style, duration int) (string, error) {
	gen, err := c.Generate(prompt, style, duration)
	if err != nil {
		return "", err
	}
	localPath, err := c.downloadAndCache(gen)
	if err != nil {
		return "", err
	}
	return localPath, nil
}

func (c *Client) downloadAndCache(gen *GenerateResponse) (string, error) {
	resp, err := c.httpClient.Get(gen.AudioURL)
	if err != nil {
		return "", fmt.Errorf("acestep download: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("acestep read body: %w", err)
	}

	localPath := filepath.Join(c.cacheDir, gen.TrackID+".mp3")
	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return "", fmt.Errorf("acestep write cache: %w", err)
	}

	c.mu.Lock()
	c.cache[gen.TrackID] = localPath
	c.mu.Unlock()

	return localPath, nil
}


// Cached handles the Cached HTTP request.
func (c *Client) Cached(id string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[id]
}


// Healthy handles the Healthy HTTP request.
func (c *Client) Healthy() bool {
	hc := &http.Client{Timeout: 3 * time.Second}
	resp, err := hc.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
