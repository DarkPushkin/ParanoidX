// Package vault provides encryption and secure storage
package vault

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// QuotaMB is the free tier vault quota per device (16 GB).
const QuotaMB = 16384

// PeerCacheMB is additional space for P2P radio seeding (optional, user can grant more).
const PeerCacheMB = 16384

type FileInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	MTime string `json:"mtime"`
}

type Service struct {
	Path      string
	PeerCache string
}


// New handles the New HTTP request.
func New(dataDir string) *Service {
	vp := filepath.Join(dataDir, "vault")
	os.MkdirAll(vp, 0700)
	pc := filepath.Join(dataDir, "peer_cache")
	os.MkdirAll(pc, 0700)

	svc := &Service{Path: vp, PeerCache: pc}
	svc.ReserveQuota()
	return svc
}

// ReserveQuota ensures the vault has at least QuotaMB reserved as a sparse file.
// This guarantees baseline storage for the network without immediately using disk space.
func (s *Service) ReserveQuota() {
	reservePath := filepath.Join(s.Path, ".vault_reserve")
	stat, err := os.Stat(reservePath)
	if err == nil && stat.Size() >= int64(QuotaMB)*1024*1024 {
		return // already reserved
	}

	// Create a sparse file to reserve space
	f, err := os.Create(reservePath)
	if err != nil {
		return
	}
	defer f.Close()

	// Sparse file: allocate virtual space without writing blocks
	targetSize := int64(QuotaMB) * 1024 * 1024
	if err := syscall.Ftruncate(int(f.Fd()), targetSize); err != nil {
		// fallback: write a sparse marker
		f.Truncate(targetSize)
	}

	// Check actual disk usage; log warning if insufficient
	var diskAvail int64
	var statFs syscall.Statfs_t
	if err := syscall.Statfs(s.Path, &statFs); err == nil {
		diskAvail = int64(statFs.Bavail) * statFs.Bsize
		if diskAvail < targetSize {
			fmt.Printf("[vault] WARNING: only %d GB available, need %d GB for vault quota\n",
				diskAvail/(1024*1024*1024), QuotaMB/(1024))
		}
	}
}

// UsedMB returns total used MB in vault + peer_cache.
func (s *Service) UsedMB() float64 {
	return dirSizeMB(s.Path) + dirSizeMB(s.PeerCache)
}

// UsedPct returns percentage of quota used.
func (s *Service) UsedPct() float64 {
	return (s.UsedMB() / float64(QuotaMB)) * 100
}

func dirSizeMB(path string) float64 {
	var total int64
	filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		if filepath.Base(p) == ".vault_reserve" {
			return nil
		}
		total += fi.Size()
		return nil
	})
	return float64(total) / (1024 * 1024)
}


// SizeMB handles the SizeMB HTTP request.
func (s *Service) SizeMB() float64 {
	return dirSizeMB(s.Path)
}


// FileCount handles the FileCount HTTP request.
func (s *Service) FileCount() int {
	return countFiles(s.Path) + countFiles(s.PeerCache)
}

func countFiles(path string) int {
	count := 0
	ents, _ := os.ReadDir(path)
	for _, e := range ents {
		if e.IsDir() || e.Name() == ".vault_reserve" {
			continue
		}
		count++
	}
	return count
}


// List handles the List HTTP request.
func (s *Service) List() []FileInfo {
	return listFiles(s.Path)
}

func listFiles(dir string) []FileInfo {
	var files []FileInfo
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		if e.IsDir() || e.Name() == ".vault_reserve" {
			continue
		}
		if fi, err := e.Info(); err == nil {
			files = append(files, FileInfo{
				Name:  e.Name(),
				Size:  fi.Size(),
				MTime: fi.ModTime().Format(time.RFC3339),
			})
		}
	}
	return files
}


// Upload handles the Upload HTTP request.
func (s *Service) Upload(name string, reader io.Reader, size int64) (string, error) {
	used := s.SizeMB() * 1024 * 1024
	if used+float64(size) > float64(QuotaMB)*1024*1024 {
		return "", io.ErrUnexpectedEOF
	}
	name = filepath.Base(name)
	if name == "" || name == "." || name == ".." {
		name = "unnamed-" + time.Now().Format("20060102-150405")
	}
	dst, err := os.Create(filepath.Join(s.Path, name))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, reader); err != nil {
		return "", err
	}
	return name, nil
}


// Download handles the Download HTTP request.
func (s *Service) Download(name string) string {
	name = filepath.Base(name)
	return filepath.Join(s.Path, name)
}


// Delete handles the Delete HTTP request.
func (s *Service) Delete(name string) error {
	name = filepath.Base(name)
	return os.Remove(filepath.Join(s.Path, name))
}


// SaveNote handles the SaveNote HTTP request.
func (s *Service) SaveNote(name, content string) (string, error) {
	if name == "" {
		name = "note-" + time.Now().Format("20060102-150405") + ".txt"
	}
	name = filepath.Base(name)
	if err := os.WriteFile(filepath.Join(s.Path, name), []byte(content), 0600); err != nil {
		return "", err
	}
	return name, nil
}
