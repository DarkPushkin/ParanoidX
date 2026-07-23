// Package webrtc provides WebRTC signaling and peer management
package webrtc

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

type SignalState struct {
	mu    sync.Mutex
	rooms map[string]map[string]any
}


// NewSignalState handles the NewSignalState HTTP request.
func NewSignalState() *SignalState {
	return &SignalState{
		rooms: make(map[string]map[string]any),
	}
}


// PostSignal handles the PostSignal HTTP request.
func (s *SignalState) PostSignal(room string, payload map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rooms[room] == nil {
		s.rooms[room] = map[string]any{"candidates": []string{}}
	}
	state := s.rooms[room]
	if o, ok := payload["offer"].(string); ok && o != "" {
		state["offer"] = o
	}
	if a, ok := payload["answer"].(string); ok && a != "" {
		state["answer"] = a
	}
	if c, ok := payload["candidate"].(string); ok && c != "" {
		cands, _ := state["candidates"].([]string)
		state["candidates"] = append(cands, c)
	}
	if candsIface, ok := payload["candidates"]; ok {
		if arr, ok2 := candsIface.([]any); ok2 {
			for _, ci := range arr {
				if cs, ok3 := ci.(string); ok3 {
					cands, _ := state["candidates"].([]string)
					state["candidates"] = append(cands, cs)
				}
			}
		}
	}
	s.rooms[room] = state
}


// GetState handles the GetState HTTP request.
func (s *SignalState) GetState(room string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.rooms[room]
	if state == nil {
		return map[string]any{"candidates": []string{}}
	}
	out := map[string]any{}
	for k, v := range state {
		out[k] = v
	}
	return out
}

// ICEConfig generates TURN credentials using HMAC-SHA1 (12h validity).
// Used by coturn server for WebRTC relay authentication.
type ICEConfig struct {
	Secret string
	Onion  string
}

// GenerateCredentials returns fresh TURN credentials with a 12-hour expiry window.
func (c *ICEConfig) GenerateCredentials() (username, credential string, expiry int64) {
	expiry = time.Now().Unix() + 43200
	username = fmt.Sprintf("%d", expiry)
	h := hmac.New(sha1.New, []byte(c.Secret))
	h.Write([]byte(username))
	credential = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}


// GenerateConfig handles the GenerateConfig HTTP request.
func (c *ICEConfig) GenerateConfig() map[string]any {
	if c.Onion == "" {
		return map[string]any{"iceServers": []any{}}
	}
	secret := c.Secret
	if secret == "" {
		secret = "kgQk2982rPZScVKSI9gIuPOhroefjRgt3rRr1Y1f"
	}
	username, credential, _ := c.GenerateCredentials()

	return map[string]any{
		"iceServers": []map[string]any{
			{
				"urls": []string{
					fmt.Sprintf("turn:%s:3478?transport=tcp", c.Onion),
					fmt.Sprintf("turns:%s:5349?transport=tcp", c.Onion),
				},
				"username":   username,
				"credential": credential,
			},
		},
		"pasteLines": []string{
			fmt.Sprintf("turn:%s:%s@%s:3478?transport=tcp", username, credential, c.Onion),
			fmt.Sprintf("turns:%s:%s@%s:5349?transport=tcp", username, credential, c.Onion),
		},
		"iceOnion": c.Onion,
		"note":     "Use the tcp line first (no cert). For turns:// trust the turn_cert.pem from dashboard. Valid ~12 hours.",
	}
}
