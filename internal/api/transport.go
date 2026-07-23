// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var startTime = time.Now()

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TransportEnvelope wraps API requests/responses for relay transport.
type TransportEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	ID      string          `json:"id,omitempty"`
}

type apiRequest struct {
	Method string          `json:"method"`
	Path   string          `json:"path"`
	Body   json.RawMessage `json:"body,omitempty"`
}

type apiResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}


// TransportHandler serves transport info and health status endpoints.
func TransportHandler(dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// GET /api/transport/info or /api/transport/health
		if strings.HasSuffix(r.URL.Path, "/health") || strings.HasSuffix(r.URL.Path, "/status") {
			status := map[string]interface{}{
				"tor":   fileExists(filepath.Join(dataDir, "dashboard_onion.txt")),
				"smp":   fileExists(filepath.Join(dataDir, "smp_client_address.txt")),
				"xftp":  fileExists(filepath.Join(dataDir, "xftp_client_address.txt")),
				"radio": true,
				"uptime": time.Since(startTime).String(),
			}
			writeJSON(w, status)
			return
		}
		info := map[string]string{
			"smp":     readTrim(filepath.Join(dataDir, "smp_client_address.txt")),
			"xftp":    readTrim(filepath.Join(dataDir, "xftp_client_address.txt")),
			"ice":     readTrim(filepath.Join(dataDir, "ice_onion.txt")),
			"onion":   readTrim(filepath.Join(dataDir, "dashboard_onion.txt")),
			"contact": readTrim(filepath.Join(dataDir, "island_contact_link.txt")),
			"label":   "Saint Mary Liberty Island Node",
			"version": "simplex-transport-v1",
		}
		writeJSON(w, info)
	}
}


// TransportSendHandler forwards relayed API requests to the local server.
func TransportSendHandler(dataDir string) http.HandlerFunc {
	inner := &http.Client{Timeout: 30 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		var env TransportEnvelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			http.Error(w, `{"type":"error","payload":"bad json"}`, 400)
			return
		}
		r.Body.Close()

		switch env.Type {
		case "api.request":
			var req apiRequest
			if err := json.Unmarshal(env.Payload, &req); err != nil {
				writeJSON(w, TransportEnvelope{
					Type:    "api.response",
					Payload: json.RawMessage(`{"status":400,"body":"bad request format"}`),
					ID:      env.ID,
				})
				return
			}
			body, status := proxyAPICall(inner, req)
			respPayload, _ := json.Marshal(apiResponse{Status: status, Body: body})
			writeJSON(w, TransportEnvelope{
				Type:    "api.response",
				Payload: respPayload,
				ID:      env.ID,
			})
		case "ping":
			writeJSON(w, TransportEnvelope{
				Type:    "pong",
				Payload: json.RawMessage(`{"status":"ok","time":"` + time.Now().Format(time.RFC3339) + `"}`),
				ID:      env.ID,
			})
		default:
			writeJSON(w, TransportEnvelope{
				Type:    "error",
				Payload: json.RawMessage(`{"error":"unknown type"}`),
				ID:      env.ID,
			})
		}
	}
}

func proxyAPICall(client *http.Client, req apiRequest) (json.RawMessage, int) {
	u := fmt.Sprintf("http://127.0.0.1:8080%s", req.Path)
	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "GET"
	}
	var bodyReader io.Reader
	if len(req.Body) > 0 && string(req.Body) != "null" {
		bodyReader = bytes.NewReader(req.Body)
	}
	httpReq, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		errBody, _ := json.Marshal(map[string]string{"error": err.Error()})
		return errBody, 500
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		errBody, _ := json.Marshal(map[string]string{"error": err.Error()})
		return errBody, 502
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errBody, _ := json.Marshal(map[string]string{"error": err.Error()})
		return errBody, 502
	}
	return respBody, resp.StatusCode
}
