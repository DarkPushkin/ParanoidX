// Package acestep provides Acestream integration for P2P streaming
package acestep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)


// TestStyleConstants handles the TestStyleConstants HTTP request.
func TestStyleConstants(t *testing.T) {
	tests := []struct {
		style Style
		want  string
	}{
		{StyleGospel, "gospel"},
		{StyleTragedy, "tragedy"},
		{StyleRoyal, "royal"},
		{StyleMusic, "music"},
	}
	for _, tt := range tests {
		if string(tt.style) != tt.want {
			t.Errorf("Style(%s) = %q, want %q", tt.style, string(tt.style), tt.want)
		}
	}
}


// TestNewClient handles the TestNewClient HTTP request.
func TestNewClient(t *testing.T) {
	cacheDir := t.TempDir()
	c := NewClient("http://test:8080", cacheDir)

	if c.baseURL != "http://test:8080" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://test:8080")
	}
	if c.cacheDir != cacheDir {
		t.Errorf("cacheDir = %q, want %q", c.cacheDir, cacheDir)
	}
	if c.httpClient.Timeout != 120*time.Second {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 120*time.Second)
	}

	// Verify cache dir was created
	info, err := os.Stat(cacheDir)
	if err != nil {
		t.Fatalf("cache dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("cache dir is not a directory")
	}

	// Verify cache map is initialized
	if c.cache == nil {
		t.Error("cache map is nil")
	}
}


// TestNewGenerator handles the TestNewGenerator HTTP request.
func TestNewGenerator(t *testing.T) {
	cacheDir := t.TempDir()
	g := NewGenerator("http://test:8080", cacheDir)

	expectedAcDir := filepath.Join(cacheDir, "acestep")
	if g.cacheDir != expectedAcDir {
		t.Errorf("cacheDir = %q, want %q", g.cacheDir, expectedAcDir)
	}
	if g.client == nil {
		t.Error("client is nil")
	}
	if g.client.baseURL != "http://test:8080" {
		t.Errorf("client baseURL = %q, want %q", g.client.baseURL, "http://test:8080")
	}
	if g.client.cacheDir != expectedAcDir {
		t.Errorf("client cacheDir = %q, want %q", g.client.cacheDir, expectedAcDir)
	}

	// Verify acestep subdir was created
	info, err := os.Stat(expectedAcDir)
	if err != nil {
		t.Fatalf("acestep dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("acestep dir is not a directory")
	}
}


// TestNewLiveBroadcast handles the TestNewLiveBroadcast HTTP request.
func TestNewLiveBroadcast(t *testing.T) {
	cacheDir := t.TempDir()
	gen := NewGenerator("http://test:8080", cacheDir)
	interval := 30 * time.Second
	lb := NewLiveBroadcast(gen, interval)

	if lb.gen != gen {
		t.Error("generator not set correctly")
	}
	if lb.interval != interval {
		t.Errorf("interval = %v, want %v", lb.interval, interval)
	}
	if cap(lb.queue) != 10 {
		t.Errorf("queue capacity = %d, want 10", cap(lb.queue))
	}
	if lb.style != StyleMusic {
		t.Errorf("default style = %q, want %q", lb.style, StyleMusic)
	}
	if lb.cancel == nil {
		t.Error("cancel func is nil")
	}
}


// TestClientCached handles the TestClientCached HTTP request.
func TestClientCached(t *testing.T) {
	cacheDir := t.TempDir()
	c := NewClient("http://test:8080", cacheDir)

	// Before caching anything
	if got := c.Cached("nonexistent"); got != "" {
		t.Errorf("Cached(\"nonexistent\") = %q, want empty string", got)
	}

	// After manual cache insert
	c.mu.Lock()
	c.cache["track1"] = "/path/to/track1.mp3"
	c.mu.Unlock()

	if got := c.Cached("track1"); got != "/path/to/track1.mp3" {
		t.Errorf("Cached(\"track1\") = %q, want %q", got, "/path/to/track1.mp3")
	}
}


// TestGeneratorGetCached handles the TestGeneratorGetCached HTTP request.
func TestGeneratorGetCached(t *testing.T) {
	cacheDir := t.TempDir()
	g := NewGenerator("http://test:8080", cacheDir)

	if got := g.GetCached("gospel"); got != "" {
		t.Errorf("GetCached(\"gospel\") = %q, want empty string", got)
	}

	g.mu.Lock()
	g.generated["gospel"] = "/path/to/gospel.mp3"
	g.mu.Unlock()

	if got := g.GetCached("gospel"); got != "/path/to/gospel.mp3" {
		t.Errorf("GetCached(\"gospel\") = %q, want %q", got, "/path/to/gospel.mp3")
	}
}


// TestLiveBroadcastInitialState handles the TestLiveBroadcastInitialState HTTP request.
func TestLiveBroadcastInitialState(t *testing.T) {
	cacheDir := t.TempDir()
	gen := NewGenerator("http://test:8080", cacheDir)
	lb := NewLiveBroadcast(gen, 30*time.Second)

	if n := lb.QueueLen(); n != 0 {
		t.Errorf("QueueLen() = %d, want 0", n)
	}

	status := lb.Status()
	if status["running"] != true {
		t.Error("Status()[\"running\"] should be true before Start() or Stop()")
	}
	if status["queued"] != 0 {
		t.Errorf("Status()[\"queued\"] = %d, want 0", status["queued"])
	}
	if status["interval"] != "30s" {
		t.Errorf("Status()[\"interval\"] = %q, want \"30s\"", status["interval"])
	}
	if status["style"] != "rotating (music/ad/news/decree)" {
		t.Errorf("Status()[\"style\"] = %q, want \"rotating (music/ad/news/decree)\"", status["style"])
	}
}


// TestLiveBroadcastStop handles the TestLiveBroadcastStop HTTP request.
func TestLiveBroadcastStop(t *testing.T) {
	cacheDir := t.TempDir()
	gen := NewGenerator("http://test:8080", cacheDir)
	lb := NewLiveBroadcast(gen, 30*time.Second)

	lb.Stop()

	status := lb.Status()
	if status["running"] != false {
		t.Error("Status()[\"running\"] should be false after Stop()")
	}
}


// TestClientHealthyOK handles the TestClientHealthyOK HTTP request.
func TestClientHealthyOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cacheDir := t.TempDir()
	c := NewClient(ts.URL, cacheDir)

	if !c.Healthy() {
		t.Error("Healthy() = false, want true")
	}
}


