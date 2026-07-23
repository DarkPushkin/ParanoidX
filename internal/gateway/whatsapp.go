// Package gateway provides messaging gateway integrations
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type WhatsAppAdapter struct {
	Token        string
	PhoneID      string
	APIToken     string
	VerifyToken  string
	client       *http.Client
	webhookPath  string
}


// NewWhatsAppAdapter handles the NewWhatsAppAdapter HTTP request.
func NewWhatsAppAdapter(token, phoneID, apiToken string) *WhatsAppAdapter {
	return &WhatsAppAdapter{
		Token:    token,
		PhoneID:  phoneID,
		APIToken: apiToken,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}


// Name handles the Name HTTP request.
func (w *WhatsAppAdapter) Name() string { return "whatsapp" }


// Start handles the Start HTTP request.
func (w *WhatsAppAdapter) Start(ctx context.Context, router *Router) error {
	slog.Info("whatsapp adapter running (webhook-based — configure your WhatsApp Business webhook to POST to /api/webhook/whatsapp)")
	<-ctx.Done()
	return ctx.Err()
}


// Send handles the Send HTTP request.
func (w *WhatsAppAdapter) Send(chatID string, msg OutMessage) error {
	return w.sendText(chatID, msg.Text)
}


// SendText handles the SendText HTTP request.
func (w *WhatsAppAdapter) SendText(chatID, text string) error {
	return w.sendText(chatID, text)
}

func (w *WhatsAppAdapter) sendText(to, text string) error {
	if len(text) > 4096 {
		text = text[:4096]
	}
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "text",
		"text": map[string]string{
			"body": text,
		},
	}
	return w.apiPost(payload)
}


// SendButton handles the SendButton HTTP request.
func (w *WhatsAppAdapter) SendButton(to, text string, buttons []Button) error {
	if len(buttons) > 3 {
		buttons = buttons[:3]
	}
	rows := make([]map[string]string, len(buttons))
	for i, b := range buttons {
		rows[i] = map[string]string{"id": b.Data, "title": b.Text}
	}
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]any{
			"type":   "button",
			"body":   map[string]string{"text": text},
			"action": map[string]any{"buttons": rows},
		},
	}
	return w.apiPost(payload)
}

func (w *WhatsAppAdapter) apiPost(payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("whatsapp marshal: %w", err)
	}
	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", w.PhoneID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("whatsapp request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.APIToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp api %d: %s", resp.StatusCode, string(b))
	}
	return nil
}


// WebhookVerify handles the WebhookVerify HTTP request.
func (w *WhatsAppAdapter) WebhookVerify(r *http.Request) (string, bool) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")
	if mode == "subscribe" && token == w.VerifyToken {
		return challenge, true
	}
	return "", false
}


// WebhookHandler handles the WebhookHandler HTTP request.
func (w *WhatsAppAdapter) WebhookHandler(r *http.Request) ([]Message, error) {
	defer r.Body.Close()
	var payload struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Messages []struct {
						From string `json:"from"`
						ID   string `json:"id"`
						Text struct {
							Body string `json:"body"`
						} `json:"text"`
						Type string `json:"type"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("whatsapp webhook decode: %w", err)
	}
	var msgs []Message
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, m := range change.Value.Messages {
				msgs = append(msgs, Message{
					Platform:  "whatsapp",
					ChatID:    m.From,
					SenderID:  m.From,
					Text:      m.Text.Body,
					IsCommand: len(m.Text.Body) > 0 && m.Text.Body[0] == '/',
					Raw:       m,
				})
			}
		}
	}
	return msgs, nil
}
