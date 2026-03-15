package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Webhook handles incoming Telegram bot updates.
type Webhook struct {
	token  string
	appURL string
}

func NewWebhook(token, appURL string) *Webhook {
	return &Webhook{token: token, appURL: appURL}
}

type tgUpdate struct {
	Message *tgMessage `json:"message"`
}
type tgMessage struct {
	Chat tgChat `json:"chat"`
	Text string `json:"text"`
}
type tgChat struct {
	ID int64 `json:"id"`
}

func (h *Webhook) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	var upd tgUpdate
	if err := json.Unmarshal(body, &upd); err != nil || upd.Message == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	if upd.Message.Text == "/start" || upd.Message.Text == "/app" {
		h.sendApp(upd.Message.Chat.ID)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Webhook) sendApp(chatID int64) {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    "💰 Финансовый трекер — отслеживай доходы и расходы прямо в Telegram:",
		"reply_markup": map[string]any{
			"inline_keyboard": [][]map[string]any{{{
				"text":    "📊 Открыть трекер",
				"web_app": map[string]string{"url": h.appURL},
			}}},
		},
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(
		"https://api.telegram.org/bot"+h.token+"/sendMessage",
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		log.Printf("sendApp: %v", err)
		return
	}
	resp.Body.Close()
}
