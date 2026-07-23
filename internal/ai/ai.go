// Package ai provides an Ollama client and AI Steward for the simplex-node economy.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  Options   `json:"options,omitempty"`
}

type Options struct {
	NumPredict  int     `json:"num_predict,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Model      string  `json:"model"`
	Message    Message `json:"message"`
	Done       bool    `json:"done"`
	DoneReason string  `json:"done_reason,omitempty"`
}

type GenerateRequest struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	System  string  `json:"system,omitempty"`
	Options Options `json:"options,omitempty"`
}

type GenerateResponse struct {
	Model      string `json:"model"`
	Response   string `json:"response"`
	Done       bool   `json:"done"`
	DoneReason string `json:"done_reason,omitempty"`
	Thinking   string `json:"thinking,omitempty"`
}

type Client struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}


// NewClient handles the NewClient HTTP request.
func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "minimax-m3:cloud"
	}
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{Timeout: 1800 * time.Second},
	}
}


// Chat handles the Chat HTTP request.
func (c *Client) Chat(messages []Message, opts Options) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
		Options:  opts,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ollama marshal: %w", err)
	}
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama chat: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama chat read body: %w", err)
	}
	var chatResp ChatResponse
	if err := json.Unmarshal(b, &chatResp); err != nil {
		return nil, fmt.Errorf("ollama decode: %w: %s", err, string(b))
	}
	return &chatResp, nil
}


// Generate handles the Generate HTTP request.
func (c *Client) Generate(prompt string, system string, opts Options) (*GenerateResponse, error) {
	req := GenerateRequest{
		Model:   c.Model,
		Prompt:  prompt,
		System:  system,
		Stream:  false,
		Options: opts,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ollama marshal: %w", err)
	}
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama generate: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama generate read body: %w", err)
	}
	var genResp GenerateResponse
	if err := json.Unmarshal(b, &genResp); err != nil {
		return nil, fmt.Errorf("ollama decode: %w: %s", err, string(b))
	}
	return &genResp, nil
}


// ChatStream handles the ChatStream HTTP request.
func (c *Client) ChatStream(messages []Message, opts Options) (<-chan string, error) {
	req := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
		Options:  opts,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ollama marshal: %w", err)
	}
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama chat stream: %w", err)
	}

	tokenCh := make(chan string, 10)
	go func() {
		defer resp.Body.Close()
		defer close(tokenCh)
		dec := json.NewDecoder(resp.Body)
		for {
			var chatResp ChatResponse
			if err := dec.Decode(&chatResp); err != nil {
				if err == io.EOF {
					return
				}
				return
			}
			if chatResp.Message.Content != "" {
				tokenCh <- chatResp.Message.Content
			}
			if chatResp.Done {
				return
			}
		}
	}()
	return tokenCh, nil
}


// GenerateStream handles the GenerateStream HTTP request.
func (c *Client) GenerateStream(prompt string, system string, opts Options) (<-chan string, error) {
	req := GenerateRequest{
		Model:   c.Model,
		Prompt:  prompt,
		System:  system,
		Stream:  true,
		Options: opts,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ollama marshal: %w", err)
	}
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama generate stream: %w", err)
	}

	tokenCh := make(chan string, 10)
	go func() {
		defer resp.Body.Close()
		defer close(tokenCh)
		dec := json.NewDecoder(resp.Body)
		for {
			var genResp GenerateResponse
			if err := dec.Decode(&genResp); err != nil {
				if err == io.EOF {
					return
				}
				return
			}
			if genResp.Response != "" {
				tokenCh <- genResp.Response
			}
			if genResp.Done {
				return
			}
		}
	}()
	return tokenCh, nil
}


// IsAvailable handles the IsAvailable HTTP request.
func (c *Client) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
