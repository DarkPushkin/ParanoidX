// Package gateway provides messaging gateway integrations
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type DiscordAdapter struct {
	Token   string
	AppID   string
	GuildID string
	client  *http.Client
}


// NewDiscordAdapter handles the NewDiscordAdapter HTTP request.
func NewDiscordAdapter(token, appID, guildID string) *DiscordAdapter {
	return &DiscordAdapter{
		Token:   token,
		AppID:   appID,
		GuildID: guildID,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}


// Name handles the Name HTTP request.
func (d *DiscordAdapter) Name() string { return "discord" }


// Start handles the Start HTTP request.
func (d *DiscordAdapter) Start(ctx context.Context, router *Router) error {
	slog.Info("discord adapter running")
	<-ctx.Done()
	return ctx.Err()
}


// Send handles the Send HTTP request.
func (d *DiscordAdapter) Send(chatID string, msg OutMessage) error {
	return d.sendText(chatID, msg.Text)
}


// SendText handles the SendText HTTP request.
func (d *DiscordAdapter) SendText(chatID, text string) error {
	return d.sendText(chatID, text)
}

func (d *DiscordAdapter) sendText(channelID, text string) error {
	if len(text) > 2000 {
		text = text[:2000]
	}
	payload := map[string]any{
		"content": text,
	}
	if channelID == "" {
		return fmt.Errorf("discord: empty channel id")
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord request: %w", err)
	}
	req.Header.Set("Authorization", "Bot "+d.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("discord post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord api %d", resp.StatusCode)
	}
	return nil
}
