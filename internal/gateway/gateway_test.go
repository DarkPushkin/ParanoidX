// Package gateway provides messaging gateway integrations
package gateway

import (
	"net/http/httptest"
	"strings"
	"testing"
)


// TestRouterBasic handles the TestRouterBasic HTTP request.
func TestRouterBasic(t *testing.T) {
	r := NewRouter()
	var called bool
	r.Handle("/test", func(msg Message) (*OutMessage, error) {
		called = true
		return &OutMessage{Text: "ok"}, nil
	})
	msg := Message{
		Platform:  "telegram",
		ChatID:    "123",
		Text:      "/test",
		IsCommand: true,
	}
	out, err := r.Route(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("handler not called")
	}
	if out.Text != "ok" {
		t.Fatalf("expected 'ok', got %q", out.Text)
	}
}


// TestRouterButton handles the TestRouterButton HTTP request.
func TestRouterButton(t *testing.T) {
	r := NewRouter()
	var called bool
	r.Handle("btn_help", func(msg Message) (*OutMessage, error) {
		called = true
		return &OutMessage{Text: "help"}, nil
	})
	msg := Message{
		Platform:   "telegram",
		ChatID:     "123",
		IsButton:   true,
		ButtonData: "btn_help",
	}
	out, err := r.Route(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("button handler not called")
	}
	if out.Text != "help" {
		t.Fatalf("expected 'help', got %q", out.Text)
	}
}


// TestRouterFallback handles the TestRouterFallback HTTP request.
func TestRouterFallback(t *testing.T) {
	r := NewRouter()
	r.Fallback(func(msg Message) (*OutMessage, error) {
		return &OutMessage{Text: "fallback"}, nil
	})
	msg := Message{
		Platform: "telegram",
		ChatID:   "123",
		Text:     "something unknown",
	}
	out, err := r.Route(msg)
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "fallback" {
		t.Fatalf("expected 'fallback', got %q", out.Text)
	}
}


// TestAdapterNames handles the TestAdapterNames HTTP request.
func TestAdapterNames(t *testing.T) {
	adapters := []struct {
		name     string
		adapter  Platform
		expected string
	}{
		{"telegram", NewTelegramAdapter("token", "testbot"), "testbot"},
		{"whatsapp", NewWhatsAppAdapter("", "", ""), "whatsapp"},
		{"signal", NewSignalAdapter("", ""), "signal"},
		{"matrix", NewMatrixAdapter("", "", ""), "matrix"},
		{"discord", NewDiscordAdapter("", "", ""), "discord"},
	}
	for _, a := range adapters {
		if a.adapter.Name() != a.expected {
			t.Errorf("%s: expected %q, got %q", a.name, a.expected, a.adapter.Name())
		}
	}
}


// TestWhatsAppWebhookVerify handles the TestWhatsAppWebhookVerify HTTP request.
func TestWhatsAppWebhookVerify(t *testing.T) {
	wa := NewWhatsAppAdapter("", "phoneid", "apitoken")
	wa.VerifyToken = "my_verify_token"

	r := httptest.NewRequest("GET", "/webhook?hub.mode=subscribe&hub.verify_token=my_verify_token&hub.challenge=challenge123", nil)
	challenge, ok := wa.WebhookVerify(r)
	if !ok {
		t.Fatal("expected verification to succeed")
	}
	if challenge != "challenge123" {
		t.Fatalf("expected 'challenge123', got %q", challenge)
	}
}


// TestWhatsAppWebhookVerifyFail handles the TestWhatsAppWebhookVerifyFail HTTP request.
func TestWhatsAppWebhookVerifyFail(t *testing.T) {
	wa := NewWhatsAppAdapter("", "phoneid", "apitoken")
	wa.VerifyToken = "my_verify_token"

	r := httptest.NewRequest("GET", "/webhook?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=challenge123", nil)
	_, ok := wa.WebhookVerify(r)
	if ok {
		t.Fatal("expected verification to fail")
	}
}


