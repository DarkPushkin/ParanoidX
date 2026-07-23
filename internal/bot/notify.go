// Package bot — notification helpers for bot fleet.
package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type NotifySender struct {
	Token  string
	ChatID int64
	Client *http.Client
}


// NewNotifySender handles the NewNotifySender HTTP request.
func NewNotifySender(token string, chatID int64) *NotifySender {
	return &NotifySender{
		Token:  token,
		ChatID: chatID,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}


// Send handles the Send HTTP request.
func (n *NotifySender) Send(text string) error {
	if n.Token == "" || n.ChatID == 0 {
		slog.Debug("notify: no token or chat configured")
		return nil
	}
	if len(text) > 4000 {
		text = text[:4000]
	}
	payload := map[string]any{
		"chat_id": n.ChatID,
		"text":    text,
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.Token)
	resp, err := n.Client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify send: %w", err)
	}
	defer resp.Body.Close()
	return nil
}


// IsConflictError handles the IsConflictError HTTP request.
func IsConflictError(err error) bool {
	return strings.Contains(err.Error(), `"error_code":409`)
}


// HandlePollError handles the HandlePollError HTTP request.
func HandlePollError(logger *slog.Logger, name string, err error) {
	if IsConflictError(err) {
		logger.Warn(name+" got 409 conflict — old instance still polling, waiting 15s", "error", err)
		time.Sleep(15 * time.Second)
	} else {
		logger.Error(name+" getUpdates", "error", err)
		time.Sleep(5 * time.Second)
	}
}
