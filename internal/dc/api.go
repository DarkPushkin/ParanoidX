// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type DCHandlers struct {
	cloud *Cloud
}


// NewDCHandlers handles the NewDCHandlers HTTP request.
func NewDCHandlers(cloud *Cloud) *DCHandlers {
	return &DCHandlers{cloud: cloud}
}


// SeedHandler handles the SeedHandler HTTP request.
func (h *DCHandlers) SeedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ContainerPath string `json:"container_path"`
			ContainerID   string `json:"container_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"error": "invalid JSON"})
			return
		}
		if req.ContainerPath == "" || req.ContainerID == "" {
			writeJSON(w, map[string]any{"error": "container_path and container_id required"})
			return
		}
		if _, err := os.Stat(req.ContainerPath); os.IsNotExist(err) {
			writeJSON(w, map[string]any{"error": "container file not found"})
			return
		}
		manifest, err := h.cloud.SeedContainer(req.ContainerPath, req.ContainerID)
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, map[string]any{
			"ok":         true,
			"infohash":   manifest.Infohash,
			"container":  manifest.ContainerID,
			"size_mb":    manifest.Size / (1024 * 1024),
			"pieces":     manifest.PieceCount,
		})
	}
}


// AnnonceHandler handles the AnnonceHandler HTTP request.
func (h *DCHandlers) AnnonceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Infohash  string `json:"infohash"`
			PeerAddr  string `json:"peer_addr"`
			PeerID    string `json:"peer_id"`
			IsSeeder  bool   `json:"is_seeder"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"error": "invalid JSON"})
			return
		}
		if req.Infohash == "" || req.PeerAddr == "" {
			writeJSON(w, map[string]any{"error": "infohash and peer_addr required"})
			return
		}
		if req.IsSeeder {
			h.cloud.AddSeeder(req.Infohash, req.PeerAddr)
		} else {
			h.cloud.AddLeecher(req.Infohash, req.PeerAddr)
		}
		meta := h.cloud.GetContainer(req.Infohash)
		if meta == nil {
			writeJSON(w, map[string]any{"ok": false, "error": "container not found"})
			return
		}
		writeJSON(w, map[string]any{
			"ok":       true,
			"infohash": req.Infohash,
			"replicas": meta.Replicas,
			"seeders":  meta.Seeders,
		})
	}
}


// SwarmHandler handles the SwarmHandler HTTP request.
func (h *DCHandlers) SwarmHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, `{"error":"GET required"}`, http.StatusMethodNotAllowed)
			return
		}
		infohash := r.URL.Query().Get("infohash")
		if infohash == "" {
			writeJSON(w, map[string]any{
				"containers": h.cloud.ListContainers(),
				"count":      len(h.cloud.ListContainers()),
			})
			return
		}
		meta := h.cloud.GetContainer(infohash)
		if meta == nil {
			writeJSON(w, map[string]any{"error": "container not found"})
			return
		}
		writeJSON(w, map[string]any{
			"infohash":  meta.Infohash,
			"size_mb":   meta.Size / (1024 * 1024),
			"pieces":    meta.PieceCount,
			"seeders":   meta.Seeders,
			"leechers":  meta.Leechers,
			"replicas":  meta.Replicas,
			"target_rep": meta.TargetRep,
		})
	}
}


