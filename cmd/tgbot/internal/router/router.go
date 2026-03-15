// Package router wires all HTTP handlers into a single ServeMux.
package router

import (
	"net/http"
	"strings"

	"tgfinance/cmd/tgbot/internal/auth"
	"tgfinance/cmd/tgbot/internal/handler"
	"tgfinance/internal/finance"
	"tgfinance/internal/tasks"
)

// Config holds all dependencies needed by the bot.
type Config struct {
	BotToken    string
	AppURL      string
	UsageLog    string
	InternalKey string // for ZeroClaw server-to-server auth
	InternalUID int64  // owner's Telegram user ID
}

// New builds and returns the bot's HTTP handler.
func New(cfg Config, finStore *finance.Store, taskStore *tasks.Store, webappHTML []byte) http.Handler {
	a := auth.New(cfg.BotToken, cfg.InternalKey, cfg.InternalUID)

	webhook := handler.NewWebhook(cfg.BotToken, cfg.AppURL)
	stats := handler.NewStats(a, finStore)
	txns := handler.NewTransactions(a, finStore)
	imp := handler.NewImport(a, finStore, cfg.UsageLog)
	tsk := handler.NewTasks(a, taskStore)

	mux := http.NewServeMux()

	mux.HandleFunc("/webhook", webhook.Handle)

	mux.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(webappHTML)
	})

	mux.HandleFunc("/api/stats", cors(stats.Handle))

	mux.HandleFunc("/api/transactions", cors(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			txns.List(w, r)
		case http.MethodPost:
			txns.Create(w, r)
		case http.MethodDelete:
			txns.Clear(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/transactions/", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		txns.Delete(w, r)
	}))

	mux.HandleFunc("/api/import/claude", cors(imp.Claude))

	// Tasks
	mux.HandleFunc("/api/tasks/stats", cors(tsk.Stats))
	mux.HandleFunc("/api/tasks/", cors(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// /api/tasks/{id}/status
		if strings.HasSuffix(path, "/status") {
			if r.Method == http.MethodPatch {
				tsk.SetStatus(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
		// /api/tasks/{id}
		switch r.Method {
		case http.MethodPatch:
			tsk.Update(w, r)
		case http.MethodDelete:
			tsk.Delete(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/tasks", cors(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tsk.List(w, r)
		case http.MethodPost:
			tsk.Create(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	return mux
}

func cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Init-Data")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}
