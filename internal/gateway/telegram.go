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
	"strconv"
	"strings"
	"time"
)

type TelegramAdapter struct {
	Token    string
	botName  string
	BaseURL  string
	Client   *http.Client
	lastID   int64
}

type tgUpdate struct {
	UpdateID      int64          `json:"update_id"`
	Message       *tgMessage     `json:"message,omitempty"`
	CallbackQuery *tgCallbackQ   `json:"callback_query,omitempty"`
}

type tgMessage struct {
	MessageID int64   `json:"message_id"`
	Chat      tgChat  `json:"chat"`
	From      tgUser  `json:"from"`
	Text      string  `json:"text"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

type tgUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
	IsBot     bool   `json:"is_bot"`
}

type tgCallbackQ struct {
	ID      string     `json:"id"`
	From    tgUser     `json:"from"`
	Message *tgMessage `json:"message,omitempty"`
	Data    string     `json:"data"`
}

type tgUpdateResp struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

type tgSendResp struct {
	OK bool `json:"ok"`
}


// NewTelegramAdapter handles the NewTelegramAdapter HTTP request.
func NewTelegramAdapter(token, name string) *TelegramAdapter {
	return &TelegramAdapter{
		Token:   token,
		botName: name,
		BaseURL: "https://api.telegram.org/bot" + token,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}


// Name handles the Name HTTP request.
func (t *TelegramAdapter) Name() string { return t.botName }


// Start handles the Start HTTP request.
func (t *TelegramAdapter) Start(ctx context.Context, router *Router) error {
	slog.Info("telegram adapter starting", "bot", t.Name)
	for {
		select {
		case <-ctx.Done():
			slog.Info("telegram adapter stopping", "bot", t.Name)
			return ctx.Err()
		default:
		}
		updates, err := t.getUpdates(ctx)
		if err != nil {
			if strings.Contains(err.Error(), `"error_code":409`) {
				slog.Warn("telegram 409 conflict", "bot", t.Name)
				time.Sleep(15 * time.Second)
			} else {
				slog.Error("telegram getUpdates", "bot", t.Name, "error", err)
				time.Sleep(5 * time.Second)
			}
			continue
		}
		for _, upd := range updates {
			if upd.UpdateID >= t.lastID {
				t.lastID = upd.UpdateID + 1
			}
			msg := t.toMessage(upd)
			if msg == nil {
				continue
			}
			out, err := router.Route(*msg)
			if err != nil {
				slog.Error("telegram route", "bot", t.Name, "error", err)
				continue
			}
			if out != nil {
				t.sendMsg(upd, *out)
			}
		}
		time.Sleep(2 * time.Second)
	}
}


// Send handles the Send HTTP request.
func (t *TelegramAdapter) Send(chatID string, msg OutMessage) error {
	return t.sendMessage(chatID, msg.Text, toTGKeyboard(msg.Buttons))
}


// SendText handles the SendText HTTP request.
func (t *TelegramAdapter) SendText(chatID, text string) error {
	return t.sendMessage(chatID, text, nil)
}

func (t *TelegramAdapter) getUpdates(ctx context.Context) ([]tgUpdate, error) {
	url := fmt.Sprintf("%s/getUpdates?timeout=10&allowed_updates=[\"message\",\"callback_query\"]", t.BaseURL)
	if t.lastID > 0 {
		url += "&offset=" + strconv.FormatInt(t.lastID, 10)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var r tgUpdateResp
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("json: %w: %s", err, string(body))
	}
	if !r.OK {
		return nil, fmt.Errorf("telegram api: %s", string(body))
	}
	return r.Result, nil
}

func (t *TelegramAdapter) sendMessage(chatID, text string, kb *tgKeyboard) error {
	if len(text) > 4000 {
		text = text[:4000]
	}
	payload := map[string]any{"chat_id": chatID, "text": text}
	if kb != nil {
		payload["reply_markup"] = kb
	}
	body, _ := json.Marshal(payload)
	url := t.BaseURL + "/sendMessage"
	resp, err := t.Client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (t *TelegramAdapter) toMessage(upd tgUpdate) *Message {
	if upd.CallbackQuery != nil {
		cq := upd.CallbackQuery
		chatID := int64(0)
		if cq.Message != nil {
			chatID = cq.Message.Chat.ID
		}
		return &Message{
			Platform:   t.Name(),
			ChatID:     strconv.FormatInt(chatID, 10),
			SenderID:   strconv.FormatInt(cq.From.ID, 10),
			SenderName: cq.From.Username,
			Text:       cq.Data,
			IsCommand:  false,
			IsButton:   true,
			ButtonData: cq.Data,
			Raw:        cq,
		}
	}
	if upd.Message != nil && !upd.Message.From.IsBot && upd.Message.Text != "" {
		m := upd.Message
		isCmd := strings.HasPrefix(m.Text, "/")
		return &Message{
			Platform:   t.Name(),
			ChatID:     strconv.FormatInt(m.Chat.ID, 10),
			SenderID:   strconv.FormatInt(m.From.ID, 10),
			SenderName: m.From.Username,
			Text:       m.Text,
			IsCommand:  isCmd,
			IsButton:   false,
			Raw:        m,
		}
	}
	return nil
}

func (t *TelegramAdapter) sendMsg(upd tgUpdate, msg OutMessage) {
	chatID := ""
	if upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil {
		chatID = strconv.FormatInt(upd.CallbackQuery.Message.Chat.ID, 10)
		t.answerCb(upd.CallbackQuery.ID, "")
	} else if upd.Message != nil {
		chatID = strconv.FormatInt(upd.Message.Chat.ID, 10)
	}
	if chatID == "" {
		return
	}
	kb := toTGKeyboard(msg.Buttons)
	t.sendMessage(chatID, msg.Text, kb)
}

func (t *TelegramAdapter) answerCb(cbID, text string) {
	payload := map[string]any{"callback_query_id": cbID}
	if text != "" {
		payload["text"] = text
	}
	body, _ := json.Marshal(payload)
	url := t.BaseURL + "/answerCallbackQuery"
	resp, err := t.Client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Error("answerCallbackQuery", "error", err)
		return
	}
	resp.Body.Close()
}

type tgKeyboard struct {
	InlineKeyboard [][]tgButton `json:"inline_keyboard"`
}

type tgButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

func toTGKeyboard(buttons [][]Button) *tgKeyboard {
	if len(buttons) == 0 {
		return nil
	}
	kb := &tgKeyboard{}
	for _, row := range buttons {
		var tgRow []tgButton
		for _, btn := range row {
			tgRow = append(tgRow, tgButton{
				Text:         btn.Text,
				CallbackData: btn.Data,
				URL:          btn.URL,
			})
		}
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgRow)
	}
	return kb
}
