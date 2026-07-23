// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)


// FetchContainer handles the FetchContainer HTTP request.
func (c *Cloud) FetchContainer(infohash string, outputPath string) error {
	c.mu.RLock()
	_, ok := c.containers[infohash]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("container %s not found in cloud", infohash)
	}

	manifestPath := filepath.Join(ManifestPath(c.dataDir), infohash+".dc")
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	t := c.getTransport()
	if !c.IsSeeding(infohash) && t != nil {
		if err := c.fetchMissingPieces(manifest); err != nil {
			return fmt.Errorf("fetch pieces: %w", err)
		}
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	for i, pieceHash := range manifest.Pieces {
		piecePath := filepath.Join(PieceCachePath(c.dataDir), pieceHash)
		data, err := os.ReadFile(piecePath)
		if err != nil {
			return fmt.Errorf("read piece %d (%s): %w", i, pieceHash, err)
		}
		if !manifest.VerifyPiece(i, data) {
			return fmt.Errorf("piece %d hash mismatch", i)
		}
		if _, err := out.Write(data); err != nil {
			return fmt.Errorf("write piece %d: %w", i, err)
		}
	}

	slog.Info("dc fetched container",
		"infohash", infohash,
		"size_mb", int64(len(manifest.Pieces))*int64(ContainerPieceSize)/(1024*1024),
		"output", outputPath,
	)
	return nil
}

func (c *Cloud) fetchMissingPieces(manifest *Manifest) error {
	c.mu.RLock()
	meta, ok := c.containers[manifest.Infohash]
	var seeders []string
	if ok && meta != nil {
		seeders = make([]string, len(meta.Seeders))
		copy(seeders, meta.Seeders)
	}
	c.mu.RUnlock()

	if len(seeders) == 0 {
		return fmt.Errorf("no seeders for %s", manifest.Infohash)
	}

	t := c.getTransport()
	for i, pieceHash := range manifest.Pieces {
		piecePath := filepath.Join(PieceCachePath(c.dataDir), pieceHash)
		if _, err := os.Stat(piecePath); err == nil {
			continue
		}
		if t == nil {
			continue
		}
		for _, seeder := range seeders {
			data, err := t.RequestPiece(seeder, manifest.Infohash, i)
			if err != nil {
				continue
			}
			if manifest.VerifyPiece(i, data) {
				os.WriteFile(piecePath, data, 0644)
				t.AnnounceHave(pieceHash)
				break
			}
		}
	}
	return nil
}


// GetPiece handles the GetPiece HTTP request.
func (c *Cloud) GetPiece(infohash string, pieceIndex int) ([]byte, error) {
	c.mu.RLock()
	active, ok := c.seeding[infohash]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("not seeding %s", infohash)
	}
	if pieceIndex >= len(active.Manifest.Pieces) {
		return nil, fmt.Errorf("piece index %d out of range (max %d)", pieceIndex, len(active.Manifest.Pieces)-1)
	}
	pieceHash := active.Manifest.Pieces[pieceIndex]
	piecePath := filepath.Join(PieceCachePath(c.dataDir), pieceHash)
	data, err := os.ReadFile(piecePath)
	if err != nil {
		return nil, fmt.Errorf("read piece %d: %w", pieceIndex, err)
	}
	return data, nil
}
