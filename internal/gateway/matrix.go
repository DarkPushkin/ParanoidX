// Package gateway provides messaging gateway integrations
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type MatrixAdapter struct {
	Homeserver string
	UserID     string
	Token      string
	client     *http.Client
}


// NewMatrixAdapter handles the NewMatrixAdapter HTTP request.
func NewMatrixAdapter(homeserver, userID, token string) *MatrixAdapter {
	return &MatrixAdapter{
		Homeserver: strings.TrimRight(homeserver, "/"),
		UserID:     userID,
		Token:      token,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}


// Name handles the Name HTTP request.
func (m *MatrixAdapter) Name() string { return "matrix" }


// Start handles the Start HTTP request.
func (m *MatrixAdapter) Start(ctx context.Context, router *Router) error {
	slog.Info("matrix adapter running")
	<-ctx.Done()
	return ctx.Err()
}


// Send handles the Send HTTP request.
func (m *MatrixAdapter) Send(chatID string, msg OutMessage) error {
	return m.sendText(chatID, msg.Text)
}


// SendText handles the SendText HTTP request.
func (m *MatrixAdapter) SendText(chatID, text string) error {
	return m.sendText(chatID, text)
}

func (m *MatrixAdapter) sendText(roomID, text string) error {
	content := map[string]string{"msgtype": "m.text", "body": text}
	payload := map[string]any{"msgtype": "m.text", "body": text, "content": content}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message", m.Homeserver, roomID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("matrix request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("matrix post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("matrix api %d", resp.StatusCode)
	}
	return nil
}