// TestWhatsAppWebhookHandler handles the TestWhatsAppWebhookHandler HTTP request.
func TestWhatsAppWebhookHandler(t *testing.T) {
	wa := NewWhatsAppAdapter("", "phoneid", "apitoken")
	body := `{
		"entry": [{
			"changes": [{
				"value": {
					"messages": [{
						"from": "15551234567",
						"id": "msg1",
						"text": {"body": "/help"},
						"type": "text"
					}]
				}
			}]
		}]
	}`
	r := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	msgs, err := wa.WebhookHandler(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Platform != "whatsapp" {
		t.Fatalf("expected platform 'whatsapp', got %q", msgs[0].Platform)
	}
	if msgs[0].ChatID != "15551234567" {
		t.Fatalf("expected chat '15551234567', got %q", msgs[0].ChatID)
	}
	if msgs[0].Text != "/help" {
		t.Fatalf("expected text '/help', got %q", msgs[0].Text)
	}
	if !msgs[0].IsCommand {
		t.Fatal("expected IsCommand true")
	}
}


// TestOutMessageButtons handles the TestOutMessageButtons HTTP request.
func TestOutMessageButtons(t *testing.T) {
	msg := OutMessage{
		Text: "Choose:",
		Buttons: [][]Button{
			{{Text: "Yes", Data: "yes"}, {Text: "No", Data: "no"}},
			{{Text: "Cancel", Data: "cancel"}},
		},
	}
	if msg.Text != "Choose:" {
		t.Fatalf("expected 'Choose:', got %q", msg.Text)
	}
	if len(msg.Buttons) != 2 {
		t.Fatalf("expected 2 button rows, got %d", len(msg.Buttons))
	}
	if len(msg.Buttons[0]) != 2 {
		t.Fatalf("expected 2 buttons in first row, got %d", len(msg.Buttons[0]))
	}
}


// TestRouteDetectsCommand handles the TestRouteDetectsCommand HTTP request.
func TestRouteDetectsCommand(t *testing.T) {
	r := NewRouter()
	r.Handle("/start", func(msg Message) (*OutMessage, error) {
		return &OutMessage{Text: "started"}, nil
	})
	msg := Message{Text: "/start", IsCommand: true}
	out, err := r.Route(msg)
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "started" {
		t.Fatalf("expected 'started', got %q", out.Text)
	}
}


// TestRouteIgnoresNonCommand handles the TestRouteIgnoresNonCommand HTTP request.
func TestRouteIgnoresNonCommand(t *testing.T) {
	r := NewRouter()
	var called bool
	r.Handle("/test", func(msg Message) (*OutMessage, error) {
		called = true
		return nil, nil
	})
	msg := Message{Text: "/test", IsCommand: false}
	_, _ = r.Route(msg)
	if called {
		t.Fatal("handler should not be called for non-command message")
	}
}


// TestSignalAdapterCLI handles the TestSignalAdapterCLI HTTP request.
func TestSignalAdapterCLI(t *testing.T) {
	s := NewSignalAdapter("/usr/bin/signal-cli", "+15551234567")
	if s.Name() != "signal" {
		t.Fatalf("expected 'signal', got %q", s.Name())
	}
	err := s.SendText("+15559876543", "test")
	if err != nil {
		// This will fail without signal-cli installed, but shouldn't panic
		t.Logf("expected error without signal-cli: %v", err)
	}
}


// TestMatrixAdapterSend handles the TestMatrixAdapterSend HTTP request.
func TestMatrixAdapterSend(t *testing.T) {
	m := NewMatrixAdapter("https://matrix.example.com", "@bot:example.com", "token")
	if m.Name() != "matrix" {
		t.Fatalf("expected 'matrix', got %q", m.Name())
	}
	err := m.SendText("!room:example.com", "test")
	if err != nil {
		t.Logf("expected error without matrix server: %v", err)
	}
}


// TestDiscordAdapterSend handles the TestDiscordAdapterSend HTTP request.
func TestDiscordAdapterSend(t *testing.T) {
	d := NewDiscordAdapter("token", "appid", "guildid")
	if d.Name() != "discord" {
		t.Fatalf("expected 'discord', got %q", d.Name())
	}
	err := d.SendText("123456", "test")
	if err != nil {
		t.Logf("expected error without discord token: %v", err)
	}
}