// FetchHandler handles the FetchHandler HTTP request.
func (h *DCHandlers) FetchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Infohash    string `json:"infohash"`
			OutputPath  string `json:"output_path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"error": "invalid JSON"})
			return
		}
		if req.Infohash == "" {
			writeJSON(w, map[string]any{"error": "infohash required"})
			return
		}
		if req.OutputPath == "" {
			req.OutputPath = filepath.Join(h.cloud.dataDir, ContainerDir, req.Infohash+".bin")
		}
		if err := h.cloud.FetchContainer(req.Infohash, req.OutputPath); err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, map[string]any{
			"ok":          true,
			"infohash":    req.Infohash,
			"output_path": req.OutputPath,
		})
	}
}


// ListHandler handles the ListHandler HTTP request.
func (h *DCHandlers) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, `{"error":"GET required"}`, http.StatusMethodNotAllowed)
			return
		}
		containers := h.cloud.ListContainers()
		type brief struct {
			Infohash  string   `json:"infohash"`
			SizeMB    int64    `json:"size_mb"`
			Pieces    int      `json:"pieces"`
			Seeders   []string `json:"seeders"`
			Replicas  int      `json:"replicas"`
		}
		list := make([]brief, 0, len(containers))
		for _, c := range containers {
			list = append(list, brief{
				Infohash: c.Infohash,
				SizeMB:   c.Size / (1024 * 1024),
				Pieces:   c.PieceCount,
				Seeders:  c.Seeders,
				Replicas: c.Replicas,
			})
		}
		writeJSON(w, map[string]any{
			"containers": list,
			"count":      len(list),
		})
	}
}


// StatusHandler handles the StatusHandler HTTP request.
func (h *DCHandlers) StatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, `{"error":"GET required"}`, http.StatusMethodNotAllowed)
			return
		}
		seeding := h.cloud.SeedingStatus()
		containers := h.cloud.ListContainers()
		totalSize := int64(0)
		for _, c := range containers {
			totalSize += c.Size
		}
		writeJSON(w, map[string]any{
			"seeding":     len(seeding),
			"containers":  len(containers),
			"total_mb":    totalSize / (1024 * 1024),
			"data_dir":    h.cloud.dataDir,
		})
	}
}


// ManifestHandler handles the ManifestHandler HTTP request.
func (h *DCHandlers) ManifestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, `{"error":"GET required"}`, http.StatusMethodNotAllowed)
			return
		}
		infohash := r.URL.Query().Get("infohash")
		if infohash == "" {
			writeJSON(w, map[string]any{"error": "infohash required"})
			return
		}
		manifestPath := filepath.Join(ManifestPath(h.cloud.dataDir), infohash+".dc")
		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			writeJSON(w, map[string]any{"error": "manifest not found"})
			return
		}
		writeJSON(w, manifest)
	}
}


// PieceHandler handles the PieceHandler HTTP request.
func (h *DCHandlers) PieceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, `{"error":"GET required"}`, http.StatusMethodNotAllowed)
			return
		}
		infohash := r.URL.Query().Get("infohash")
		pieceStr := r.URL.Query().Get("piece")
		if infohash == "" || pieceStr == "" {
			writeJSON(w, map[string]any{"error": "infohash and piece required"})
			return
		}
		pieceIndex := 0
		if _, err := fmt.Sscanf(pieceStr, "%d", &pieceIndex); err != nil {
			writeJSON(w, map[string]any{"error": "invalid piece index"})
			return
		}
		data, err := h.cloud.GetPiece(infohash, pieceIndex)
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Piece-Index", pieceStr)
		w.Header().Set("X-Infohash", infohash)
		w.Write(data)
	}
}


// HealReportHandler handles the HealReportHandler HTTP request.
func (h *DCHandlers) HealReportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, `{"error":"GET required"}`, http.StatusMethodNotAllowed)
			return
		}
		stats := h.cloud.swarm.GetHealStats()
		writeJSON(w, map[string]any{
			"total_repairs":      stats.TotalRepairs,
			"hash_mismatches":    stats.HashMismatches,
			"last_repair_time":   stats.LastRepairTime,
			"last_mismatch_info": stats.LastMismatchInfo,
			"by_infohash":        stats.ByInfohash,
		})
	}
}


// UnseedHandler handles the UnseedHandler HTTP request.
func (h *DCHandlers) UnseedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Infohash string `json:"infohash"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"error": "invalid JSON"})
			return
		}
		if err := h.cloud.StopSeeding(req.Infohash); err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "infohash": req.Infohash})
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("dc writeJSON failed", "error", err)
	}
}
