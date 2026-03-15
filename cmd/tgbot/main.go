// tgbot — Telegram Mini App backend for income/expense tracking.
//
// Environment variables:
//
//	TG_BOT_TOKEN  — Telegram bot token (required)
//	TG_APP_URL    — Public URL of the Mini App (default: https://zhandos.top/app)
//	FINANCE_DB    — Path to SQLite database (default: finance.db)
//	BOT_PORT      — Port to listen on (default: 4002)
//	USAGE_LOG     — Path to claude-usage.jsonl (default: claude-usage.jsonl)
package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"

	"strconv"

	"tgfinance/cmd/tgbot/internal/router"
	"tgfinance/internal/finance"
	"tgfinance/internal/tasks"
)

//go:embed webapp/index.html
var webappHTML []byte

func main() {
	botToken := os.Getenv("TG_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TG_BOT_TOKEN not set")
	}

	dbPath := envOr("FINANCE_DB", "finance.db")

	finStore, err := finance.New(dbPath)
	if err != nil {
		log.Fatalf("finance store: %v", err)
	}
	taskStore, err := tasks.New(dbPath)
	if err != nil {
		log.Fatalf("tasks store: %v", err)
	}

	internalUID, _ := strconv.ParseInt(envOr("INTERNAL_USER_ID", "0"), 10, 64)
	cfg := router.Config{
		BotToken:    botToken,
		AppURL:      envOr("TG_APP_URL", "https://zhandos.top/app"),
		UsageLog:    envOr("USAGE_LOG", "claude-usage.jsonl"),
		InternalKey: envOr("INTERNAL_API_KEY", ""),
		InternalUID: internalUID,
	}

	addr := ":" + envOr("BOT_PORT", "4002")
	log.Printf("tgbot listening on %s  app=%s", addr, cfg.AppURL)
	log.Fatal(http.ListenAndServe(addr, router.New(cfg, finStore, taskStore, webappHTML)))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
