// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Manifest struct {
	Infohash     string   `json:"infohash"`
	ContainerID  string   `json:"container_id"`
	Size         int64    `json:"size"`
	PieceSize    int      `json:"piece_size"`
	PieceCount   int      `json:"piece_count"`
	Pieces       []string `json:"pieces"`
	Encrypted    bool     `json:"encrypted"`
	SourcePeer   string   `json:"source_peer,omitempty"`
}


// BuildManifest handles the BuildManifest HTTP request.
func BuildManifest(containerID, filePath string) (*Manifest, error) {
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
		ContainerID: containerID,
		Size:        stat.Size(),
		PieceSize:   ContainerPieceSize,
	}

	buf := make([]byte, ContainerPieceSize)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			h := sha256.Sum256(buf[:n])
			m.Pieces = append(m.Pieces, fmt.Sprintf("%x", h))
			m.PieceCount++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
	}

	m.Infohash = computeInfohash(m)
	return m, nil
}

func computeInfohash(m *Manifest) string {
	h := sha256.New()
	h.Write([]byte(m.ContainerID))
	h.Write([]byte(fmt.Sprintf("%d", m.Size)))
	h.Write([]byte(fmt.Sprintf("%d", m.PieceSize)))
	for _, p := range m.Pieces {
		h.Write([]byte(p))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}


// Save handles the Save HTTP request.
func (m *Manifest) Save(dir string) error {
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, m.Infohash+".dc")
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}


// LoadManifest handles the LoadManifest HTTP request.
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


// VerifyPiece handles the VerifyPiece HTTP request.
func (m *Manifest) VerifyPiece(pieceIndex int, data []byte) bool {
	if pieceIndex >= len(m.Pieces) {
		return false
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h) == m.Pieces[pieceIndex]
}


// VerifyAll handles the VerifyAll HTTP request.
func (m *Manifest) VerifyAll(filePath string) bool {
	if m.PieceSize <= 0 {
		return false
	}
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, m.PieceSize)
	for _, expected := range m.Pieces {
		n, err := io.ReadFull(f, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return false
		}
		if n == 0 {
			return false
		}
		h := sha256.Sum256(buf[:n])
		if fmt.Sprintf("%x", h) != expected {
			return false
		}
	}
	return true
}


// PieceCountForSize handles the PieceCountForSize HTTP request.
func (m *Manifest) PieceCountForSize(size int64) int {
	count := int(size / int64(ContainerPieceSize))
	if size%int64(ContainerPieceSize) > 0 {
		count++
	}
	return count
}
