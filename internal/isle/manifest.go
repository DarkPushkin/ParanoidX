// Package isle provides torrent-like manifest generation for distributed radio tracks.
// A .isle file describes a track broken into fixed-size pieces with SHA-256 hashes,
// enabling P2P distribution across the peer network.
package isle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	PieceSize = 256 * 1024 // 256KB per piece
)

// Manifest describes a distributable track.
type Manifest struct {
	TrackID   string   `json:"track_id"`
	Title     string   `json:"title"`
	Kind      string   `json:"kind"`
	Size      int64    `json:"size"`
	PieceSize int      `json:"piece_size"`
	Pieces    []string `json:"pieces"` // SHA-256 hashes of each piece
	Seeders   []string `json:"seeders,omitempty"` // peer addresses
}

// BuildManifest creates a .isle manifest from an audio file.
func BuildManifest(trackID, filePath, title, kind string) (*Manifest, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	m := &Manifest{
		TrackID:   trackID,
		Title:     title,
		Kind:      kind,
		Size:      stat.Size(),
		PieceSize: PieceSize,
	}

	buf := make([]byte, PieceSize)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			h := sha256.Sum256(buf[:n])
			m.Pieces = append(m.Pieces, fmt.Sprintf("%x", h))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
	}

	return m, nil
}

// Save writes the manifest to a .isle file.
func (m *Manifest) Save(dir string) error {
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, m.TrackID+".isle")
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadManifest reads a .isle file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Verify checks a file against the manifest piece hashes.
func (m *Manifest) Verify(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, m.PieceSize)
	for _, expected := range m.Pieces {
		n, err := f.Read(buf)
		if n == 0 {
			return false
		}
		h := sha256.Sum256(buf[:n])
		if fmt.Sprintf("%x", h) != expected {
			return false
		}
		if err == io.EOF {
			break
		}
	}
	return true
}
