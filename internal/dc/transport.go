// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

const maxMsgSize = 4 * ContainerPieceSize

type P2PTransport struct {
	peerAddr string
	nodeID   string
	timeout  time.Duration
}


// NewP2PTransport handles the NewP2PTransport HTTP request.
func NewP2PTransport(peerAddr, nodeID string) *P2PTransport {
	return &P2PTransport{
		peerAddr: peerAddr,
		nodeID:   nodeID,
		timeout:  10 * time.Second,
	}
}


// AnnounceHave handles the AnnounceHave HTTP request.
func (t *P2PTransport) AnnounceHave(hash string) {}


// AddPeer handles the AddPeer HTTP request.
func (t *P2PTransport) AddPeer(addr, id string) {}


// RequestPiece handles the RequestPiece HTTP request.
func (t *P2PTransport) RequestPiece(peerAddr, infohash string, pieceIndex int) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", peerAddr, t.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(t.timeout))

	payload, _ := json.Marshal(map[string]any{
		"infohash":    infohash,
		"piece_index": pieceIndex,
	})
	msg := map[string]any{
		"t": "dc_piece_req",
		"p": string(payload),
		"f": t.nodeID,
	}
	if err := t.sendJSON(conn, msg); err != nil {
		return nil, err
	}

	raw, err := t.recvRaw(conn)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Error string `json:"error,omitempty"`
		Data  string `json:"data,omitempty"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("peer: %s", resp.Error)
	}
	return base64.StdEncoding.DecodeString(resp.Data)
}


// RequestManifest handles the RequestManifest HTTP request.
func (t *P2PTransport) RequestManifest(peerAddr, infohash string) (*Manifest, error) {
	conn, err := net.DialTimeout("tcp", peerAddr, t.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(t.timeout))

	payload, _ := json.Marshal(map[string]string{"infohash": infohash})
	msg := map[string]any{
		"t": "dc_manifest_req",
		"p": string(payload),
		"f": t.nodeID,
	}
	if err := t.sendJSON(conn, msg); err != nil {
		return nil, err
	}

	raw, err := t.recvRaw(conn)
	if err != nil {
		return nil, err
	}
	var resp Manifest
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &resp, nil
}

func (t *P2PTransport) sendJSON(conn net.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

func (t *P2PTransport) recvRaw(conn net.Conn) ([]byte, error) {
	buf := make([]byte, maxMsgSize)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read: %w", err)
	}
	if n == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	if n >= len(buf) {
		return nil, fmt.Errorf("message too large")
	}
	out := make([]byte, n)
	copy(out, buf[:n])
	return out, nil
}


// HandleDCPieceReq handles the HandleDCPieceReq HTTP request.
func HandleDCPieceReq(cloud *Cloud) func(net.Conn, map[string]any) {
	return func(conn net.Conn, req map[string]any) {
		infohash, _ := req["infohash"].(string)
		pieceIndex, _ := req["piece_index"].(float64)
		data, err := cloud.GetPiece(infohash, int(pieceIndex))
		resp := map[string]any{}
		if err != nil {
			resp["error"] = err.Error()
		} else {
			resp["data"] = base64.StdEncoding.EncodeToString(data)
		}
		respData, err := json.Marshal(resp)
		if err != nil {
			return
		}
		conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
		conn.Write(respData)
	}
}


// HandleDCManifestReq handles the HandleDCManifestReq HTTP request.
func HandleDCManifestReq(cloud *Cloud) func(net.Conn, map[string]any) {
	return func(conn net.Conn, req map[string]any) {
		infohash, _ := req["infohash"].(string)
		manifestPath := ManifestPath(cloud.dataDir) + "/" + infohash + ".dc"
		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			resp := map[string]string{"error": err.Error()}
			respData, _ := json.Marshal(resp)
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			conn.Write(respData)
			return
		}
		respData, _ := json.Marshal(manifest)
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.Write(respData)
		if err != nil {
			slog.Warn("dc manifest write failed", "error", err)
		}
	}
}

func init() {
	slog.Info("dc transport handlers registered")
}
