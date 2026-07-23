// Package acestep provides Acestream integration for P2P streaming
package acestep

import (
	"log"
	"path/filepath"
	"sync"
)

// Generator manages AI-generated radio content via Acestep.
// It injects generated tracks into the Formula playlist by style.
type Generator struct {
	client     *Client
	cacheDir   string
	mu         sync.RWMutex
	generated  map[string]string // style -> local path of latest generated track
}


// NewGenerator handles the NewGenerator HTTP request.
func NewGenerator(acestepURL, cacheDir string) *Generator {
	acDir := filepath.Join(cacheDir, "acestep")
	return &Generator{
		client:    NewClient(acestepURL, acDir),
		cacheDir:  acDir,
		generated: make(map[string]string),
	}
}


// Healthy handles the Healthy HTTP request.
func (g *Generator) Healthy() bool {
	return g.client.Healthy()
}


// GenerateAdvert handles the GenerateAdvert HTTP request.
func (g *Generator) GenerateAdvert(prompt string) (string, error) {
	full := prompt + " — проповедь спасения через silver, радостно, энергично, как чёрный проповедник в церкви"
	return g.client.GenerateAndCache(full, StyleGospel, 30)
}


// GenerateNews handles the GenerateNews HTTP request.
func (g *Generator) GenerateNews(prompt string) (string, error) {
	full := prompt + " — BREAKING NEWS, трагедия в голосе, драматично, срочно"
	return g.client.GenerateAndCache(full, StyleTragedy, 20)
}


// GenerateDecree handles the GenerateDecree HTTP request.
func (g *Generator) GenerateDecree(prompt string) (string, error) {
	full := prompt + " — королевский указ, величественно, торжественно, голос Короля"
	return g.client.GenerateAndCache(full, StyleRoyal, 40)
}


// GenerateMusic handles the GenerateMusic HTTP request.
func (g *Generator) GenerateMusic(mood string) (string, error) {
	full := "Создай музыку для острова: " + mood + " — расслабляющая, атмосферная, островные мотивы"
	return g.client.GenerateAndCache(full, StyleMusic, 120)
}


// Generate handles the Generate HTTP request.
func (g *Generator) Generate(prompt string, style Style, duration int) (string, error) {
	return g.client.GenerateAndCache(prompt, style, duration)
}


// GetCached handles the GetCached HTTP request.
func (g *Generator) GetCached(style string) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.generated[style]
}

func (g *Generator) log(s string) {
	log.Printf("[acestep generator] %s", s)
}