// TestClientHealthyNotOK handles the TestClientHealthyNotOK HTTP request.
func TestClientHealthyNotOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	cacheDir := t.TempDir()
	c := NewClient(ts.URL, cacheDir)

	if c.Healthy() {
		t.Error("Healthy() = true, want false")
	}
}


// TestClientGenerate handles the TestClientGenerate HTTP request.
func TestClientGenerate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/generate" {
			t.Errorf("path = %s, want /generate", r.URL.Path)
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Prompt != "test prompt" {
			t.Errorf("prompt = %q, want %q", req.Prompt, "test prompt")
		}
		if req.Style != StyleGospel {
			t.Errorf("style = %q, want %q", req.Style, StyleGospel)
		}
		if req.Duration != 30 {
			t.Errorf("duration = %d, want 30", req.Duration)
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		audioURL := scheme + "://" + r.Host + "/audio/track123.mp3"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GenerateResponse{
			AudioURL:  audioURL,
			Duration:  30,
			TrackID:   "track123",
			CreatedAt: "2026-06-23T12:00:00Z",
		})
	}))
	defer ts.Close()

	cacheDir := t.TempDir()
	c := NewClient(ts.URL, cacheDir)

	resp, err := c.Generate("test prompt", StyleGospel, 30)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if resp == nil {
		t.Fatal("Generate() returned nil response")
	}
	if resp.TrackID != "track123" {
		t.Errorf("TrackID = %q, want %q", resp.TrackID, "track123")
	}
	if resp.Duration != 30 {
		t.Errorf("Duration = %d, want 30", resp.Duration)
	}
	expectedAudioURL := ts.URL + "/audio/track123.mp3"
	if resp.AudioURL != expectedAudioURL {
		t.Errorf("AudioURL = %q, want %q", resp.AudioURL, expectedAudioURL)
	}
	if resp.CreatedAt != "2026-06-23T12:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", resp.CreatedAt, "2026-06-23T12:00:00Z")
	}
}


