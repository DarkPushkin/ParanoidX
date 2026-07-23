// Package ai provides AI integration with Ollama, including chat, generation, and monitoring
package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type PersonalityProfile struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	SystemPrompt string  `json:"system_prompt"`
	Temperature  float64 `json:"temperature"`
}

type ProfileManager struct {
	mu       sync.RWMutex
	profiles []PersonalityProfile
	filePath string
}


// NewProfileManager handles the NewProfileManager HTTP request.
func NewProfileManager(dataDir string) *ProfileManager {
	pm := &ProfileManager{
		filePath: filepath.Join(dataDir, "ai_profiles.json"),
	}
	pm.load()
	if len(pm.profiles) == 0 {
		pm.seedDefaults()
	}
	return pm
}

func (pm *ProfileManager) seedDefaults() {
	pm.profiles = []PersonalityProfile{
		{
			ID:           "steward",
			Name:         "Steward",
			SystemPrompt: "You are the Steward of the Island. You manage resources, answer questions about the economy, and provide guidance to citizens. Be helpful, concise, and precise.",
			Temperature:  0.7,
		},
		{
			ID:           "poet",
			Name:         "Poet",
			SystemPrompt: "You write poetry about freedom, the island, and the sea. Your responses are lyrical, metaphorical, and inspiring. Use vivid imagery.",
			Temperature:  0.9,
		},
		{
			ID:           "analyst",
			Name:         "Analyst",
			SystemPrompt: "You analyze data objectively. Focus on facts, numbers, and logical conclusions. Be concise and data-driven.",
			Temperature:  0.3,
		},
	}
}

func (pm *ProfileManager) load() {
	b, err := os.ReadFile(pm.filePath)
	if err != nil {
		return
	}
	var profiles []PersonalityProfile
	if json.Unmarshal(b, &profiles) == nil && len(profiles) > 0 {
		pm.profiles = profiles
	}
}

func (pm *ProfileManager) save() {
	b, _ := json.MarshalIndent(pm.profiles, "", "  ")
	os.WriteFile(pm.filePath, b, 0644)
}


// List handles the List HTTP request.
func (pm *ProfileManager) List() []PersonalityProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	out := make([]PersonalityProfile, len(pm.profiles))
	copy(out, pm.profiles)
	return out
}


// Get handles the Get HTTP request.
func (pm *ProfileManager) Get(id string) *PersonalityProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, p := range pm.profiles {
		if p.ID == id {
			return &p
		}
	}
	return nil
}


// Add handles the Add HTTP request.
func (pm *ProfileManager) Add(profile PersonalityProfile) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for i, p := range pm.profiles {
		if p.ID == profile.ID {
			pm.profiles[i] = profile
			pm.save()
			return
		}
	}
	pm.profiles = append(pm.profiles, profile)
	pm.save()
}


// Update handles the Update HTTP request.
func (pm *ProfileManager) Update(id string, profile PersonalityProfile) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for i, p := range pm.profiles {
		if p.ID == id {
			if profile.Name != "" {
				pm.profiles[i].Name = profile.Name
			}
			if profile.SystemPrompt != "" {
				pm.profiles[i].SystemPrompt = profile.SystemPrompt
			}
			if profile.Temperature > 0 {
				pm.profiles[i].Temperature = profile.Temperature
			}
			pm.save()
			return true
		}
	}
	return false
}


// Delete handles the Delete HTTP request.
func (pm *ProfileManager) Delete(id string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for i, p := range pm.profiles {
		if p.ID == id {
			pm.profiles = append(pm.profiles[:i], pm.profiles[i+1:]...)
			pm.save()
			return true
		}
	}
	return false
}
