// Package gateway provides messaging gateway integrations
package gateway

import (
	"context"
	"strings"
)

type Message struct {
	Platform   string
	ChatID     string
	SenderID   string
	SenderName string
	Text       string
	IsCommand  bool
	IsButton   bool
	ButtonData string
	Raw        any
}

type Button struct {
	Text string
	Data string
	URL  string
}

type OutMessage struct {
	Text    string
	Buttons [][]Button
}

type Handler func(msg Message) (*OutMessage, error)

type Platform interface {
	Name() string
	Start(ctx context.Context, router *Router) error
	Send(chatID string, msg OutMessage) error
	SendText(chatID, text string) error
}

type Router struct {
	handlers  map[string]Handler
	fallbacks []Handler
}


// NewRouter handles the NewRouter HTTP request.
func NewRouter() *Router {
	return &Router{handlers: make(map[string]Handler)}
}


// Handle handles the Handle HTTP request.
func (r *Router) Handle(cmd string, h Handler) {
	r.handlers[strings.ToLower(cmd)] = h
}


// Fallback handles the Fallback HTTP request.
func (r *Router) Fallback(h Handler) {
	r.fallbacks = append(r.fallbacks, h)
}


// Route handles the Route HTTP request.
func (r *Router) Route(msg Message) (*OutMessage, error) {
	if msg.IsButton && msg.ButtonData != "" {
		if h, ok := r.handlers[strings.ToLower(msg.ButtonData)]; ok {
			return h(msg)
		}
	}
	if msg.IsCommand {
		parts := strings.Fields(msg.Text)
		cmd := strings.ToLower(parts[0])
		if h, ok := r.handlers[cmd]; ok {
			return h(msg)
		}
	}
	for _, fb := range r.fallbacks {
		out, err := fb(msg)
		if err != nil {
			continue
		}
		return out, nil
	}
	return &OutMessage{Text: "Unknown command. Try /help."}, nil
}