// TestClientGenerateAndCache handles the TestClientGenerateAndCache HTTP request.
func TestClientGenerateAndCache(t *testing.T) {
	var ts *httptest.Server
	mux := http.NewServeMux()
	// Handle generate endpoint
	mux.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GenerateResponse{
			AudioURL:  ts.URL + "/audio/track456.mp3",
			Duration:  20,
			TrackID:   "track456",
			CreatedAt: "2026-06-23T12:00:00Z",
		})
	})
	// Handle audio download
	mux.HandleFunc("/audio/track456.mp3", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mock audio data"))
	})

	ts = httptest.NewServer(mux)
	defer ts.Close()

	cacheDir := t.TempDir()
	c := NewClient(ts.URL, cacheDir)

	localPath, err := c.GenerateAndCache("test prompt", StyleTragedy, 20)
	if err != nil {
		t.Fatalf("GenerateAndCache() error: %v", err)
	}
	if localPath == "" {
		t.Fatal("GenerateAndCache() returned empty path")
	}

	// Verify cache entry
	if got := c.Cached("track456"); got != localPath {
		t.Errorf("Cached(\"track456\") = %q, want %q", got, localPath)
	}

	// Verify file was written
	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if string(data) != "mock audio data" {
		t.Errorf("file content = %q, want %q", string(data), "mock audio data")
	}
}


// TestGenerateRequestJSON handles the TestGenerateRequestJSON HTTP request.
func TestGenerateRequestJSON(t *testing.T) {
	req := GenerateRequest{
		Prompt:   "hello world",
		Style:    StyleRoyal,
		Duration: 40,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded GenerateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if decoded.Prompt != "hello world" {
		t.Errorf("Prompt = %q, want %q", decoded.Prompt, "hello world")
	}
	if decoded.Style != StyleRoyal {
		t.Errorf("Style = %q, want %q", decoded.Style, StyleRoyal)
	}
	if decoded.Duration != 40 {
		t.Errorf("Duration = %d, want 40", decoded.Duration)
	}
}


// TestGenerateResponseJSON handles the TestGenerateResponseJSON HTTP request.
func TestGenerateResponseJSON(t *testing.T) {
	resp := GenerateResponse{
		AudioURL:  "http://example.com/audio.mp3",
		Duration:  60,
		TrackID:   "track789",
		CreatedAt: "2026-01-01T00:00:00Z",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded GenerateResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if decoded.AudioURL != "http://example.com/audio.mp3" {
		t.Errorf("AudioURL = %q, want %q", decoded.AudioURL, "http://example.com/audio.mp3")
	}
	if decoded.Duration != 60 {
		t.Errorf("Duration = %d, want 60", decoded.Duration)
	}
	if decoded.TrackID != "track789" {
		t.Errorf("TrackID = %q, want %q", decoded.TrackID, "track789")
	}
	if decoded.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", decoded.CreatedAt, "2026-01-01T00:00:00Z")
	}
}


// TestGenerateRequestVoiceOmitted handles the TestGenerateRequestVoiceOmitted HTTP request.
func TestGenerateRequestVoiceOmitted(t *testing.T) {
	req := GenerateRequest{
		Prompt:   "test",
		Style:    StyleMusic,
		Duration: 120,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	// Verify "voice" is omitted when empty (omitempty tag)
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, exists := raw["voice"]; exists {
		t.Error("voice field should be omitted when empty")
	}
}


// TestGenerateRequestWithVoice handles the TestGenerateRequestWithVoice HTTP request.
func TestGenerateRequestWithVoice(t *testing.T) {
	req := GenerateRequest{
		Prompt:   "test",
		Style:    StyleMusic,
		Duration: 120,
		Voice:    "male-deep",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded GenerateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded.Voice != "male-deep" {
		t.Errorf("Voice = %q, want %q", decoded.Voice, "male-deep")
	}
}
