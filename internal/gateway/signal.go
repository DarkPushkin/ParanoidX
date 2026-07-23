// Package gateway provides messaging gateway integrations
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
)

type SignalAdapter struct {
	CLIPath      string
	Number       string
	daemonMode   bool
	httpEndpoint string
}


// NewSignalAdapter handles the NewSignalAdapter HTTP request.
func NewSignalAdapter(cliPath, number string) *SignalAdapter {
	return &SignalAdapter{CLIPath: cliPath, Number: number}
}


// NewSignalDaemonAdapter handles the NewSignalDaemonAdapter HTTP request.
func NewSignalDaemonAdapter(endpoint, number string) *SignalAdapter {
	return &SignalAdapter{
		httpEndpoint: endpoint,
		Number:       number,
		daemonMode:   true,
	}
}


// Name handles the Name HTTP request.
func (s *SignalAdapter) Name() string { return "signal" }


// Start handles the Start HTTP request.
func (s *SignalAdapter) Start(ctx context.Context, router *Router) error {
	slog.Info("signal adapter running")
	<-ctx.Done()
	return ctx.Err()
}


// Send handles the Send HTTP request.
func (s *SignalAdapter) Send(chatID string, msg OutMessage) error {
	return s.sendText(chatID, msg.Text)
}


// SendText handles the SendText HTTP request.
func (s *SignalAdapter) SendText(chatID, text string) error {
	return s.sendText(chatID, text)
}

func (s *SignalAdapter) sendText(to, text string) error {
	if s.daemonMode && s.httpEndpoint != "" {
		return s.sendDaemon(to, text)
	}
	return s.sendCLI(to, text)
}

func (s *SignalAdapter) sendCLI(to, text string) error {
	args := []string{"send", "-u", s.Number, to, "-m", text}
	cmd := exec.Command(s.CLIPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("signal-cli send: %w: %s", err, string(out))
	}
	return nil
}

func (s *SignalAdapter) sendDaemon(to, text string) error {
	payload := map[string]any{
		"recipient": []string{to},
		"message":   text,
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/v2/send", strings.TrimRight(s.httpEndpoint, "/"))
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("signal daemon send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("signal daemon api %d", resp.StatusCode)
	}
	return nil
}
