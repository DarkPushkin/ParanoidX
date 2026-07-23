// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)


// SeedContainer handles the SeedContainer HTTP request.
func (c *Cloud) SeedContainer(containerPath, containerID string) (*Manifest, error) {
	manifest, err := BuildManifest(containerID, containerPath)
	if err != nil {
		return nil, fmt.Errorf("build manifest: %w", err)
	}

	if err := manifest.Save(ManifestPath(c.dataDir)); err != nil {
		return nil, fmt.Errorf("save manifest: %w", err)
	}

	dstPath := filepath.Join(ContainerPath(c.dataDir), manifest.Infohash)
	if err := copyFile(containerPath, dstPath); err != nil {
		return nil, fmt.Errorf("copy container: %w", err)
	}

	if err := cacheAllPieces(c.dataDir, dstPath, manifest); err != nil {
		return nil, fmt.Errorf("cache pieces: %w", err)
	}

	active := &ActiveSeed{
		Infohash:   manifest.Infohash,
		Manifest:   manifest,
		PieceCount: len(manifest.Pieces),
		Started:    time.Now(),
	}

	c.mu.Lock()
	c.seeding[manifest.Infohash] = active
	if _, ok := c.containers[manifest.Infohash]; !ok {
		c.registerContainerLocked(manifest.Infohash, manifest.Size, manifest.PieceCount)
	}
	c.mu.Unlock()

	if t := c.getTransport(); t != nil {
		t.AnnounceHave(manifest.Infohash)
	}

	slog.Info("dc seeding container",
		"infohash", manifest.Infohash,
		"container_id", containerID,
		"size_mb", manifest.Size/(1024*1024),
		"pieces", manifest.PieceCount,
	)
	return manifest, nil
}


// StopSeeding handles the StopSeeding HTTP request.
func (c *Cloud) StopSeeding(infohash string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.seeding, infohash)
	return nil
}

func cacheAllPieces(dataDir, containerPath string, m *Manifest) error {
	f, err := os.Open(containerPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, m.PieceSize)
	for i := 0; i < len(m.Pieces); i++ {
		n, err := f.Read(buf)
		if n > 0 {
			piecePath := filepath.Join(PieceCachePath(dataDir), m.Pieces[i])
			if err := os.WriteFile(piecePath, buf[:n], 0644); err != nil {
				return fmt.Errorf("cache piece %d: %w", i, err)
			}
		}
		if err != nil {
			break
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}


// ManifestPath handles the ManifestPath HTTP request.
func ManifestPath(dataDir string) string {
	return filepath.Join(dataDir, ManifestDir)
}


// ContainerPath handles the ContainerPath HTTP request.
func ContainerPath(dataDir string) string {
	return filepath.Join(dataDir, ContainerDir)
}


// PieceCachePath handles the PieceCachePath HTTP request.
func PieceCachePath(dataDir string) string {
	return filepath.Join(dataDir, PieceCacheDir)
}
